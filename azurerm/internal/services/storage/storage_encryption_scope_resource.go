package storage

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/arm/storage/2019-06-01/armstorage"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	keyVaultValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/keyvault/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/parse"
	storageValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/validate"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceStorageEncryptionScope() *schema.Resource {
	return &schema.Resource{
		Create: resourceStorageEncryptionScopeCreate,
		Read:   resourceStorageEncryptionScopeRead,
		Update: resourceStorageEncryptionScopeUpdate,
		Delete: resourceStorageEncryptionScopeDelete,

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.EncryptionScopeID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageEncryptionScopeName,
			},

			"storage_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: storageValidate.StorageAccountID,
			},

			"source": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(armstorage.EncryptionScopeSourceMicrosoftKeyVault),
					string(armstorage.EncryptionScopeSourceMicrosoftStorage),
				}, false),
			},

			"key_vault_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: keyVaultValidate.KeyVaultChildID,
			},
		},
	}
}

func resourceStorageEncryptionScopeCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	accountId, err := parse.StorageAccountID(d.Get("storage_account_id").(string))
	if err != nil {
		return err
	}

	resourceId := parse.NewEncryptionScopeID(accountId.SubscriptionId, accountId.ResourceGroup, accountId.Name, name).ID()
	existing, err := client.Get(ctx, accountId.ResourceGroup, accountId.Name, name, nil)
	if err != nil {
		if !utils.Track2ResponseWasNotFound(err) {
			return fmt.Errorf("checking for present of existing Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", name, accountId.Name, accountId.ResourceGroup, err)
		}
	}
	if existing.EncryptionScope.EncryptionScopeProperties != nil && *existing.EncryptionScope.EncryptionScopeProperties.State == armstorage.EncryptionScopeStateEnabled {
		return tf.ImportAsExistsError("azurerm_storage_encryption_scope", resourceId)
	}

	if d.Get("source").(string) == string(armstorage.KeySourceMicrosoftKeyvault) {
		if _, ok := d.GetOk("key_vault_key_id"); !ok {
			return fmt.Errorf("`key_vault_key_id` is required when source is `%s`", string(armstorage.KeySourceMicrosoftKeyvault))
		}
	}

	esSource := armstorage.EncryptionScopeSource(d.Get("source").(string))
	enable := armstorage.EncryptionScopeStateEnabled

	props := armstorage.EncryptionScope{
		EncryptionScopeProperties: &armstorage.EncryptionScopeProperties{
			Source: &esSource,
			State:  &enable,
			KeyVaultProperties: &armstorage.EncryptionScopeKeyVaultProperties{
				KeyURI: utils.String(d.Get("key_vault_key_id").(string)),
			},
		},
	}
	if _, err := client.Put(ctx, accountId.ResourceGroup, accountId.Name, name, props, nil); err != nil {
		return fmt.Errorf("creating Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", name, accountId.Name, accountId.ResourceGroup, err)
	}

	d.SetId(resourceId)
	return resourceStorageEncryptionScopeRead(d, meta)
}

func resourceStorageEncryptionScopeUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}

	if d.Get("source").(string) == string(armstorage.KeySourceMicrosoftKeyvault) {
		if _, ok := d.GetOk("key_vault_key_id"); !ok {
			return fmt.Errorf("`key_vault_key_id` is required when source is `%s`", string(armstorage.KeySourceMicrosoftKeyvault))
		}
	}

	esSource := armstorage.EncryptionScopeSource(d.Get("source").(string))
	enable := armstorage.EncryptionScopeStateEnabled

	props := armstorage.EncryptionScope{
		EncryptionScopeProperties: &armstorage.EncryptionScopeProperties{
			Source: &esSource,
			State:  &enable,
			KeyVaultProperties: &armstorage.EncryptionScopeKeyVaultProperties{
				KeyURI: utils.String(d.Get("key_vault_key_id").(string)),
			},
		},
	}
	if _, err := client.Patch(ctx, id.ResourceGroup, id.StorageAccountName, id.Name, props, nil); err != nil {
		return fmt.Errorf("updating Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	return resourceStorageEncryptionScopeRead(d, meta)
}

func resourceStorageEncryptionScopeRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}
	accountId := parse.NewStorageAccountID(id.SubscriptionId, id.ResourceGroup, id.StorageAccountName)

	resp, err := client.Get(ctx, id.ResourceGroup, id.StorageAccountName, id.Name, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			log.Printf("[INFO] Storage Encryption Scope %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	if resp.EncryptionScope.EncryptionScopeProperties == nil {
		return fmt.Errorf("retrieving Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): `properties` was nil", id.Name, id.StorageAccountName, id.ResourceGroup)
	}

	props := *resp.EncryptionScope.EncryptionScopeProperties
	if *props.State == armstorage.EncryptionScopeStateDisabled {
		log.Printf("[INFO] Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q) does not exist - removing from state", id.Name, id.StorageAccountName, id.ResourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("name", resp.EncryptionScope.Name)
	d.Set("storage_account_id", accountId.ID())
	if props := resp.EncryptionScope.EncryptionScopeProperties; props != nil {
		d.Set("source", flattenEncryptionScopeSource(*props.Source))
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

func resourceStorageEncryptionScopeDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.EncryptionScopesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.EncryptionScopeID(d.Id())
	if err != nil {
		return err
	}

	disable := armstorage.EncryptionScopeStateDisabled

	props := armstorage.EncryptionScope{
		EncryptionScopeProperties: &armstorage.EncryptionScopeProperties{
			State: &disable,
		},
	}

	if _, err = client.Put(ctx, id.ResourceGroup, id.StorageAccountName, id.Name, props, nil); err != nil {
		return fmt.Errorf("disabling Storage Encryption Scope %q (Storage Account Name %q / Resource Group %q): %+v", id.Name, id.StorageAccountName, id.ResourceGroup, err)
	}

	return nil
}

func flattenEncryptionScopeSource(input armstorage.EncryptionScopeSource) string {
	// TODO: file a bug
	// the Storage API differs from every other API in Azure in that these Enum's can be returned case-insensitively
	if strings.EqualFold(string(input), string(armstorage.EncryptionScopeSourceMicrosoftKeyVault)) {
		return string(armstorage.EncryptionScopeSourceMicrosoftKeyVault)
	}

	return string(armstorage.EncryptionScopeSourceMicrosoftKeyVault)
}
