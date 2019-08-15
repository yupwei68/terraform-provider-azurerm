package azurerm

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/file/files"
)

func resourceArmStorageShareFile() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmStorageShareFileCreate,
		Read:   resourceArmStorageShareFileRead,
		Update: resourceArmStorageShareFileUpdate,
		Delete: resourceArmStorageShareFileDelete,
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

			"share_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"share_directory_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.StorageShareDirectoryNameAndEmpty,
			},

			"content_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				// No more than 1Tb
				ValidateFunc: validation.IntBetween(0, 1000000000000),
			},

			"content_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed: true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"content_encoding": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"content_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"content_md5": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"content_disposition": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validate.NoEmptyStrings,
			},

			"metadata": storage.MetaDataSchema(),
		},
	}
}

func resourceArmStorageShareFileCreate(d *schema.ResourceData, meta interface{}) error {
	ctx := meta.(*ArmClient).StopContext
	storageClient := meta.(*ArmClient).storage

	accountName := d.Get("storage_account_name").(string)
	shareName := d.Get("share_name").(string)
	directoryName := d.Get("share_directory_name").(string)
	fileName := d.Get("name").(string)

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := storage.ExpandMetaData(metaDataRaw)

	resourceGroup, err := storageClient.FindResourceGroup(ctx, accountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q): %s", fileName, accountName, shareName, err)
	}
	if resourceGroup == nil {
		return fmt.Errorf("Unable to locate Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q) - assuming removed & removing from state", fileName, accountName, shareName)
	}

	client, err := storageClient.FilesClient(ctx, *resourceGroup, accountName)
	if err != nil {
		return fmt.Errorf("Error building Storage Share Client: %s", err)
	}

	id := client.GetResourceID(accountName, shareName, directoryName, fileName)

	createInput := files.CreateInput{
		ContentLength:      int64(d.Get("content_length").(int)),
		ContentType:        utils.String(d.Get("content_type").(string)),
		ContentEncoding:    utils.String(d.Get("content_encoding").(string)),
		ContentLanguage:    utils.String(d.Get("content_language").(string)),
		ContentMD5:         utils.String(d.Get("content_md5").(string)),
		ContentDisposition: utils.String(d.Get("content_disposition").(string)),
		MetaData:           metaData,
	}

	if _, err := client.Create(ctx, accountName, shareName, directoryName, fileName, createInput); err != nil {
		return fmt.Errorf("Error creating Storage Share File %q (Storage Account %q / Storage Share %q): %s", fileName, accountName, shareName, err)
	}

	d.SetId(id)

	return resourceArmStorageShareFileRead(d, meta)
}

func resourceArmStorageShareFileUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := meta.(*ArmClient).StopContext
	storageClient := meta.(*ArmClient).storage

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup, err := storageClient.FindResourceGroup(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for Storage Share Directory %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
	}
	if resourceGroup == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage Share Directory %q (Storage Account %q / Storage Share %q)  - assuming removed & removing from state", id.FileName, id.AccountName, id.ShareName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.FilesClient(ctx, *resourceGroup, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error building Storage Share Client: %s", err)
	}

	if d.HasChange("metadata") {
		metaDataRaw := d.Get("metadata").(map[string]interface{})
		metaData := storage.ExpandMetaData(metaDataRaw)

		if _, err := client.SetMetaData(ctx, id.AccountName, id.ShareName, id.DirectoryName, id.FileName, metaData); err != nil {
			return fmt.Errorf("Error setting MetaData for Storage Share File %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
		}
	}

	var updateProperties bool
	props := files.SetPropertiesInput{}

	if d.HasChange("content_length") {
		props.ContentLength = utils.Int64(int64(d.Get("content_length").(int)))
		updateProperties = true
	}

	if d.HasChange("content_type") {
		props.ContentType = utils.String(d.Get("content_type").(string))
		updateProperties = true
	}

	if d.HasChange("content_encoding") {
		props.ContentEncoding = utils.String(d.Get("content_encoding").(string))
		updateProperties = true
	}

	if d.HasChange("content_language") {
		props.ContentLanguage = utils.String(d.Get("content_language").(string))
		updateProperties = true
	}

	if d.HasChange("content_md5") {
		props.ContentMD5 = utils.String(d.Get("content_md5").(string))
		updateProperties = true
	}

	if d.HasChange("content_disposition") {
		props.ContentDisposition = utils.String(d.Get("content_disposition").(string))
		updateProperties = true
	}

	if updateProperties {
		if _, err := client.SetProperties(ctx, id.AccountName, id.ShareName, id.DirectoryName, id.FileName, props); err != nil {
			return fmt.Errorf("Error updating properties for Storage Share File %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
		}
	}

	return resourceArmStorageShareFileRead(d, meta)
}

func resourceArmStorageShareFileRead(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*ArmClient).storage
	ctx := meta.(*ArmClient).StopContext

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup, err := storageClient.FindResourceGroup(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
	}
	if resourceGroup == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q) - assuming removed & removing from state", id.FileName, id.AccountName, id.ShareName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.FilesClient(ctx, *resourceGroup, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error building Storage Share Client: %s", err)
	}

	metaData, err := client.GetMetaData(ctx, id.AccountName, id.ShareName, id.DirectoryName, id.FileName)
	if err != nil {
		if !utils.ResponseWasNotFound(metaData.Response) {
			return fmt.Errorf("Error retrieving `metadata` for Storage Share File %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
		}
	}

	props, err := client.GetProperties(ctx, id.AccountName, id.ShareName, id.DirectoryName, id.FileName)
	if err != nil {
		if utils.ResponseWasNotFound(props.Response) {
			log.Printf("[WARN] Unable to find properties for Storage Share File %q (Storage Account %q / Storage Share %q) - assuming removed & removing from state", id.FileName, id.AccountName, id.ShareName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Storage Share File %q (Account %s)", id.FileName, id.AccountName)
	}

	d.Set("name", id.FileName)
	d.Set("storage_account_name", id.AccountName)
	d.Set("share_name", id.ShareName)
	d.Set("share_directory_name", id.DirectoryName)

	d.Set("content_length", props.ContentLength)
	d.Set("content_type", props.ContentType)
	d.Set("content_encoding", props.ContentEncoding)
	d.Set("content_language", props.ContentLanguage)
	d.Set("content_md5", props.ContentMD5)
	d.Set("content_disposition", props.ContentDisposition)

	if err := d.Set("metadata", storage.FlattenMetaData(metaData.MetaData)); err != nil {
		return fmt.Errorf("Error setting `metadata`: %s", err)
	}

	return nil
}

func resourceArmStorageShareFileDelete(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*ArmClient).storage
	ctx := meta.(*ArmClient).StopContext

	id, err := files.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup, err := storageClient.FindResourceGroup(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error locating Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q): %s", id.FileName, id.AccountName, id.ShareName, err)
	}
	if resourceGroup == nil {
		log.Printf("[WARN] Unable to determine Resource Group for Storage Share File %q (Storage Account %q / Storage Share %q) - assuming removed & removing from state", id.FileName, id.AccountName, id.ShareName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.FilesClient(ctx, *resourceGroup, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error building Storage Share Client: %s", err)
	}

	if _, err := client.Delete(ctx, id.AccountName, id.ShareName, id.DirectoryName, id.FileName); err != nil {
		return fmt.Errorf("Error deleting Storage Share File %q: %s", id.FileName, err)
	}

	return nil
}
