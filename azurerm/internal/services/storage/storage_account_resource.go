package storage

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/arm/storage/2019-06-01/armstorage"
	azautorest "github.com/Azure/go-autorest/autorest"
	autorestAzure "github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/accounts"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/queue/queues"
)

var storageAccountResourceName = "azurerm_storage_account"

func resourceStorageAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceStorageAccountCreate,
		Read:   resourceStorageAccountRead,
		Update: resourceStorageAccountUpdate,
		Delete: resourceStorageAccountDelete,

		MigrateState:  ResourceStorageAccountMigrateState,
		SchemaVersion: 2,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidateStorageAccountName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"account_kind": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(armstorage.KindStorage),
					string(armstorage.KindBlobStorage),
					string(armstorage.KindBlockBlobStorage),
					string(armstorage.KindFileStorage),
					string(armstorage.KindStorageV2),
				}, true),
				Default: string(armstorage.KindStorageV2),
			},

			"account_tier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Standard",
					"Premium",
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"account_replication_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"LRS",
					"ZRS",
					"GRS",
					"RAGRS",
					"GZRS",
					"RAGZRS",
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			// Only valid for BlobStorage & StorageV2 accounts, defaults to "Hot" in create function
			"access_tier": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(armstorage.AccessTierCool),
					string(armstorage.AccessTierHot),
				}, true),
			},

			"custom_domain": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"use_subdomain": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"enable_https_traffic_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"min_tls_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(armstorage.MinimumTLSVersionTLS10),
				ValidateFunc: validation.StringInSlice([]string{
					string(armstorage.MinimumTLSVersionTLS10),
					string(armstorage.MinimumTLSVersionTLS11),
					string(armstorage.MinimumTLSVersionTLS12),
				}, false),
			},

			"is_hns_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"allow_blob_public_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"network_rules": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bypass": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									string(armstorage.BypassAzureServices),
									string(armstorage.BypassLogging),
									string(armstorage.BypassMetrics),
									string(armstorage.BypassNone),
								}, true),
							},
							Set: schema.HashString,
						},

						"ip_rules": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.StorageAccountIpRule,
							},
							Set: schema.HashString,
						},

						"virtual_network_subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"default_action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(armstorage.DefaultActionAllow),
								string(armstorage.DefaultActionDeny),
							}, false),
						},
					},
				},
			},

			"identity": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								"SystemAssigned",
							}, true),
						},
						"principal_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"blob_properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cors_rule": schemaStorageAccountCorsRule(true),
						"delete_retention_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      7,
										ValidateFunc: validation.IntBetween(1, 365),
									},
								},
							},
						},
					},
				},
			},

			"queue_properties": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cors_rule": schemaStorageAccountCorsRule(false),
						"logging": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"version": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"delete": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"read": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"write": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"retention_policy_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 365),
									},
								},
							},
						},
						"hour_metrics": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"version": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"include_apis": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"retention_policy_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 365),
									},
								},
							},
						},
						"minute_metrics": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"version": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"include_apis": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"retention_policy_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 365),
									},
								},
							},
						},
					},
				},
			},

			"static_website": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"error_404_document": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"large_file_share_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"primary_location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_location": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_blob_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_blob_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_blob_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_blob_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_queue_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_queue_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_queue_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_queue_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_table_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_table_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_table_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_table_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_web_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_web_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_web_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_web_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_dfs_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_dfs_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_dfs_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_dfs_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_file_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_file_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_file_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"secondary_file_host": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"primary_access_key": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},

			"secondary_access_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_blob_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_blob_connection_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"tags": {
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateAzureRMStorageAccountTags,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		CustomizeDiff: func(d *schema.ResourceDiff, v interface{}) error {
			if d.HasChange("account_kind") {
				accountKind, changedKind := d.GetChange("account_kind")

				if accountKind != string(armstorage.KindStorage) && changedKind != string(armstorage.KindStorageV2) {
					log.Printf("[DEBUG] recreate storage account, could't be migrated from %s to %s", accountKind, changedKind)
					d.ForceNew("account_kind")
				} else {
					log.Printf("[DEBUG] storage account can be upgraded from %s to %s", accountKind, changedKind)
				}
			}

			if d.HasChange("large_file_share_enabled") {
				lfsEnabled, changedEnabled := d.GetChange("large_file_share_enabled")
				if lfsEnabled.(bool) && !changedEnabled.(bool) {
					return fmt.Errorf("`large_file_share_enabled` cannot be disabled once it's been enabled")
				}
			}

			return nil
		},
	}
}

func validateAzureRMStorageAccountTags(v interface{}, _ string) (warnings []string, errors []error) {
	tagsMap := v.(map[string]interface{})

	if len(tagsMap) > 50 {
		errors = append(errors, fmt.Errorf("a maximum of 50 tags can be applied to storage account ARM resource"))
	}

	for k, v := range tagsMap {
		if len(k) > 128 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag key is 128 characters: %q is %d characters", k, len(k)))
		}

		value, err := tags.TagValueToString(v)
		if err != nil {
			errors = append(errors, err)
		} else if len(value) > 256 {
			errors = append(errors, fmt.Errorf("the maximum length for a tag value is 256 characters: the value for %q is %d characters", k, len(value)))
		}
	}

	return warnings, errors
}

func resourceStorageAccountCreate(d *schema.ResourceData, meta interface{}) error {
	envName := meta.(*clients.Client).Account.Environment.Name
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	storageAccountName := d.Get("name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	locks.ByName(storageAccountName, storageAccountResourceName)
	defer locks.UnlockByName(storageAccountName, storageAccountResourceName)

	existing, err := client.GetProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		if !utils.Track2ResponseWasNotFound(err) {
			return fmt.Errorf("Error checking for presence of existing Storage Account %q (Resource Group %q): %s", storageAccountName, resourceGroupName, err)
		}
	}

	if existing.StorageAccount != nil && existing.StorageAccount.ID != nil && *existing.StorageAccount.ID != "" {
		return tf.ImportAsExistsError("azurerm_storage_account", *existing.StorageAccount.ID)
	}

	accountKind := d.Get("account_kind").(string)
	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})
	enableHTTPSTrafficOnly := d.Get("enable_https_traffic_only").(bool)
	minimumTLSVersion := d.Get("min_tls_version").(string)
	isHnsEnabled := d.Get("is_hns_enabled").(bool)
	allowBlobPublicAccess := d.Get("allow_blob_public_access").(bool)

	accountTier := d.Get("account_tier").(string)
	replicationType := d.Get("account_replication_type").(string)
	storageType := fmt.Sprintf("%s_%s", accountTier, replicationType)

	sku := armstorage.SKUName(storageType)
	kind := armstorage.Kind(accountKind)
	parameters := armstorage.StorageAccountCreateParameters{
		Location: &location,
		SKU: &armstorage.SKU{
			Name: &sku,
		},
		Tags: tags.Track2ExpandString(t),
		Kind: &kind,
		Properties: &armstorage.StorageAccountPropertiesCreateParameters{
			EnableHTTPsTrafficOnly: &enableHTTPSTrafficOnly,
			NetworkRuleSet:         expandStorageAccountNetworkRules(d),
			IsHnsEnabled:           &isHnsEnabled,
		},
	}

	// For all Clouds except Public and USGovernmentCloud, don't specify "allow_blob_public_access" and "min_tls_version" in request body.
	// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7812
	// https://github.com/terraform-providers/terraform-provider-azurerm/issues/8083
	// USGovernmentCloud allow_blob_public_access and min_tls_version allowed as of issue 9128
	// https://github.com/terraform-providers/terraform-provider-azurerm/issues/9128
	if envName != autorestAzure.PublicCloud.Name && envName != autorestAzure.USGovernmentCloud.Name {
		if allowBlobPublicAccess && minimumTLSVersion != string(armstorage.MinimumTLSVersionTLS10) {
			return fmt.Errorf(`"allow_blob_public_access" and "min_tls_version" are not supported for a Storage Account located in %q`, envName)
		}
	} else {
		parameters.Properties.AllowBlobPublicAccess = &allowBlobPublicAccess
		minVersion := armstorage.MinimumTLSVersion(minimumTLSVersion)
		parameters.Properties.MinimumTLSVersion = &minVersion
	}

	if _, ok := d.GetOk("identity"); ok {
		storageAccountIdentity := expandAzureRmStorageAccountIdentity(d)
		parameters.IDentity = storageAccountIdentity
	}

	if _, ok := d.GetOk("custom_domain"); ok {
		parameters.Properties.CustomDomain = expandStorageAccountCustomDomain(d)
	}

	// BlobStorage does not support ZRS
	if accountKind == string(armstorage.KindBlobStorage) {
		if *parameters.SKU.Name == armstorage.SKUNameStandardZrs {
			return fmt.Errorf("A `account_replication_type` of `ZRS` isn't supported for Blob Storage accounts.")
		}
	}

	// AccessTier is only valid for BlobStorage, StorageV2, and FileStorage accounts
	if accountKind == string(armstorage.KindBlobStorage) || accountKind == string(armstorage.KindStorageV2) || accountKind == string(armstorage.KindFileStorage) {
		accessTier, ok := d.GetOk("access_tier")
		if !ok {
			// default to "Hot"
			accessTier = string(armstorage.AccessTierHot)
		}

		tier := armstorage.AccessTier(accessTier.(string))
		parameters.Properties.AccessTier = &tier
	} else if isHnsEnabled && accountKind != string(armstorage.KindBlockBlobStorage) {
		return fmt.Errorf("`is_hns_enabled` can only be used with account kinds `StorageV2`, `BlobStorage` and `BlockBlobStorage`")
	}

	// AccountTier must be Premium for FileStorage
	if accountKind == string(armstorage.KindFileStorage) {
		if parameters.SKU != nil && parameters.SKU.Tier != nil && *parameters.SKU.Tier == armstorage.SKUTierStandard {
			return fmt.Errorf("A `account_tier` of `Standard` is not supported for FileStorage accounts.")
		}
	}

	// nolint staticcheck
	if v, ok := d.GetOkExists("large_file_share_enabled"); ok {
		disableLargeFile := armstorage.LargeFileSharesStateDisabled
		enableLargeFile := armstorage.LargeFileSharesStateEnabled
		parameters.Properties.LargeFileSharesState = &disableLargeFile
		if v.(bool) {
			parameters.Properties.LargeFileSharesState = &enableLargeFile
		}
	}

	// Create
	future, err := client.BeginCreate(ctx, resourceGroupName, storageAccountName, parameters, nil)
	if err != nil {
		return fmt.Errorf("Error creating Azure Storage Account %q: %+v", storageAccountName, err)
	}

	if _, err = future.PollUntilDone(ctx, common.DefaultPollingInterval); err != nil {
		return fmt.Errorf("Error waiting for Azure Storage Account %q to be created: %+v", storageAccountName, err)
	}

	account, err := client.GetProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return fmt.Errorf("Error retrieving Azure Storage Account %q: %+v", storageAccountName, err)
	}

	if account.StorageAccount.ID == nil {
		return fmt.Errorf("Cannot read Storage Account %q (resource group %q) ID",
			storageAccountName, resourceGroupName)
	}
	log.Printf("[INFO] storage account %q ID: %q", storageAccountName, *account.StorageAccount.ID)
	d.SetId(*account.StorageAccount.ID)

	if val, ok := d.GetOk("blob_properties"); ok {
		// FileStorage does not support blob settings
		if accountKind != string(armstorage.KindFileStorage) {
			blobClient := meta.(*clients.Client).Storage.BlobServicesClient

			blobProperties := expandBlobProperties(val.([]interface{}))

			if _, err = blobClient.SetServiceProperties(ctx, resourceGroupName, storageAccountName, blobProperties, nil); err != nil {
				return fmt.Errorf("Error updating Azure Storage Account `blob_properties` %q: %+v", storageAccountName, err)
			}
		} else {
			return fmt.Errorf("`blob_properties` aren't supported for File Storage accounts.")
		}
	}

	if val, ok := d.GetOk("queue_properties"); ok {
		storageClient := meta.(*clients.Client).Storage
		account, err := storageClient.FindAccount(ctx, storageAccountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q: %s", storageAccountName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", storageAccountName)
		}

		queueClient, err := storageClient.QueuesClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building Queues Client: %s", err)
		}

		queueProperties, err := expandQueueProperties(val.([]interface{}))
		if err != nil {
			return fmt.Errorf("Error expanding `queue_properties` for Azure Storage Account %q: %+v", storageAccountName, err)
		}

		if err = queueClient.UpdateServiceProperties(ctx, account.ResourceGroup, storageAccountName, queueProperties); err != nil {
			return fmt.Errorf("updating Queue Properties for Storage Account %q: %+v", storageAccountName, err)
		}
	}

	if val, ok := d.GetOk("static_website"); ok {
		// static website only supported on StorageV2 and BlockBlobStorage
		if accountKind != string(armstorage.KindStorageV2) && accountKind != string(armstorage.KindBlockBlobStorage) {
			return fmt.Errorf("`static_website` is only supported for StorageV2 and BlockBlobarmstorage.")
		}
		storageClient := meta.(*clients.Client).Storage

		account, err := storageClient.FindAccount(ctx, storageAccountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q: %s", storageAccountName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", storageAccountName)
		}

		accountsClient, err := storageClient.AccountsDataPlaneClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building Accounts Data Plane Client: %s", err)
		}

		staticWebsiteProps := expandStaticWebsiteProperties(val.([]interface{}))

		if _, err = accountsClient.SetServiceProperties(ctx, storageAccountName, staticWebsiteProps); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account `static_website` %q: %+v", storageAccountName, err)
		}
	}

	return resourceStorageAccountRead(d, meta)
}

func resourceStorageAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	envName := meta.(*clients.Client).Account.Environment.Name
	client := meta.(*clients.Client).Storage.AccountsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	storageAccountName := id.Path["storageAccounts"]
	resourceGroupName := id.ResourceGroup

	locks.ByName(storageAccountName, storageAccountResourceName)
	defer locks.UnlockByName(storageAccountName, storageAccountResourceName)

	accountTier := d.Get("account_tier").(string)
	replicationType := d.Get("account_replication_type").(string)
	storageType := fmt.Sprintf("%s_%s", accountTier, replicationType)
	accountKind := d.Get("account_kind").(string)

	if accountKind == string(armstorage.KindBlobStorage) {
		if storageType == string(armstorage.SKUNameStandardZrs) {
			return fmt.Errorf("A `account_replication_type` of `ZRS` isn't supported for Blob Storage accounts.")
		}
	}

	if d.HasChange("account_replication_type") {
		skuName := armstorage.SKUName(storageType)
		sku := armstorage.SKU{
			Name: &skuName,
		}

		opts := armstorage.StorageAccountUpdateParameters{
			SKU: &sku,
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account type %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("account_kind") {
		kind := armstorage.Kind(accountKind)
		opts := armstorage.StorageAccountUpdateParameters{
			Kind: &kind,
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account account_kind %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("access_tier") {
		accessTier := armstorage.AccessTier(d.Get("access_tier").(string))

		opts := armstorage.StorageAccountUpdateParameters{
			Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
				AccessTier: &accessTier,
			},
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account access_tier %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("tags") {
		t := d.Get("tags").(map[string]interface{})

		opts := armstorage.StorageAccountUpdateParameters{
			Tags: tags.Track2ExpandString(t),
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account tags %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("custom_domain") {
		opts := armstorage.StorageAccountUpdateParameters{
			Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
				CustomDomain: expandStorageAccountCustomDomain(d),
			},
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account Custom Domain %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("enable_https_traffic_only") {
		enableHTTPSTrafficOnly := d.Get("enable_https_traffic_only").(bool)

		opts := armstorage.StorageAccountUpdateParameters{
			Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
				EnableHTTPsTrafficOnly: &enableHTTPSTrafficOnly,
			},
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account enable_https_traffic_only %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("min_tls_version") {
		minimumTLSVersion := d.Get("min_tls_version").(string)

		// For all Clouds except Public and USGovernmentCloud, don't specify "min_tls_version" in request body.
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/8083
		// USGovernmentCloud "min_tls_version" allowed as of issue 9128
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/9128
		if envName != autorestAzure.PublicCloud.Name && envName != autorestAzure.USGovernmentCloud.Name {
			if minimumTLSVersion != string(armstorage.MinimumTLSVersionTLS10) {
				return fmt.Errorf(`"min_tls_version" is not supported for a Storage Account located in %q`, envName)
			}
		} else {
			minVersion := armstorage.MinimumTLSVersion(minimumTLSVersion)
			opts := armstorage.StorageAccountUpdateParameters{
				Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
					MinimumTLSVersion: &minVersion,
				},
			}

			if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
				return fmt.Errorf("Error updating Azure Storage Account min_tls_version %q: %+v", storageAccountName, err)
			}
		}
	}

	if d.HasChange("allow_blob_public_access") {
		allowBlobPublicAccess := d.Get("allow_blob_public_access").(bool)

		// For all Clouds except Public and USGovernmentCloud, don't specify "allow_blob_public_access" in request body.
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7812
		// USGovernmentCloud "allow_blob_public_access" allowed as of issue 9128
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/9128
		if envName != autorestAzure.PublicCloud.Name && envName != autorestAzure.USGovernmentCloud.Name {
			if allowBlobPublicAccess {
				return fmt.Errorf(`"allow_blob_public_access" is not supported for a Storage Account located in %q`, envName)
			}
		} else {
			opts := armstorage.StorageAccountUpdateParameters{
				Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
					AllowBlobPublicAccess: &allowBlobPublicAccess,
				},
			}

			if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
				return fmt.Errorf("Error updating Azure Storage Account allow_blob_public_access %q: %+v", storageAccountName, err)
			}
		}
	}

	if d.HasChange("identity") {
		opts := armstorage.StorageAccountUpdateParameters{
			IDentity: expandAzureRmStorageAccountIdentity(d),
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account identity %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("network_rules") {
		opts := armstorage.StorageAccountUpdateParameters{
			Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
				NetworkRuleSet: expandStorageAccountNetworkRules(d),
			},
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account network_rules %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("large_file_share_enabled") {
		isEnabled := armstorage.LargeFileSharesStateDisabled
		if v := d.Get("large_file_share_enabled").(bool); v {
			isEnabled = armstorage.LargeFileSharesStateEnabled
		}
		opts := armstorage.StorageAccountUpdateParameters{
			Properties: &armstorage.StorageAccountPropertiesUpdateParameters{
				LargeFileSharesState: &isEnabled,
			},
		}

		if _, err := client.Update(ctx, resourceGroupName, storageAccountName, opts, nil); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account network_rules %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("blob_properties") {
		// FileStorage does not support blob settings
		if accountKind != string(armstorage.KindFileStorage) {
			blobClient := meta.(*clients.Client).Storage.BlobServicesClient
			blobProperties := expandBlobProperties(d.Get("blob_properties").([]interface{}))

			if _, err = blobClient.SetServiceProperties(ctx, resourceGroupName, storageAccountName, blobProperties, nil); err != nil {
				return fmt.Errorf("Error updating Azure Storage Account `blob_properties` %q: %+v", storageAccountName, err)
			}
		} else {
			return fmt.Errorf("`blob_properties` aren't supported for File Storage accounts.")
		}
	}

	if d.HasChange("queue_properties") {
		storageClient := meta.(*clients.Client).Storage
		account, err := storageClient.FindAccount(ctx, storageAccountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q: %s", storageAccountName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", storageAccountName)
		}

		queueClient, err := storageClient.QueuesClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building Queues Client: %s", err)
		}

		queueProperties, err := expandQueueProperties(d.Get("queue_properties").([]interface{}))
		if err != nil {
			return fmt.Errorf("Error expanding `queue_properties` for Azure Storage Account %q: %+v", storageAccountName, err)
		}

		if err = queueClient.UpdateServiceProperties(ctx, account.ResourceGroup, storageAccountName, queueProperties); err != nil {
			return fmt.Errorf("updating Queue Properties for Storage Account %q: %+v", storageAccountName, err)
		}
	}

	if d.HasChange("static_website") {
		// static website only supported on StorageV2 and BlockBlobStorage
		if accountKind != string(armstorage.KindStorageV2) && accountKind != string(armstorage.KindBlockBlobStorage) {
			return fmt.Errorf("`static_website` is only supported for StorageV2 and BlockBlobarmstorage.")
		}
		storageClient := meta.(*clients.Client).Storage

		account, err := storageClient.FindAccount(ctx, storageAccountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q: %s", storageAccountName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", storageAccountName)
		}

		accountsClient, err := storageClient.AccountsDataPlaneClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building Accounts Data Plane Client: %s", err)
		}

		staticWebsiteProps := expandStaticWebsiteProperties(d.Get("static_website").([]interface{}))

		if _, err = accountsClient.SetServiceProperties(ctx, storageAccountName, staticWebsiteProps); err != nil {
			return fmt.Errorf("Error updating Azure Storage Account `static_website` %q: %+v", storageAccountName, err)
		}
	}

	return resourceStorageAccountRead(d, meta)
}

func resourceStorageAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Storage.AccountsClient
	endpointSuffix := meta.(*clients.Client).Account.Environment.StorageEndpointSuffix
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	name := id.Path["storageAccounts"]
	resGroup := id.ResourceGroup

	resp, err := client.GetProperties(ctx, resGroup, name, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading the state of AzureRM Storage Account %q: %+v", name, err)
	}

	// handle the user not having permissions to list the keys
	d.Set("primary_connection_string", "")
	d.Set("secondary_connection_string", "")
	d.Set("primary_blob_connection_string", "")
	d.Set("secondary_blob_connection_string", "")
	d.Set("primary_access_key", "")
	d.Set("secondary_access_key", "")

	keys, err := client.ListKeys(ctx, resGroup, name, nil)
	if err != nil {
		// the API returns a 200 with an inner error of a 409..
		var hasWriteLock bool
		var doesntHavePermissions bool
		if e, ok := err.(azautorest.DetailedError); ok {
			if status, ok := e.StatusCode.(int); ok {
				hasWriteLock = status == http.StatusConflict
				doesntHavePermissions = status == http.StatusUnauthorized
			}
		}

		if !hasWriteLock && !doesntHavePermissions {
			return fmt.Errorf("Error listing Keys for Storage Account %q (Resource Group %q): %s", name, resGroup, err)
		}
	}

	d.Set("name", resp.StorageAccount.Name)
	d.Set("resource_group_name", resGroup)
	if location := resp.StorageAccount.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}
	d.Set("account_kind", resp.StorageAccount.Kind)

	if sku := resp.StorageAccount.SKU; sku != nil {
		d.Set("account_tier", sku.Tier)
		d.Set("account_replication_type", strings.Split(fmt.Sprintf("%v", *sku.Name), "_")[1])
	}

	if props := resp.StorageAccount.Properties; props != nil {
		d.Set("access_tier", props.AccessTier)
		d.Set("enable_https_traffic_only", props.EnableHTTPsTrafficOnly)
		d.Set("is_hns_enabled", props.IsHnsEnabled)
		d.Set("allow_blob_public_access", props.AllowBlobPublicAccess)
		// For all Clouds except Public and USGovernmentCloud, "min_tls_version" is not returned from Azure so always persist the default values for "min_tls_version".
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/7812
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/8083
		// USGovernmentCloud "min_tls_version" allowed as of issue 9128
		// https://github.com/terraform-providers/terraform-provider-azurerm/issues/9128
		envName := meta.(*clients.Client).Account.Environment.Name
		if envName != autorestAzure.PublicCloud.Name && envName != autorestAzure.USGovernmentCloud.Name {
			d.Set("min_tls_version", string(armstorage.MinimumTLSVersionTLS10))
		} else {
			// For storage account created using old API, the response of GET call will not return "min_tls_version", either.
			minTlsVersion := string(armstorage.MinimumTLSVersionTLS10)
			if props.MinimumTLSVersion != nil {
				minTlsVersion = string(*props.MinimumTLSVersion)
			}
			d.Set("min_tls_version", minTlsVersion)
		}

		if customDomain := props.CustomDomain; customDomain != nil {
			if err := d.Set("custom_domain", flattenStorageAccountCustomDomain(customDomain)); err != nil {
				return fmt.Errorf("Error setting `custom_domain`: %+v", err)
			}
		}

		// Computed
		d.Set("primary_location", props.PrimaryLocation)
		d.Set("secondary_location", props.SecondaryLocation)

		if accessKeys := keys.StorageAccountListKeysResult.Keys; accessKeys != nil {
			storageAccountKeys := *accessKeys
			if len(storageAccountKeys) > 0 {
				pcs := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=%s", *resp.StorageAccount.Name, *storageAccountKeys[0].Value, endpointSuffix)
				d.Set("primary_connection_string", pcs)
			}

			if len(storageAccountKeys) > 1 {
				scs := fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=%s", *resp.StorageAccount.Name, *storageAccountKeys[1].Value, endpointSuffix)
				d.Set("secondary_connection_string", scs)
			}
		}

		if err := flattenAndSetAzureRmStorageAccountPrimaryEndpoints(d, props.PrimaryEndpoints); err != nil {
			return fmt.Errorf("error setting primary endpoints and hosts for blob, queue, table and file: %+v", err)
		}

		if accessKeys := keys.StorageAccountListKeysResult.Keys; accessKeys != nil {
			storageAccountKeys := *accessKeys
			var primaryBlobConnectStr string
			if v := props.PrimaryEndpoints; v != nil {
				primaryBlobConnectStr = getBlobConnectionString(v.Blob, resp.StorageAccount.Name, storageAccountKeys[0].Value)
			}
			d.Set("primary_blob_connection_string", primaryBlobConnectStr)
		}

		if err := flattenAndSetAzureRmStorageAccountSecondaryEndpoints(d, props.SecondaryEndpoints); err != nil {
			return fmt.Errorf("error setting secondary endpoints and hosts for blob, queue, table: %+v", err)
		}

		if accessKeys := keys.StorageAccountListKeysResult.Keys; accessKeys != nil {
			storageAccountKeys := *accessKeys
			var secondaryBlobConnectStr string
			if v := props.SecondaryEndpoints; v != nil {
				secondaryBlobConnectStr = getBlobConnectionString(v.Blob, resp.StorageAccount.Name, storageAccountKeys[1].Value)
			}
			d.Set("secondary_blob_connection_string", secondaryBlobConnectStr)
		}

		if err := d.Set("network_rules", flattenStorageAccountNetworkRules(props.NetworkRuleSet)); err != nil {
			return fmt.Errorf("Error setting `network_rules`: %+v", err)
		}

		if props.LargeFileSharesState != nil {
			d.Set("large_file_share_enabled", *props.LargeFileSharesState == armstorage.LargeFileSharesStateEnabled)
		}
	}

	if accessKeys := keys.StorageAccountListKeysResult.Keys; accessKeys != nil {
		storageAccountKeys := *accessKeys
		d.Set("primary_access_key", storageAccountKeys[0].Value)
		d.Set("secondary_access_key", storageAccountKeys[1].Value)
	}

	identity := flattenAzureRmStorageAccountIdentity(resp.StorageAccount.IDentity)
	if err := d.Set("identity", identity); err != nil {
		return err
	}

	storageClient := meta.(*clients.Client).Storage
	account, err := storageClient.FindAccount(ctx, name)
	if err != nil {
		return fmt.Errorf("Error retrieving Account %q: %s", name, err)
	}
	if account == nil {
		return fmt.Errorf("Unable to locate Storage Account %q!", name)
	}

	blobClient := storageClient.BlobServicesClient

	// FileStorage does not support blob settings
	if *resp.StorageAccount.Kind != armstorage.KindFileStorage {
		blobProps, err := blobClient.GetServiceProperties(ctx, resGroup, name, nil)
		if err != nil {
			if !utils.Track2ResponseWasNotFound(err) {
				return fmt.Errorf("Error reading blob properties for AzureRM Storage Account %q: %+v", name, err)
			}
		}

		if err := d.Set("blob_properties", flattenBlobProperties(*blobProps.BlobServiceProperties)); err != nil {
			return fmt.Errorf("Error setting `blob_properties `for AzureRM Storage Account %q: %+v", name, err)
		}
	}

	// queue is only available for certain tier and kind (as specified below)
	if resp.StorageAccount.SKU == nil {
		return fmt.Errorf("Error retrieving Storage Account %q (Resource Group %q): `sku` was nil", name, resGroup)
	}

	if *resp.StorageAccount.SKU.Tier == armstorage.SKUTierStandard {
		if *resp.StorageAccount.Kind == armstorage.KindStorage || *resp.StorageAccount.Kind == armstorage.KindStorageV2 {
			queueClient, err := storageClient.QueuesClient(ctx, *account)
			if err != nil {
				return fmt.Errorf("Error building Queues Client: %s", err)
			}

			queueProps, err := queueClient.GetServiceProperties(ctx, account.ResourceGroup, name)
			if err != nil {
				return fmt.Errorf("Error reading queue properties for AzureRM Storage Account %q: %+v", name, err)
			}

			if err := d.Set("queue_properties", flattenQueueProperties(queueProps)); err != nil {
				return fmt.Errorf("setting `queue_properties`: %+v", err)
			}
		}
	}

	var staticWebsite []interface{}

	// static website only supported on StorageV2 and BlockBlobStorage
	if *resp.StorageAccount.Kind == armstorage.KindStorageV2 || *resp.StorageAccount.Kind == armstorage.KindBlockBlobStorage {
		storageClient := meta.(*clients.Client).Storage

		account, err := storageClient.FindAccount(ctx, name)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q: %s", name, err)
		}

		accountsClient, err := storageClient.AccountsDataPlaneClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building Accounts Data Plane Client: %s", err)
		}

		staticWebsiteProps, err := accountsClient.GetServiceProperties(ctx, name)
		if err != nil {
			if staticWebsiteProps.Response.Response != nil && !utils.ResponseWasNotFound(staticWebsiteProps.Response) {
				return fmt.Errorf("Error reading static website for AzureRM Storage Account %q: %+v", name, err)
			}
		}

		staticWebsite = flattenStaticWebsiteProperties(staticWebsiteProps)
	}

	if err := d.Set("static_website", staticWebsite); err != nil {
		return fmt.Errorf("Error setting `static_website `for AzureRM Storage Account %q: %+v", name, err)
	}

	return tags.Track2FlattenAndSetString(d, resp.StorageAccount.Tags)
}

func resourceStorageAccountDelete(d *schema.ResourceData, meta interface{}) error {
	storageClient := meta.(*clients.Client).Storage
	client := storageClient.AccountsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	name := id.Path["storageAccounts"]
	resourceGroup := id.ResourceGroup

	locks.ByName(name, storageAccountResourceName)
	defer locks.UnlockByName(name, storageAccountResourceName)

	read, err := client.GetProperties(ctx, resourceGroup, name, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			return nil
		}

		return fmt.Errorf("Error retrieving Storage Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	// the networking api's only allow a single change to be made to a network layout at once, so let's lock to handle that
	virtualNetworkNames := make([]string, 0)
	if props := read.StorageAccount.Properties; props != nil {
		if rules := props.NetworkRuleSet; rules != nil {
			if vnr := rules.VirtualNetworkRules; vnr != nil {
				for _, v := range *vnr {
					if v.VirtualNetworkResourceID == nil {
						continue
					}

					id, err2 := azure.ParseAzureResourceID(*v.VirtualNetworkResourceID)
					if err2 != nil {
						return err2
					}

					networkName := id.Path["virtualNetworks"]
					for _, virtualNetworkName := range virtualNetworkNames {
						if networkName == virtualNetworkName {
							continue
						}
					}
					virtualNetworkNames = append(virtualNetworkNames, networkName)
				}
			}
		}
	}

	locks.MultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)
	defer locks.UnlockMultipleByName(&virtualNetworkNames, network.VirtualNetworkResourceName)

	_, err = client.Delete(ctx, resourceGroup, name, nil)
	if err != nil {
		if !utils.Track2ResponseWasNotFound(err) {
			return fmt.Errorf("Error issuing delete request for Storage Account %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	// remove this from the cache
	storageClient.RemoveAccountFromCache(name)

	return nil
}

func expandStorageAccountCustomDomain(d *schema.ResourceData) *armstorage.CustomDomain {
	domains := d.Get("custom_domain").([]interface{})
	if len(domains) == 0 {
		return &armstorage.CustomDomain{
			Name: utils.String(""),
		}
	}

	domain := domains[0].(map[string]interface{})
	name := domain["name"].(string)
	useSubDomain := domain["use_subdomain"].(bool)
	return &armstorage.CustomDomain{
		Name:             utils.String(name),
		UseSubDomainName: utils.Bool(useSubDomain),
	}
}

func flattenStorageAccountCustomDomain(input *armstorage.CustomDomain) []interface{} {
	domain := make(map[string]interface{})

	if v := input.Name; v != nil {
		domain["name"] = *v
	}

	// use_subdomain isn't returned
	return []interface{}{domain}
}

func expandStorageAccountNetworkRules(d *schema.ResourceData) *armstorage.NetworkRuleSet {
	networkRules := d.Get("network_rules").([]interface{})
	if len(networkRules) == 0 {
		// Default access is enabled when no network rules are set.
		defaultActionAllow := armstorage.DefaultActionAllow
		return &armstorage.NetworkRuleSet{DefaultAction: &defaultActionAllow}
	}

	networkRule := networkRules[0].(map[string]interface{})
	networkRuleSet := &armstorage.NetworkRuleSet{
		IPRules:             expandStorageAccountIPRules(networkRule),
		VirtualNetworkRules: expandStorageAccountVirtualNetworks(networkRule),
		Bypass:              expandStorageAccountBypass(networkRule),
	}

	if v := networkRule["default_action"]; v != nil {
		defaultAction := armstorage.DefaultAction(v.(string))
		networkRuleSet.DefaultAction = &defaultAction
	}

	return networkRuleSet
}

func expandStorageAccountIPRules(networkRule map[string]interface{}) *[]armstorage.IPRule {
	ipRulesInfo := networkRule["ip_rules"].(*schema.Set).List()
	ipRules := make([]armstorage.IPRule, len(ipRulesInfo))

	for i, ipRuleConfig := range ipRulesInfo {
		attrs := ipRuleConfig.(string)
		ipRule := armstorage.IPRule{
			IPAddressOrRange: utils.String(attrs),
			Action:           utils.String("Allow"),
		}
		ipRules[i] = ipRule
	}

	return &ipRules
}

func expandStorageAccountVirtualNetworks(networkRule map[string]interface{}) *[]armstorage.VirtualNetworkRule {
	virtualNetworkInfo := networkRule["virtual_network_subnet_ids"].(*schema.Set).List()
	virtualNetworks := make([]armstorage.VirtualNetworkRule, len(virtualNetworkInfo))

	for i, virtualNetworkConfig := range virtualNetworkInfo {
		attrs := virtualNetworkConfig.(string)
		virtualNetwork := armstorage.VirtualNetworkRule{
			VirtualNetworkResourceID: utils.String(attrs),
			Action:                   utils.String("Allow"),
		}
		virtualNetworks[i] = virtualNetwork
	}

	return &virtualNetworks
}

func expandStorageAccountBypass(networkRule map[string]interface{}) *armstorage.Bypass {
	bypassInfo := networkRule["bypass"].(*schema.Set).List()
	if len(bypassInfo) == 0 {
		return nil
	}

	var bypassValues []string
	for _, bypassConfig := range bypassInfo {
		bypassValues = append(bypassValues, bypassConfig.(string))
	}
	result := armstorage.Bypass(strings.Join(bypassValues, ", "))
	return &result
}

func expandBlobProperties(input []interface{}) armstorage.BlobServiceProperties {
	props := armstorage.BlobServiceProperties{
		BlobServiceProperties: &armstorage.BlobServicePropertiesAutoGenerated{
			Cors: &armstorage.CorsRules{
				CorsRules: &[]armstorage.CorsRule{},
			},
			DeleteRetentionPolicy: &armstorage.DeleteRetentionPolicy{
				Enabled: utils.Bool(false),
			},
		},
	}

	if len(input) == 0 || input[0] == nil {
		return props
	}

	v := input[0].(map[string]interface{})

	deletePolicyRaw := v["delete_retention_policy"].([]interface{})
	props.BlobServiceProperties.DeleteRetentionPolicy = expandBlobPropertiesDeleteRetentionPolicy(deletePolicyRaw)

	corsRaw := v["cors_rule"].([]interface{})
	props.BlobServiceProperties.Cors = expandBlobPropertiesCors(corsRaw)

	return props
}

func expandBlobPropertiesDeleteRetentionPolicy(input []interface{}) *armstorage.DeleteRetentionPolicy {
	deleteRetentionPolicy := armstorage.DeleteRetentionPolicy{
		Enabled: utils.Bool(false),
	}

	if len(input) == 0 {
		return &deleteRetentionPolicy
	}

	policy := input[0].(map[string]interface{})
	days := policy["days"].(int)
	deleteRetentionPolicy.Enabled = utils.Bool(true)
	deleteRetentionPolicy.Days = utils.Int32(int32(days))

	return &deleteRetentionPolicy
}

func expandBlobPropertiesCors(input []interface{}) *armstorage.CorsRules {
	blobCorsRules := armstorage.CorsRules{}

	if len(input) == 0 {
		return &blobCorsRules
	}

	corsRules := make([]armstorage.CorsRule, 0)
	for _, attr := range input {
		corsRuleAttr := attr.(map[string]interface{})
		corsRule := armstorage.CorsRule{}

		allowedOrigins := *utils.ExpandStringSlice(corsRuleAttr["allowed_origins"].([]interface{}))
		allowedHeaders := *utils.ExpandStringSlice(corsRuleAttr["allowed_headers"].([]interface{}))
		exposedHeaders := *utils.ExpandStringSlice(corsRuleAttr["exposed_headers"].([]interface{}))
		maxAgeInSeconds := int32(corsRuleAttr["max_age_in_seconds"].(int))

		corsRule.AllowedOrigins = &allowedOrigins
		corsRule.AllowedHeaders = &allowedHeaders
		corsRule.AllowedMethods = expandStorageCorsMethods(corsRuleAttr["allowed_methods"].([]interface{}))
		corsRule.ExposedHeaders = &exposedHeaders
		corsRule.MaxAgeInSeconds = &maxAgeInSeconds

		corsRules = append(corsRules, corsRule)
	}

	blobCorsRules.CorsRules = &corsRules

	return &blobCorsRules
}

func expandStorageCorsMethods(input []interface{}) *[]armstorage.CorsRuleAllowedMethodsItem {
	result := make([]armstorage.CorsRuleAllowedMethodsItem, 0)
	for _, item := range input {
		if item != nil {
			result = append(result, armstorage.CorsRuleAllowedMethodsItem(item.(string)))
		} else {
			result = append(result, "")
		}
	}
	return &result
}

func expandQueueProperties(input []interface{}) (queues.StorageServiceProperties, error) {
	var err error
	properties := queues.StorageServiceProperties{
		Cors: &queues.Cors{
			CorsRule: []queues.CorsRule{},
		},
		HourMetrics: &queues.MetricsConfig{
			Enabled: false,
		},
		MinuteMetrics: &queues.MetricsConfig{
			Enabled: false,
		},
	}
	if len(input) == 0 {
		return properties, nil
	}

	attrs := input[0].(map[string]interface{})

	properties.Cors = expandQueuePropertiesCors(attrs["cors_rule"].([]interface{}))
	properties.Logging = expandQueuePropertiesLogging(attrs["logging"].([]interface{}))
	properties.MinuteMetrics, err = expandQueuePropertiesMetrics(attrs["minute_metrics"].([]interface{}))
	if err != nil {
		return properties, fmt.Errorf("Error expanding `minute_metrics`: %+v", err)
	}
	properties.HourMetrics, err = expandQueuePropertiesMetrics(attrs["hour_metrics"].([]interface{}))
	if err != nil {
		return properties, fmt.Errorf("Error expanding `hour_metrics`: %+v", err)
	}

	return properties, nil
}

func expandQueuePropertiesMetrics(input []interface{}) (*queues.MetricsConfig, error) {
	if len(input) == 0 {
		return &queues.MetricsConfig{}, nil
	}

	metricsAttr := input[0].(map[string]interface{})

	metrics := &queues.MetricsConfig{
		Version: metricsAttr["version"].(string),
		Enabled: metricsAttr["enabled"].(bool),
	}

	if v, ok := metricsAttr["retention_policy_days"]; ok {
		if days := v.(int); days > 0 {
			metrics.RetentionPolicy = queues.RetentionPolicy{
				Days:    days,
				Enabled: true,
			}
		}
	}

	if v, ok := metricsAttr["include_apis"]; ok {
		includeAPIs := v.(bool)
		if metrics.Enabled {
			metrics.IncludeAPIs = &includeAPIs
		} else if includeAPIs {
			return nil, fmt.Errorf("`include_apis` may only be set when `enabled` is true")
		}
	}

	return metrics, nil
}

func expandQueuePropertiesLogging(input []interface{}) *queues.LoggingConfig {
	if len(input) == 0 {
		return &queues.LoggingConfig{}
	}

	loggingAttr := input[0].(map[string]interface{})
	logging := &queues.LoggingConfig{
		Version: loggingAttr["version"].(string),
		Delete:  loggingAttr["delete"].(bool),
		Read:    loggingAttr["read"].(bool),
		Write:   loggingAttr["write"].(bool),
	}

	if v, ok := loggingAttr["retention_policy_days"]; ok {
		if days := v.(int); days > 0 {
			logging.RetentionPolicy = queues.RetentionPolicy{
				Days:    days,
				Enabled: true,
			}
		}
	}

	return logging
}

func expandQueuePropertiesCors(input []interface{}) *queues.Cors {
	if len(input) == 0 {
		return &queues.Cors{}
	}

	corsRules := make([]queues.CorsRule, 0)
	for _, attr := range input {
		corsRuleAttr := attr.(map[string]interface{})
		corsRule := queues.CorsRule{}

		corsRule.AllowedOrigins = strings.Join(*utils.ExpandStringSlice(corsRuleAttr["allowed_origins"].([]interface{})), ",")
		corsRule.ExposedHeaders = strings.Join(*utils.ExpandStringSlice(corsRuleAttr["exposed_headers"].([]interface{})), ",")
		corsRule.AllowedHeaders = strings.Join(*utils.ExpandStringSlice(corsRuleAttr["allowed_headers"].([]interface{})), ",")
		corsRule.AllowedMethods = strings.Join(*utils.ExpandStringSlice(corsRuleAttr["allowed_methods"].([]interface{})), ",")
		corsRule.MaxAgeInSeconds = corsRuleAttr["max_age_in_seconds"].(int)

		corsRules = append(corsRules, corsRule)
	}

	cors := &queues.Cors{
		CorsRule: corsRules,
	}
	return cors
}

func expandStaticWebsiteProperties(input []interface{}) accounts.StorageServiceProperties {
	properties := accounts.StorageServiceProperties{
		StaticWebsite: &accounts.StaticWebsite{
			Enabled: false,
		},
	}
	if len(input) == 0 {
		return properties
	}

	properties.StaticWebsite.Enabled = true

	// @tombuildsstuff: this looks weird, doesn't it?
	// Since the presence of this block signifies the website's enabled however all fields within it are optional
	// TF Core returns a nil object when there's no keys defined within the block, rather than an empty map. As
	// such this hack allows us to have a Static Website block with only Enabled configured, without the optional
	// inner properties.
	if val := input[0]; val != nil {
		attr := val.(map[string]interface{})
		if v, ok := attr["index_document"]; ok {
			properties.StaticWebsite.IndexDocument = v.(string)
		}

		if v, ok := attr["error_404_document"]; ok {
			properties.StaticWebsite.ErrorDocument404Path = v.(string)
		}
	}

	return properties
}

func flattenStorageAccountNetworkRules(input *armstorage.NetworkRuleSet) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	networkRules := make(map[string]interface{})

	networkRules["ip_rules"] = schema.NewSet(schema.HashString, flattenStorageAccountIPRules(input.IPRules))
	networkRules["virtual_network_subnet_ids"] = schema.NewSet(schema.HashString, flattenStorageAccountVirtualNetworks(input.VirtualNetworkRules))
	networkRules["bypass"] = schema.NewSet(schema.HashString, flattenStorageAccountBypass(*input.Bypass))
	networkRules["default_action"] = string(*input.DefaultAction)

	return []interface{}{networkRules}
}

func flattenStorageAccountIPRules(input *[]armstorage.IPRule) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	ipRules := make([]interface{}, 0)
	for _, ipRule := range *input {
		if ipRule.IPAddressOrRange == nil {
			continue
		}

		ipRules = append(ipRules, *ipRule.IPAddressOrRange)
	}

	return ipRules
}

func flattenStorageAccountVirtualNetworks(input *[]armstorage.VirtualNetworkRule) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	virtualNetworks := make([]interface{}, 0)
	for _, virtualNetwork := range *input {
		if virtualNetwork.VirtualNetworkResourceID == nil {
			continue
		}

		virtualNetworks = append(virtualNetworks, *virtualNetwork.VirtualNetworkResourceID)
	}

	return virtualNetworks
}

func flattenBlobProperties(input armstorage.BlobServiceProperties) []interface{} {
	if input.BlobServiceProperties == nil {
		return []interface{}{}
	}

	flattenedCorsRules := make([]interface{}, 0)
	if corsRules := input.BlobServiceProperties.Cors; corsRules != nil {
		flattenedCorsRules = flattenBlobPropertiesCorsRule(corsRules)
	}

	flattenedDeletePolicy := make([]interface{}, 0)
	if deletePolicy := input.BlobServiceProperties.DeleteRetentionPolicy; deletePolicy != nil {
		flattenedDeletePolicy = flattenBlobPropertiesDeleteRetentionPolicy(deletePolicy)
	}

	if len(flattenedCorsRules) == 0 && len(flattenedDeletePolicy) == 0 {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"cors_rule":               flattenedCorsRules,
			"delete_retention_policy": flattenedDeletePolicy,
		},
	}
}

func flattenBlobPropertiesCorsRule(input *armstorage.CorsRules) []interface{} {
	corsRules := make([]interface{}, 0)

	if input == nil || input.CorsRules == nil {
		return corsRules
	}

	for _, corsRule := range *input.CorsRules {
		allowedOrigins := make([]string, 0)
		if corsRule.AllowedOrigins != nil {
			allowedOrigins = *corsRule.AllowedOrigins
		}

		allowedMethods := make([]string, 0)
		for _, i := range *corsRule.AllowedMethods {
			if i != "" {
				allowedMethods = append(allowedMethods, string(i))
			}
		}

		allowedHeaders := make([]string, 0)
		if corsRule.AllowedHeaders != nil {
			allowedHeaders = *corsRule.AllowedHeaders
		}

		exposedHeaders := make([]string, 0)
		if corsRule.ExposedHeaders != nil {
			exposedHeaders = *corsRule.ExposedHeaders
		}

		maxAgeInSeconds := 0
		if corsRule.MaxAgeInSeconds != nil {
			maxAgeInSeconds = int(*corsRule.MaxAgeInSeconds)
		}

		corsRules = append(corsRules, map[string]interface{}{
			"allowed_headers":    allowedHeaders,
			"allowed_origins":    allowedOrigins,
			"allowed_methods":    allowedMethods,
			"exposed_headers":    exposedHeaders,
			"max_age_in_seconds": maxAgeInSeconds,
		})
	}

	return corsRules
}

func flattenBlobPropertiesDeleteRetentionPolicy(input *armstorage.DeleteRetentionPolicy) []interface{} {
	deleteRetentionPolicy := make([]interface{}, 0)

	if input == nil {
		return deleteRetentionPolicy
	}

	if enabled := input.Enabled; enabled != nil && *enabled {
		days := 0
		if input.Days != nil {
			days = int(*input.Days)
		}

		deleteRetentionPolicy = append(deleteRetentionPolicy, map[string]interface{}{
			"days": days,
		})
	}

	return deleteRetentionPolicy
}

func flattenQueueProperties(input *queues.StorageServiceProperties) []interface{} {
	if input == nil {
		return []interface{}{}
	}

	queueProperties := make(map[string]interface{})

	if cors := input.Cors; cors != nil {
		if len(cors.CorsRule) > 0 {
			if cors.CorsRule[0].AllowedOrigins != "" {
				queueProperties["cors_rule"] = flattenQueuePropertiesCorsRule(input.Cors.CorsRule)
			}
		}
	}

	if logging := input.Logging; logging != nil {
		if logging.Version != "" {
			queueProperties["logging"] = flattenQueuePropertiesLogging(*logging)
		}
	}

	if hourMetrics := input.HourMetrics; hourMetrics != nil {
		if hourMetrics.Version != "" {
			queueProperties["hour_metrics"] = flattenQueuePropertiesMetrics(*hourMetrics)
		}
	}

	if minuteMetrics := input.MinuteMetrics; minuteMetrics != nil {
		if minuteMetrics.Version != "" {
			queueProperties["minute_metrics"] = flattenQueuePropertiesMetrics(*minuteMetrics)
		}
	}

	if len(queueProperties) == 0 {
		return []interface{}{}
	}
	return []interface{}{queueProperties}
}

func flattenQueuePropertiesMetrics(input queues.MetricsConfig) []interface{} {
	metrics := make(map[string]interface{})

	metrics["version"] = input.Version
	metrics["enabled"] = input.Enabled

	if input.IncludeAPIs != nil {
		metrics["include_apis"] = *input.IncludeAPIs
	}

	if input.RetentionPolicy.Enabled {
		metrics["retention_policy_days"] = input.RetentionPolicy.Days
	}

	return []interface{}{metrics}
}

func flattenQueuePropertiesCorsRule(input []queues.CorsRule) []interface{} {
	corsRules := make([]interface{}, 0)

	for _, corsRule := range input {
		attr := make(map[string]interface{})

		attr["allowed_origins"] = flattenCorsProperty(corsRule.AllowedOrigins)
		attr["allowed_methods"] = flattenCorsProperty(corsRule.AllowedMethods)
		attr["allowed_headers"] = flattenCorsProperty(corsRule.AllowedHeaders)
		attr["exposed_headers"] = flattenCorsProperty(corsRule.ExposedHeaders)
		attr["max_age_in_seconds"] = corsRule.MaxAgeInSeconds

		corsRules = append(corsRules, attr)
	}

	return corsRules
}

func flattenQueuePropertiesLogging(input queues.LoggingConfig) []interface{} {
	logging := make(map[string]interface{})

	logging["version"] = input.Version
	logging["delete"] = input.Delete
	logging["read"] = input.Read
	logging["write"] = input.Write

	if input.RetentionPolicy.Enabled {
		logging["retention_policy_days"] = input.RetentionPolicy.Days
	}

	return []interface{}{logging}
}

func flattenCorsProperty(input string) []interface{} {
	results := make([]interface{}, 0, len(input))

	origins := strings.Split(input, ",")
	for _, origin := range origins {
		results = append(results, origin)
	}

	return results
}

func flattenStaticWebsiteProperties(input accounts.GetServicePropertiesResult) []interface{} {
	if storageServiceProps := input.StorageServiceProperties; storageServiceProps != nil {
		if staticWebsite := storageServiceProps.StaticWebsite; staticWebsite != nil {
			if !staticWebsite.Enabled {
				return []interface{}{}
			}

			return []interface{}{
				map[string]interface{}{
					"index_document":     staticWebsite.IndexDocument,
					"error_404_document": staticWebsite.ErrorDocument404Path,
				},
			}
		}
	}
	return []interface{}{}
}

func flattenStorageAccountBypass(input armstorage.Bypass) []interface{} {
	bypassValues := strings.Split(string(input), ", ")
	bypass := make([]interface{}, len(bypassValues))

	for i, value := range bypassValues {
		bypass[i] = value
	}

	return bypass
}

func ValidateStorageAccountName(v interface{}, _ string) (warnings []string, errors []error) {
	input := v.(string)

	if !regexp.MustCompile(`\A([a-z0-9]{3,24})\z`).MatchString(input) {
		errors = append(errors, fmt.Errorf("name (%q) can only consist of lowercase letters and numbers, and must be between 3 and 24 characters long", input))
	}

	return warnings, errors
}

func expandAzureRmStorageAccountIdentity(d *schema.ResourceData) *armstorage.IDentity {
	identities := d.Get("identity").([]interface{})
	identity := identities[0].(map[string]interface{})
	identityType := identity["type"].(string)
	return &armstorage.IDentity{
		Type: &identityType,
	}
}

func flattenAzureRmStorageAccountIdentity(identity *armstorage.IDentity) []interface{} {
	if identity == nil {
		return make([]interface{}, 0)
	}

	result := make(map[string]interface{})
	if identity.Type != nil {
		result["type"] = *identity.Type
	}
	if identity.PrincipalID != nil {
		result["principal_id"] = *identity.PrincipalID
	}
	if identity.TenantID != nil {
		result["tenant_id"] = *identity.TenantID
	}

	return []interface{}{result}
}

func getBlobConnectionString(blobEndpoint *string, acctName *string, acctKey *string) string {
	var endpoint string
	if blobEndpoint != nil {
		endpoint = *blobEndpoint
	}

	var name string
	if acctName != nil {
		name = *acctName
	}

	var key string
	if acctKey != nil {
		key = *acctKey
	}

	return fmt.Sprintf("DefaultEndpointsProtocol=https;BlobEndpoint=%s;AccountName=%s;AccountKey=%s", endpoint, name, key)
}

func flattenAndSetAzureRmStorageAccountPrimaryEndpoints(d *schema.ResourceData, primary *armstorage.Endpoints) error {
	if primary == nil {
		return fmt.Errorf("primary endpoints should not be empty")
	}

	if err := setEndpointAndHost(d, "primary", primary.Blob, "blob"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "primary", primary.Dfs, "dfs"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "primary", primary.File, "file"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "primary", primary.Queue, "queue"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "primary", primary.Table, "table"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "primary", primary.Web, "web"); err != nil {
		return err
	}

	return nil
}

func flattenAndSetAzureRmStorageAccountSecondaryEndpoints(d *schema.ResourceData, secondary *armstorage.Endpoints) error {
	if secondary == nil {
		return nil
	}

	if err := setEndpointAndHost(d, "secondary", secondary.Blob, "blob"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "secondary", secondary.Dfs, "dfs"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "secondary", secondary.File, "file"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "secondary", secondary.Queue, "queue"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "secondary", secondary.Table, "table"); err != nil {
		return err
	}
	if err := setEndpointAndHost(d, "secondary", secondary.Web, "web"); err != nil {
		return err
	}
	return nil
}

func setEndpointAndHost(d *schema.ResourceData, ordinalString string, endpointType *string, typeString string) error {
	var endpoint, host string
	if v := endpointType; v != nil {
		endpoint = *v

		u, err := url.Parse(*v)
		if err != nil {
			return fmt.Errorf("invalid %s endpoint for parsing: %q", typeString, *v)
		}
		host = u.Host
	}

	// lintignore: R001
	d.Set(fmt.Sprintf("%s_%s_endpoint", ordinalString, typeString), endpoint)
	// lintignore: R001
	d.Set(fmt.Sprintf("%s_%s_host", ordinalString, typeString), host)
	return nil
}
