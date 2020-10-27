package storage

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/tombuildsstuff/giovanni/storage/2018-11-09/file/shares"
	"log"
	"time"
)

func resourceArmStorageShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmStorageShareCreate,
		Read:   resourceArmStorageShareRead,
		Update: resourceArmStorageShareUpdate,
		Delete: resourceArmStorageShareDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 2,
		StateUpgraders: []schema.StateUpgrader{
			{
				// this should have been applied from pre-0.12 migration system; backporting just in-case
				Type:    resourceStorageShareStateResourceV0V1().CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStorageShareStateUpgradeV0ToV1,
				Version: 0,
			},
			{
				Type:    resourceStorageShareStateResourceV0V1().CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStorageShareStateUpgradeV1ToV2,
				Version: 1,
			},
		},

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
				ValidateFunc: ValidateArmStorageShareName,
			},

			"storage_account_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"quota": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5120,
				ValidateFunc: validation.IntBetween(1, 102400),
			},

			"metadata": MetaDataComputedSchema(),

			"acl": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 64),
						},
						"access_policy": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"expiry": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"permissions": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
					},
				},
			},

			"enabled_protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.NFS),
					string(storage.SMB),
				}, false),
			},

			"root_squash": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.AllSquash),
					string(storage.NoRootSquash),
					string(storage.RootSquash),
				}, false),
			},

			"access_tier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(storage.ShareAccessTierTransactionOptimized),
				ValidateFunc: validation.StringInSlice([]string{
					string(storage.ShareAccessTierCool),
					string(storage.ShareAccessTierHot),
					string(storage.ShareAccessTierPremium),
					string(storage.ShareAccessTierTransactionOptimized),
				}, false),
			},

			"resource_manager_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"remaining_retention_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}
func resourceArmStorageShareCreate(d *schema.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	accountName := d.Get("storage_account_name").(string)
	shareName := d.Get("name").(string)
	quota := d.Get("quota").(int)

	metaDataRaw := d.Get("metadata").(map[string]interface{})
	metaData := ExpandMetaDataPtr(metaDataRaw)

	account, err := storageClient.FindAccount(ctx, accountName)
	if err != nil {
		return fmt.Errorf("Error retrieving Account %q for Share %q: %s", accountName, shareName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
	}

	client, err := storageClient.FileSharesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("Error building File Share Client: %s", err)
	}

	id := client.GetResourceID(accountName, shareName)

	mgmtFileShareClient := storageClient.MgmtFileSharesClient
	existing, err := mgmtFileShareClient.Get(ctx, account.ResourceGroup, accountName, shareName, "")
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Storage Share %q (Storage Account Name %q / Resource Group %q): %+v", shareName, accountName, account.ResourceGroup, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_storage_share", *existing.ID)
	}

	params := storage.FileShare{
		FileShareProperties: &storage.FileShareProperties{
			Metadata:   metaData,
			ShareQuota: utils.Int32(int32(quota)),
			AccessTier: storage.ShareAccessTier(d.Get("access_tier").(string)),
		},
	}

	if v, ok := d.GetOk("enabled_protocol"); ok {
		params.FileShareProperties.EnabledProtocols = storage.EnabledProtocols(v.(string))
	}

	if v, ok := d.GetOk("root_squash"); ok {
		params.FileShareProperties.RootSquash = storage.RootSquashType(v.(string))
	}

	if _, err := mgmtFileShareClient.Create(ctx, account.ResourceGroup, accountName, shareName, params); err != nil {
		return fmt.Errorf("creating Storage Share %q (Storage Account Name %q / Resource Group %q): %+v", shareName, accountName, account.ResourceGroup, err)
	}

	if v, ok := d.GetOk("acl"); ok {
		aclsRaw := v.(*schema.Set).List()
		acls := expandStorageShareACLs(aclsRaw)

		if _, err := client.SetACL(ctx, accountName, shareName, acls); err != nil {
			return fmt.Errorf("setting ACL's for Share %q (Account %q / Resource Group %q): %+v", shareName, accountName, account.ResourceGroup, err)
		}
	}

	d.SetId(id)
	return resourceArmStorageShareRead(d, meta)
}

func resourceArmStorageShareRead(d *schema.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	id, err := shares.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error retrieving Account %q for Share %q: %s", id.AccountName, id.ShareName, err)
	}
	if account == nil {
		log.Printf("[WARN] Unable to determine Account %q for Storage Share %q - assuming removed & removing from state", id.AccountName, id.ShareName)
		d.SetId("")
		return nil
	}

	client, err := storageClient.FileSharesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("building File Share Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	mgmtFileShareClient := storageClient.MgmtFileSharesClient
	resp, err := mgmtFileShareClient.Get(ctx, account.ResourceGroup, id.AccountName, id.ShareName, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] File Share %q was not found in Account %q / Resource Group %q - assuming removed & removing from state", id.ShareName, id.AccountName, account.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving File Share %q (Account %q / Resource Group %q): %s", id.ShareName, id.AccountName, account.ResourceGroup, err)
	}

	d.Set("name", id.ShareName)
	d.Set("storage_account_name", id.AccountName)
	d.Set("url", client.GetResourceID(id.AccountName, id.ShareName))

	if props := resp.FileShareProperties; props != nil {
		d.Set("quota", props.ShareQuota)
		if err := d.Set("metadata", FlattenMetaDataPtr(props.Metadata)); err != nil {
			return fmt.Errorf("flattening `metadata`: %+v", err)
		}
		d.Set("enabled_protocol", string(props.EnabledProtocols))
		d.Set("access_tier", string(props.AccessTier))
		d.Set("root_squash", string(props.RootSquash))
		d.Set("deleted", props.Deleted)
		d.Set("remaining_retention_days", props.RemainingRetentionDays)
	}

	if resp.FileShareProperties != nil && resp.FileShareProperties.EnabledProtocols != storage.NFS {
		acls, err := client.GetACL(ctx, id.AccountName, id.ShareName)
		if err != nil {
			return fmt.Errorf("retrieving ACL's for File Share %q (Account %q / Resource Group %q): %s", id.ShareName, id.AccountName, account.ResourceGroup, err)
		}
		if err := d.Set("acl", flattenStorageShareACLs(acls)); err != nil {
			return fmt.Errorf("flattening `acl`: %+v", err)
		}
	}

	resourceManagerId := client.GetResourceManagerResourceID(storageClient.SubscriptionId, account.ResourceGroup, id.AccountName, id.ShareName)
	d.Set("resource_manager_id", resourceManagerId)

	return nil
}

func resourceArmStorageShareUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	id, err := shares.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error retrieving Account %q for Share %q: %s", id.AccountName, id.ShareName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}

	client, err := storageClient.FileSharesClient(ctx, *account)
	if err != nil {
		return fmt.Errorf("Error building File Share Client for Storage Account %q (Resource Group %q): %s", id.AccountName, account.ResourceGroup, err)
	}

	mgmtFileShareClient := storageClient.MgmtFileSharesClient

	params := storage.FileShare{
		FileShareProperties: &storage.FileShareProperties{},
	}

	if d.HasChange("quota") {
		params.FileShareProperties.ShareQuota = utils.Int32(int32(d.Get("quota").(int)))
	}

	if d.HasChange("metadata") {
		metaDataRaw := d.Get("metadata").(map[string]interface{})
		params.FileShareProperties.Metadata = ExpandMetaDataPtr(metaDataRaw)
	}

	if d.HasChange("root_squash") {
		params.FileShareProperties.RootSquash = storage.RootSquashType(d.Get("root_squash").(string))
	}

	if d.HasChange("access_tier") {
		params.FileShareProperties.AccessTier = storage.ShareAccessTier(d.Get("access_tier").(string))
	}

	if _, err := mgmtFileShareClient.Update(ctx, account.ResourceGroup, id.AccountName, id.ShareName, params); err != nil {
		return fmt.Errorf("updating File Share %q (Storage Account %q): %s", id.ShareName, id.AccountName, err)
	}

	if d.HasChange("acl") {
		log.Printf("[DEBUG] Updating the ACL's for File Share %q (Storage Account %q)", id.ShareName, id.AccountName)

		aclsRaw := d.Get("acl").(*schema.Set).List()
		acls := expandStorageShareACLs(aclsRaw)

		if _, err := client.SetACL(ctx, id.AccountName, id.ShareName, acls); err != nil {
			return fmt.Errorf("Error updating ACL's for File Share %q (Storage Account %q): %s", id.ShareName, id.AccountName, err)
		}

		log.Printf("[DEBUG] Updated the ACL's for File Share %q (Storage Account %q)", id.ShareName, id.AccountName)
	}

	return resourceArmStorageShareRead(d, meta)
}

func resourceArmStorageShareDelete(d *schema.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()
	storageClient := meta.(*clients.Client).Storage

	id, err := shares.ParseResourceID(d.Id())
	if err != nil {
		return err
	}

	account, err := storageClient.FindAccount(ctx, id.AccountName)
	if err != nil {
		return fmt.Errorf("Error retrieving Account %q for Share %q: %s", id.AccountName, id.ShareName, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", id.AccountName)
	}

	mgmtFileShareClient := storageClient.MgmtFileSharesClient

	if _, err := mgmtFileShareClient.Delete(ctx, account.ResourceGroup, id.AccountName, id.ShareName); err != nil {
		return fmt.Errorf("deleting File Share %q (Storage Account %q / Resource Group %q): %s", id.ShareName, id.AccountName, account.ResourceGroup, err)
	}

	return nil
}

func expandStorageShareACLs(input []interface{}) []shares.SignedIdentifier {
	results := make([]shares.SignedIdentifier, 0)

	for _, v := range input {
		vals := v.(map[string]interface{})

		policies := vals["access_policy"].([]interface{})
		policy := policies[0].(map[string]interface{})

		identifier := shares.SignedIdentifier{
			Id: vals["id"].(string),
			AccessPolicy: shares.AccessPolicy{
				Start:      policy["start"].(string),
				Expiry:     policy["expiry"].(string),
				Permission: policy["permissions"].(string),
			},
		}
		results = append(results, identifier)
	}

	return results
}

func flattenStorageShareACLs(input shares.GetACLResult) []interface{} {
	result := make([]interface{}, 0)

	for _, v := range input.SignedIdentifiers {
		output := map[string]interface{}{
			"id": v.Id,
			"access_policy": []interface{}{
				map[string]interface{}{
					"start":       v.AccessPolicy.Start,
					"expiry":      v.AccessPolicy.Expiry,
					"permissions": v.AccessPolicy.Permission,
				},
			},
		}

		result = append(result, output)
	}

	return result
}
