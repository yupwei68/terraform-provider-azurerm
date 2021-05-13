package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/arm/storage/2019-06-01/armstorage"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/parse"
)

var (
	storageAccountsCache = map[string]accountDetails{}

	accountsLock    = sync.RWMutex{}
	credentialsLock = sync.RWMutex{}
)

type accountDetails struct {
	ID            string
	ResourceGroup string
	Properties    *armstorage.StorageAccount

	accountKey *string
	name       string
}

func (ad *accountDetails) AccountKey(ctx context.Context, client Client) (*string, error) {
	credentialsLock.Lock()
	defer credentialsLock.Unlock()

	if ad.accountKey != nil {
		return ad.accountKey, nil
	}

	log.Printf("[DEBUG] Cache Miss - looking up the account key for storage account %q..", ad.name)
	props, err := client.AccountsClient.ListKeys(ctx, ad.ResourceGroup, ad.name, nil)
	if err != nil {
		return nil, fmt.Errorf("Error Listing Keys for Storage Account %q (Resource Group %q): %+v", ad.name, ad.ResourceGroup, err)
	}

	if props.StorageAccountListKeysResult.Keys == nil || len(*props.StorageAccountListKeysResult.Keys) == 0 || (*props.StorageAccountListKeysResult.Keys)[0].Value == nil {
		return nil, fmt.Errorf("Keys were nil for Storage Account %q (Resource Group %q): %+v", ad.name, ad.ResourceGroup, err)
	}

	keys := *props.StorageAccountListKeysResult.Keys
	ad.accountKey = keys[0].Value

	// force-cache this
	storageAccountsCache[ad.name] = *ad

	return ad.accountKey, nil
}

func (client Client) AddToCache(accountName string, props armstorage.StorageAccount) error {
	accountsLock.Lock()
	defer accountsLock.Unlock()

	account, err := populateAccountDetails(accountName, props)
	if err != nil {
		return err
	}

	storageAccountsCache[accountName] = *account

	return nil
}

func (client Client) RemoveAccountFromCache(accountName string) {
	accountsLock.Lock()
	delete(storageAccountsCache, accountName)
	accountsLock.Unlock()
}

func (client Client) FindAccount(ctx context.Context, accountName string) (*accountDetails, error) {
	accountsLock.Lock()
	defer accountsLock.Unlock()

	if existing, ok := storageAccountsCache[accountName]; ok {
		return &existing, nil
	}

	accountsPage := client.AccountsClient.List(nil)
	if accountsPage.Err() != nil {
		return nil, fmt.Errorf("Error retrieving storage accounts: %+v", accountsPage.Err())
	}

	var accounts []armstorage.StorageAccount
	if listResult := accountsPage.PageResponse().StorageAccountListResult; listResult != nil && listResult.Value != nil {
		for _, v := range *accountsPage.PageResponse().StorageAccountListResult.Value {
			accounts = append(accounts, v)
		}
	}
	for accountsPage.NextPage(ctx) {
		if listResult := accountsPage.PageResponse().StorageAccountListResult; listResult != nil && listResult.Value != nil {
			for _, v := range *accountsPage.PageResponse().StorageAccountListResult.Value {
				accounts = append(accounts, v)
			}
		}
	}

	for _, v := range accounts {
		if v.TrackedResource.Name == nil {
			continue
		}

		account, err := populateAccountDetails(*v.Name, v)
		if err != nil {
			return nil, err
		}

		storageAccountsCache[*v.Name] = *account
	}

	if existing, ok := storageAccountsCache[accountName]; ok {
		return &existing, nil
	}

	return nil, nil
}

func populateAccountDetails(accountName string, props armstorage.StorageAccount) (*accountDetails, error) {
	if props.ID == nil {
		return nil, fmt.Errorf("`id` was nil for Account %q", accountName)
	}

	accountId := *props.ID
	id, err := parse.StorageAccountID(accountId)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %q as a Resource ID: %+v", accountId, err)
	}

	return &accountDetails{
		name:          accountName,
		ID:            accountId,
		ResourceGroup: id.ResourceGroup,
		Properties:    &props,
	}, nil
}
