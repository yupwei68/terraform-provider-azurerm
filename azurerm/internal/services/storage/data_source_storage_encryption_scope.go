package storage

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/parse"
	storageValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceArmStorageEncryptionScope() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmStorageEncryptionScopeRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: storageValidate.StorageEncryptionScopeName,
			},

			"storage_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: storageValidate.StorageAccountID,
			},

			"source": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"key_vault_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceArmStorageEncryptionScopeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	accountId, err := parse.AccountID(d.Get("storage_account_id").(string))
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, accountId.ResourceGroup, accountId.Name, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q) was not found", name, accountId.Name, accountId.ResourceGroup)
		}

		return fmt.Errorf("retrieving Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", name, accountId.Name, accountId.ResourceGroup, err)
	}

	d.SetId(parse.NewEncryptionScopeId(*accountId, name).ID(subscriptionId))

	if props := resp.EncryptionScopeProperties; props != nil {
		d.Set("source", flattenEncryptionScopeSource(props.Source))
		var keyId string
		if kv := props.KeyVaultProperties; kv != nil {
			if kv.KeyURI != nil {
				keyId = *kv.KeyURI
			}
		}
		d.Set("key_vault_key_id", keyId)
	}

	return nil
}
