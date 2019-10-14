package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/storage/mgmt/storage"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

var (
	// TODO: remove me
	accountKeysCache        = map[string]string{}
	resourceGroupNamesCache = map[string]string{}

	storageAccountsCache = map[string]accountDetails{}
	writeLock            = sync.RWMutex{}
)

type accountDetails struct {
	resourceGroup string
	properties    *storage.AccountProperties
	credentials   *storage.AccountListKeysResult
}

func (client Client) ClearFromCache(resourceGroup, accountName string) {
	writeLock.Lock()

	log.Printf("[DEBUG] Removing Account %q (Resource Group %q) from the cache", accountName, resourceGroup)
	accountCacheKey := fmt.Sprintf("%s-%s", resourceGroup, accountName)
	delete(accountKeysCache, accountCacheKey)

	resourceGroupsCacheKey := accountName
	delete(resourceGroupNamesCache, resourceGroupsCacheKey)

	log.Printf("[DEBUG] Removed Account %q (Resource Group %q) from the cache", accountName, resourceGroup)
	writeLock.Unlock()
}

func (client Client) FindResourceGroup(ctx context.Context, accountName string) (*string, error) {
	cacheKey := accountName
	if v, ok := resourceGroupNamesCache[cacheKey]; ok {
		return &v, nil
	}

	log.Printf("[DEBUG] Cache Miss - looking up the resource group for storage account %q..", accountName)
	writeLock.Lock()
	accounts, err := client.AccountsClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error listing Storage Accounts (to find Resource Group for %q): %s", accountName, err)
	}

	if accounts.Value == nil {
		return nil, nil
	}

	var resourceGroup *string
	for _, account := range *accounts.Value {
		if account.Name == nil || account.ID == nil {
			continue
		}

		if strings.EqualFold(accountName, *account.Name) {
			id, err := azure.ParseAzureResourceID(*account.ID)
			if err != nil {
				return nil, fmt.Errorf("Error parsing ID for Storage Account %q: %s", accountName, err)
			}

			resourceGroup = &id.ResourceGroup
			break
		}
	}

	if resourceGroup != nil {
		resourceGroupNamesCache[cacheKey] = *resourceGroup
	}

	writeLock.Unlock()

	return resourceGroup, nil
}

func (client Client) findAccountKey(ctx context.Context, resourceGroup, accountName string) (*string, error) {
	cacheKey := fmt.Sprintf("%s-%s", resourceGroup, accountName)
	if v, ok := accountKeysCache[cacheKey]; ok {
		return &v, nil
	}

	writeLock.Lock()
	log.Printf("[DEBUG] Cache Miss - looking up the account key for storage account %q..", accountName)
	props, err := client.AccountsClient.ListKeys(ctx, resourceGroup, accountName)
	if err != nil {
		return nil, fmt.Errorf("Error Listing Keys for Storage Account %q (Resource Group %q): %+v", accountName, resourceGroup, err)
	}

	if props.Keys == nil || len(*props.Keys) == 0 {
		return nil, fmt.Errorf("Keys were nil for Storage Account %q (Resource Group %q): %+v", accountName, resourceGroup, err)
	}

	keys := *props.Keys
	firstKey := keys[0].Value

	accountKeysCache[cacheKey] = *firstKey
	writeLock.Unlock()

	return firstKey, nil
}

// TODO: we also need to fire an event to/back from the storage accounts resource that one's been added/removed
func (client Client) ForceCache(accountName string, props *storage.AccountProperties) {
	writeLock.Lock()
	storageAccountsCache[accountName] = props
	writeLock.Unlock()
}

func (client Client) RemoveAccountFromCache(accountName string) {
	writeLock.Lock()
	delete(storageAccountsCache, accountName)
	writeLock.Unlock()
}

func (client Client) findAccount(ctx context.Context, accountName string) (*accountDetails, error) {
	writeLock.Lock()
	defer writeLock.Unlock()

	if existing, ok := storageAccountsCache[accountName]; ok {
		// if the resource group/credentials are nil, load them now?
		storageAccountsCache[accountName] = accountName
		return &existing, nil
	}

	accounts, err := client.AccountsClient.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving storage accounts: %+v", err)
	}

	if accounts.Value == nil {
		return nil, fmt.Errorf("Error loading storage accounts: accounts was nil!")
	}

	for _, v := range *accounts.Value {
		if v.Name == nil {
			continue
		}

		// TODO: this won't work since we need to conditionally cache these at access time?!

		storageAccountsCache[*v.Name] = &accountDetails{
			resourceGroup: resourceGroup,
			properties:    v,
			credentials:   credentials,
		}
	}

	if existing, ok := storageAccountsCache[accountName]; ok {
		return &existing, nil
	}

	return nil, nil
}
