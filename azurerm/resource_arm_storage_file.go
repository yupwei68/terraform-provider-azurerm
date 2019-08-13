package azurerm

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/file/files"
)

func resourceArmStorageFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmStorageFileCreate,
		Read:   resourceArmStorageFileRead,
		Update: resourceArmStorageFileUpdate,
		Delete: resourceArmStorageFileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"storage_account_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArmStorageAccountName,
			},

			"metadata": storage.MetaDataSchema(),
		},
	}
}

func resourceArmStorageFileCreate(d *schema.ResourceData, meta interface{}) error {
	fileClient := meta.(*ArmClient).storage.FilesClient
	ctx := meta.(*ArmClient).StopContext

	fileName := d.Get("name").(string)
	accountName := d.Get("storage_account_name").(string)

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := storage.ExpandMetaData(metaDataRaw)

	resourceID := fileClient.GetResourceID(accountName, fileName)
	if requireResourcesToBeImported {
		existing, err := fileClient.GetMetaData(ctx, accountName, fileName)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing File %q (Storage Account %q): %s", fileName, accountName, err)
			}
		}

		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_storage_file", resourceID)
		}
	}

	if _, err := fileClient.Create(ctx, accountName, fileName, metaData); err != nil {
		return fmt.Errorf("Error creating File %q (Account %q): %+v", fileName, accountName, err)
	}

	d.SetId(resourceID)

	return resourceArmStorageFileRead(d, meta)
}

func resourceArmStorageFileUpdate(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*ArmClient).storage
	ctx := meta.(*ArmClient).StopContext

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := storage.ExpandMetaData(metaDataRaw)

	if _, err := storageClient.FilesClient.SetMetaData(ctx, id.AccountName, id.FileName, metaData); err != nil {
		return fmt.Errorf("Error setting MetaData for File %q (Storage Account %q): %s", id.FileName, id.AccountName, err)
	}

	return resourceArmStorageFileRead(d, meta)
}

func resourceArmStorageFileRead(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*ArmClient).storage
	ctx := meta.(*ArmClient).StopContext

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup, err := storageClient.FindResourceGroup(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for File Container %q (Account %s): %s", id.FileName, id.AccountName, err)
	}
	if resourceGroup == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage File %q (Account %s) - assuming removed & removing from state", id.FileName, id.AccountName)
		d.SetId("")
		return nil
	}

	metaData, err := storageClient.FilesClient.GetMetaData(ctx, id.AccountName, id.FileName)
	if err != nil {
		if utils.ResponseWasNotFound(metaData.Response) {
			log.Printf("[INFO] Storage File %q no longer exists, removing from state...", id.FileName)
			d.SetId("")
			return nil
		}

		return nil
	}

	d.Set("name", id.FileName)
	d.Set("storage_account_name", id.AccountName)
	d.Set("resource_group_name", *resourceGroup)

	if err := d.Set("metadata", storage.FlattenMetaData(metaData.MetaData)); err != nil {
		return fmt.Errorf("Error setting `metadata`: %s", err)
	}

	return nil
}

func resourceArmStorageFileDelete(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*ArmClient).storage
	ctx := meta.(*ArmClient).StopContext

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup, err := storageClient.FindResourceGroup(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for Storage File %q (Account %s): %s", id.FileName, id.AccountName, err)
	}
	if resourceGroup == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage File %q (Account %s) - assuming removed & removing from state", id.QueueName, id.AccountName)
		d.SetId("")
		return nil
	}

	if _, err := storageClient.FilesClient.Delete(ctx, id.AccountName, id.FileName); err != nil {
		return fmt.Errorf("Error deleting Storage File %q: %s", id.FileName, err)
	}

	return nil
}
