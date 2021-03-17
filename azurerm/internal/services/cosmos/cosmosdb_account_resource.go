package cosmos

import (
	"bytes"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-01-15/documentdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/cosmos/common"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/cosmos/parse"
	keyVaultParse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/keyvault/parse"
	keyVaultValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/keyvault/validate"
	msiparse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/msi/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/msi/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// If the consistency policy of the Cosmos DB Database Account is not bounded staleness,
// any changes to the configuration for bounded staleness should be suppressed.
func suppressConsistencyPolicyStalenessConfiguration(_, _, _ string, d *schema.ResourceData) bool {
	consistencyPolicyList := d.Get("consistency_policy").([]interface{})
	if len(consistencyPolicyList) == 0 || consistencyPolicyList[0] == nil {
		return false
	}

	consistencyPolicy := consistencyPolicyList[0].(map[string]interface{})

	return consistencyPolicy["consistency_level"].(string) != string(documentdb.BoundedStaleness)
}

func resourceCosmosDbAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceCosmosDbAccountCreate,
		Read:   resourceCosmosDbAccountRead,
		Update: resourceCosmosDbAccountUpdate,
		Delete: resourceCosmosDbAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(180 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(180 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-z0-9]{3,50}$"),
					"Cosmos DB Account name must be 3 - 50 characters long, contain only lowercase letters, numbers and hyphens.",
				),
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			// resource fields
			"offer_type": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(documentdb.Standard),
				}, true),
			},

			"kind": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          string(documentdb.GlobalDocumentDB),
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc: validation.StringInSlice([]string{
					string(documentdb.GlobalDocumentDB),
					string(documentdb.MongoDB),
					string(documentdb.Parse),
				}, true),
			},

			"ip_range_filter": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^(\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(/([1-2][0-9]|3[0-2]))?\b[,]?)*$`),
					"Cosmos DB ip_range_filter must be a set of CIDR IP addresses separated by commas with no spaces: '10.0.0.1,10.0.0.2,10.20.0.0/16'",
				),
			},

			"enable_free_tier": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"analytical_storage_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"public_network_access_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"enable_automatic_failover": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"key_vault_key_id": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressIgnoreKeyVaultKeyVersion,
				ValidateFunc:     keyVaultValidate.VersionlessNestedItemId,
			},

			"consistency_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"consistency_level": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								string(documentdb.BoundedStaleness),
								string(documentdb.ConsistentPrefix),
								string(documentdb.Eventual),
								string(documentdb.Session),
								string(documentdb.Strong),
							}, true),
						},

						"max_interval_in_seconds": {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: suppressConsistencyPolicyStalenessConfiguration,
							ValidateFunc:     validation.IntBetween(5, 86400), // single region values
						},

						"max_staleness_prefix": {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: suppressConsistencyPolicyStalenessConfiguration,
							ValidateFunc:     validation.IntBetween(10, 1000000), // single region values
						},
					},
				},
			},

			"geo_location": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringMatch(
								regexp.MustCompile("^[-a-z0-9]{3,50}$"),
								"Cosmos DB location prefix (ID) must be 3 - 50 characters long, contain only lowercase letters, numbers and hyphens.",
							),
							Deprecated: "This is deprecated because the service no longer accepts this as an input since Apr 25, 2019",
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"location": location.SchemaWithoutForceNew(),

						"failover_priority": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"zone_redundant": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				Set: resourceAzureRMCosmosDBAccountGeoLocationHash,
			},

			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc: validation.StringInSlice([]string{
								"EnableAggregationPipeline",
								"EnableCassandra",
								"EnableGremlin",
								"EnableTable",
								"EnableServerless",
								"EnableMongo",
								"MongoDBv3.4",
								"mongoEnableDocLevelTTL",
								"DisableRateLimitingResponses",
								"AllowSelfServeUpgradeToMongo36",
								// TODO: Remove in 3.0 - doesn't do anything
								"EnableAnalyticalStorage",
							}, true),
						},
					},
				},
				Set: resourceAzureRMCosmosDBAccountCapabilitiesHash,
			},

			"is_virtual_network_filter_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"virtual_network_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"ignore_missing_vnet_service_endpoint": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				Set: resourceAzureRMCosmosDBAccountVirtualNetworkRuleHash,
			},

			"enable_multiple_write_locations": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"identity": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(documentdb.ResourceIdentityTypeSystemAssigned),
								string(documentdb.ResourceIdentityTypeSystemAssignedUserAssigned),
								string(documentdb.ResourceIdentityTypeUserAssigned),
							}, false),
						},

						"principal_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"identity_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validate.UserAssignedIdentityID,
							},
						},
					},
				},
			},

			"backup_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  string(documentdb.TypePeriodic),
							ValidateFunc: validation.StringInSlice([]string{
								string(documentdb.TypeContinuous),
								string(documentdb.TypePeriodic),
							}, false),
						},

						"interval_in_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      240,
							ValidateFunc: validation.IntBetween(60, 1440),
						},

						"retention_in_hours": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							ValidateFunc: validation.IntBetween(8, 720),
						},
					},
				},
			},

			"connector_offer": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(documentdb.Small),
				}, false),
			},

			"cors_rule": common.SchemaCorsRule(true),

			"enable_key_based_meta_write_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"enable_cassandra_connector": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"mongo_server_version": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(documentdb.ThreeFullStopTwo),
					string(documentdb.ThreeFullStopSix),
					string(documentdb.FourFullStopZero),
				}, false),
			},

			"network_acl_bypass": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(documentdb.NetworkACLBypassNone),
				ValidateFunc: validation.StringInSlice([]string{
					string(documentdb.NetworkACLBypassNone),
					string(documentdb.NetworkACLBypassAzureServices),
				}, false),
			},

			"network_acl_bypass_ids": {
				Type:     schema.TypeList,
				Optional: true, Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: azure.ValidateResourceID,
				},
			},

			// computed
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"read_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"write_endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"primary_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_readonly_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"secondary_readonly_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"primary_master_key": {
				Type:       schema.TypeString,
				Computed:   true,
				Sensitive:  true,
				Deprecated: "This property has been renamed to `primary_key` and will be removed in v3.0 of the provider in support of HashiCorp's inclusive language policy which can be found here: https://discuss.hashicorp.com/t/inclusive-language-changes",
			},

			"secondary_master_key": {
				Type:       schema.TypeString,
				Computed:   true,
				Sensitive:  true,
				Deprecated: "This property has been renamed to `secondary_key` and will be removed in v3.0 of the provider in support of HashiCorp's inclusive language policy which can be found here: https://discuss.hashicorp.com/t/inclusive-language-changes",
			},

			"primary_readonly_master_key": {
				Type:       schema.TypeString,
				Computed:   true,
				Sensitive:  true,
				Deprecated: "This property has been renamed to `primary_readonly_key` and will be removed in v3.0 of the provider in support of HashiCorp's inclusive language policy which can be found here: https://discuss.hashicorp.com/t/inclusive-language-changes",
			},

			"secondary_readonly_master_key": {
				Type:       schema.TypeString,
				Computed:   true,
				Sensitive:  true,
				Deprecated: "This property has been renamed to `secondary_readonly_key` and will be removed in v3.0 of the provider in support of HashiCorp's inclusive language policy which can be found here: https://discuss.hashicorp.com/t/inclusive-language-changes",
			},

			"connection_strings": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceCosmosDbAccountCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Cosmos.DatabaseClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Cosmos DB Account creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewDatabaseAccountID(subscriptionId, resourceGroup, name).ID()

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing CosmosDB Account %q (Resource Group %q): %s", name, resourceGroup, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_cosmosdb_account", id)
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})
	kind := d.Get("kind").(string)
	offerType := d.Get("offer_type").(string)
	ipRangeFilter := d.Get("ip_range_filter").(string)
	isVirtualNetworkFilterEnabled := d.Get("is_virtual_network_filter_enabled").(bool)
	enableFreeTier := d.Get("enable_free_tier").(bool)
	enableAutomaticFailover := d.Get("enable_automatic_failover").(bool)
	enableMultipleWriteLocations := d.Get("enable_multiple_write_locations").(bool)
	enableAnalyticalStorage := d.Get("analytical_storage_enabled").(bool)

	r, err := client.CheckNameExists(ctx, name)
	if err != nil {
		// todo remove when https://github.com/Azure/azure-sdk-for-go/issues/9891 is fixed
		if !utils.ResponseWasStatusCode(r, http.StatusInternalServerError) {
			return fmt.Errorf("checking if CosmosDB Account %q already exists (Resource Group %q): %+v", name, resourceGroup, err)
		}
	} else {
		if !utils.ResponseWasNotFound(r) {
			return fmt.Errorf("CosmosDB Account %s already exists, please import the resource via terraform import", name)
		}
	}
	geoLocations, err := expandAzureRmCosmosDBAccountGeoLocations(d)
	if err != nil {
		return fmt.Errorf("expanding CosmosDB Account %q (Resource Group %q) geo locations: %+v", name, resourceGroup, err)
	}

	publicNetworkAccess := documentdb.Enabled
	if enabled := d.Get("public_network_access_enabled").(bool); !enabled {
		publicNetworkAccess = documentdb.Disabled
	}

	account := documentdb.DatabaseAccountCreateUpdateParameters{
		Location: utils.String(location),
		Kind:     documentdb.DatabaseAccountKind(kind),
		DatabaseAccountCreateUpdateProperties: &documentdb.DatabaseAccountCreateUpdateProperties{
			DatabaseAccountOfferType:           utils.String(offerType),
			IPRules:                            common.CosmosDBIpRangeFilterToIpRules(ipRangeFilter),
			IsVirtualNetworkFilterEnabled:      utils.Bool(isVirtualNetworkFilterEnabled),
			EnableFreeTier:                     utils.Bool(enableFreeTier),
			EnableAutomaticFailover:            utils.Bool(enableAutomaticFailover),
			ConsistencyPolicy:                  expandAzureRmCosmosDBAccountConsistencyPolicy(d),
			Locations:                          &geoLocations,
			Capabilities:                       expandAzureRmCosmosDBAccountCapabilities(d),
			VirtualNetworkRules:                expandAzureRmCosmosDBAccountVirtualNetworkRules(d),
			EnableMultipleWriteLocations:       utils.Bool(enableMultipleWriteLocations),
			PublicNetworkAccess:                publicNetworkAccess,
			EnableAnalyticalStorage:            utils.Bool(enableAnalyticalStorage),
			ConnectorOffer:                     documentdb.ConnectorOffer(d.Get("connector_offer").(string)),
			Cors:                               common.ExpandCosmosCorsRule(d.Get("cors_rule").([]interface{})),
			DisableKeyBasedMetadataWriteAccess: utils.Bool(!d.Get("enable_key_based_meta_write_access").(bool)),
			EnableCassandraConnector:           utils.Bool(d.Get("enable_cassandra_connector").(bool)),
			NetworkACLBypass:                   documentdb.NetworkACLBypass(d.Get("network_acl_bypass").(string)),
			NetworkACLBypassResourceIds:        utils.ExpandStringSlice(d.Get("network_acl_bypass_ids").([]interface{})),
		},
		Tags: tags.Expand(t),
	}

	//mongo db default to v3.6
	if kind == string(documentdb.MongoDB) {
		account.DatabaseAccountCreateUpdateProperties.APIProperties = &documentdb.APIProperties{
			ServerVersion: documentdb.ThreeFullStopSix,
		}
		if v, ok := d.GetOk("mongo_server_version"); ok {
			account.DatabaseAccountCreateUpdateProperties.APIProperties.ServerVersion = documentdb.ServerVersion(v.(string))
		}
	}

	if v, ok := d.GetOk("backup_policy"); ok {
		policy, err := expandCosmosdbAccountBackupPolicy(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("expanding `backup_policy`: %+v", err)
		}
		account.DatabaseAccountCreateUpdateProperties.BackupPolicy = policy
	}

	if v, ok := d.GetOk("identity"); ok {
		identity, err := expandCosmosdbAccountIdentity(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("expanding `identity`: %+v", err)
		}
		account.Identity = identity
	}

	if keyVaultKeyIDRaw, ok := d.GetOk("key_vault_key_id"); ok {
		keyVaultKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(keyVaultKeyIDRaw.(string))
		if err != nil {
			return fmt.Errorf("could not parse Key Vault Key ID: %+v", err)
		}
		account.DatabaseAccountCreateUpdateProperties.KeyVaultKeyURI = utils.String(keyVaultKey.ID())
	}

	// additional validation on MaxStalenessPrefix as it varies depending on if the DB is multi region or not
	consistencyPolicy := account.DatabaseAccountCreateUpdateProperties.ConsistencyPolicy
	if len(geoLocations) > 1 && consistencyPolicy != nil && consistencyPolicy.DefaultConsistencyLevel == documentdb.BoundedStaleness {
		if msp := consistencyPolicy.MaxStalenessPrefix; msp != nil && *msp < 100000 {
			return fmt.Errorf("max_staleness_prefix (%d) must be greater then 100000 when more then one geo_location is used", *msp)
		}
		if mis := consistencyPolicy.MaxIntervalInSeconds; mis != nil && *mis < 300 {
			return fmt.Errorf("max_interval_in_seconds (%d) must be greater then 300 (5min) when more then one geo_location is used", *mis)
		}
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, account)
	if err != nil {
		return fmt.Errorf("creating CosmosDB Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of the CosmosDB Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id)

	return resourceCosmosDbAccountRead(d, meta)
}

func resourceCosmosDbAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.DatabaseClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Cosmos DB Account update.")

	id, err := parse.DatabaseAccountID(d.Id())
	if err != nil {
		return err
	}

	account := documentdb.DatabaseAccountUpdateParameters{
		DatabaseAccountUpdateProperties: &documentdb.DatabaseAccountUpdateProperties{},
	}

	if d.HasChange("ip_range_filter") {
		account.DatabaseAccountUpdateProperties.IPRules = common.CosmosDBIpRangeFilterToIpRules(d.Get("ip_range_filter").(string))
	}

	if d.HasChange("enable_free_tier") {
		account.DatabaseAccountUpdateProperties.EnableFreeTier = utils.Bool(d.Get("enable_free_tier").(bool))
	}

	if d.HasChange("analytical_storage_enabled") {
		account.DatabaseAccountUpdateProperties.EnableAnalyticalStorage = utils.Bool(d.Get("analytical_storage_enabled").(bool))
	}

	if d.HasChange("public_network_access_enabled") {
		publicNetworkAccess := documentdb.Enabled
		if !d.Get("public_network_access_enabled").(bool) {
			publicNetworkAccess = documentdb.Disabled
		}
		account.DatabaseAccountUpdateProperties.PublicNetworkAccess = publicNetworkAccess
	}

	if d.HasChange("enable_automatic_failover") {
		account.DatabaseAccountUpdateProperties.EnableAutomaticFailover = utils.Bool(d.Get("enable_automatic_failover").(bool))
	}

	if keyVaultKeyIDRaw, ok := d.GetOk("key_vault_key_id"); ok {
		keyVaultKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(keyVaultKeyIDRaw.(string))
		if err != nil {
			return fmt.Errorf("could not parse Key Vault Key ID: %+v", err)
		}
		account.DatabaseAccountUpdateProperties.KeyVaultKeyURI = utils.String(keyVaultKey.ID())
	}

	if d.HasChange("consistency_policy") {
		account.DatabaseAccountUpdateProperties.ConsistencyPolicy = expandAzureRmCosmosDBAccountConsistencyPolicy(d)
	}

	if d.HasChange("geo_location") {
		geoLocations, err := expandAzureRmCosmosDBAccountGeoLocations(d)
		if err != nil {
			return fmt.Errorf("expanding CosmosDB Account %q (Resource Group %q) geo locations: %+v", id.Name, id.ResourceGroup, err)
		}
		account.DatabaseAccountUpdateProperties.Locations = &geoLocations
	}

	if d.HasChange("capabilities") {
		account.DatabaseAccountUpdateProperties.Capabilities = expandAzureRmCosmosDBAccountCapabilities(d)
	}

	if d.HasChange("is_virtual_network_filter_enabled") {
		account.DatabaseAccountUpdateProperties.IsVirtualNetworkFilterEnabled = utils.Bool(d.Get("is_virtual_network_filter_enabled").(bool))
	}

	if d.HasChange("virtual_network_rule") {
		account.DatabaseAccountUpdateProperties.VirtualNetworkRules = expandAzureRmCosmosDBAccountVirtualNetworkRules(d)
	}

	if d.HasChange("enable_multiple_write_locations") {
		account.DatabaseAccountUpdateProperties.EnableMultipleWriteLocations = utils.Bool(d.Get("enable_multiple_write_locations").(bool))
	}

	if d.HasChange("identity") {
		identity, err := expandCosmosdbAccountIdentity(d.Get("identity").([]interface{}))
		if err != nil {
			return fmt.Errorf("expanding `identity`: %+v", err)
		}
		account.Identity = identity
	}

	if d.HasChange("backup_policy") {
		policy, err := expandCosmosdbAccountBackupPolicy(d.Get("backup_policy").([]interface{}))
		if err != nil {
			return fmt.Errorf("expanding `backup_policy`%+v", err)
		}
		account.DatabaseAccountUpdateProperties.BackupPolicy = policy
	}

	if d.HasChange("connector_offer") {
		account.DatabaseAccountUpdateProperties.ConnectorOffer = documentdb.ConnectorOffer(d.Get("connector_offer").(string))
	}

	if d.HasChange("cors_rule") {
		account.DatabaseAccountUpdateProperties.Cors = common.ExpandCosmosCorsRule(d.Get("cors_rule").([]interface{}))
	}

	if d.HasChange("enable_key_based_meta_write_access") {
		account.DatabaseAccountUpdateProperties.DisableKeyBasedMetadataWriteAccess = utils.Bool(!d.Get("enable_key_based_meta_write_access").(bool))
	}

	if d.HasChange("enable_cassandra_connector") {
		account.DatabaseAccountUpdateProperties.EnableCassandraConnector = utils.Bool(d.Get("enable_cassandra_connector").(bool))
	}

	if d.HasChange("mongo_server_version") {
		account.DatabaseAccountUpdateProperties.APIProperties = &documentdb.APIProperties{
			ServerVersion: documentdb.ServerVersion(d.Get("mongo_server_version").(string)),
		}
	}

	if d.HasChange("network_acl_bypass") {
		account.DatabaseAccountUpdateProperties.NetworkACLBypass = documentdb.NetworkACLBypass(d.Get("network_acl_bypass").(string))
	}

	if d.HasChange("network_acl_bypass_ids") {
		account.DatabaseAccountUpdateProperties.NetworkACLBypassResourceIds = utils.ExpandStringSlice(d.Get("network_acl_bypass_ids").([]interface{}))
	}

	if d.HasChange("tags") {
		account.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, account)
	if err != nil {
		return fmt.Errorf("updating CosmosDB Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of the CosmosDB Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return resourceCosmosDbAccountRead(d, meta)
}

func resourceCosmosDbAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.DatabaseClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DatabaseAccountID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving CosmosDB Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	d.Set("kind", string(resp.Kind))

	if v := resp.Identity; v != nil {
		identity, err := flattenAzureRmdocumentdbMachineIdentity(v)
		if err != nil {
			return err
		}
		if err := d.Set("identity", identity); err != nil {
			return fmt.Errorf("setting `identity`: %+v", err)
		}
	}

	if props := resp.DatabaseAccountGetProperties; props != nil {
		d.Set("offer_type", string(props.DatabaseAccountOfferType))
		d.Set("ip_range_filter", common.CosmosDBIpRulesToIpRangeFilter(props.IPRules))
		d.Set("endpoint", props.DocumentEndpoint)

		d.Set("enable_free_tier", props.EnableFreeTier)
		d.Set("analytical_storage_enabled", props.EnableAnalyticalStorage)
		d.Set("public_network_access_enabled", props.PublicNetworkAccess == documentdb.Enabled)

		if v := resp.IsVirtualNetworkFilterEnabled; v != nil {
			d.Set("is_virtual_network_filter_enabled", props.IsVirtualNetworkFilterEnabled)
		}

		if v := resp.EnableAutomaticFailover; v != nil {
			d.Set("enable_automatic_failover", props.EnableAutomaticFailover)
		}

		if v := resp.KeyVaultKeyURI; v != nil {
			d.Set("key_vault_key_id", props.KeyVaultKeyURI)
		}

		if v := resp.EnableMultipleWriteLocations; v != nil {
			d.Set("enable_multiple_write_locations", props.EnableMultipleWriteLocations)
		}

		if err = d.Set("consistency_policy", flattenAzureRmCosmosDBAccountConsistencyPolicy(props.ConsistencyPolicy)); err != nil {
			return fmt.Errorf("setting CosmosDB Account %q `consistency_policy` (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}

		if err = d.Set("geo_location", flattenAzureRmCosmosDBAccountGeoLocations(props)); err != nil {
			return fmt.Errorf("setting `geo_location`: %+v", err)
		}

		if err = d.Set("capabilities", flattenAzureRmCosmosDBAccountCapabilities(props.Capabilities)); err != nil {
			return fmt.Errorf("setting `capabilities`: %+v", err)
		}

		if err = d.Set("virtual_network_rule", flattenAzureRmCosmosDBAccountVirtualNetworkRules(props.VirtualNetworkRules)); err != nil {
			return fmt.Errorf("setting `virtual_network_rule`: %+v", err)
		}

		policy, err := flattenCosmosdbAccountBackupPolicy(props.BackupPolicy)
		if err != nil {
			return err
		}
		if err = d.Set("backup_policy", policy); err != nil {
			return fmt.Errorf("setting `backup_policy`: %+v", err)
		}

		d.Set("connector_offer", props.ConnectorOffer)
		d.Set("cors_rule", common.FlattenCosmosCorsRule(props.Cors))
		d.Set("enable_key_based_meta_write_access", !*props.DisableKeyBasedMetadataWriteAccess)
		d.Set("enable_cassandra_connector", props.EnableCassandraConnector)
		if apiProps := props.APIProperties; apiProps != nil {
			d.Set("mongo_server_version", apiProps.ServerVersion)
		}
		d.Set("network_acl_bypass", props.NetworkACLBypass)
		d.Set("network_acl_bypass_ids", utils.FlattenStringSlice(props.NetworkACLBypassResourceIds))
	}

	readEndpoints := make([]string, 0)
	if p := resp.ReadLocations; p != nil {
		for _, l := range *p {
			if l.DocumentEndpoint == nil {
				continue
			}

			readEndpoints = append(readEndpoints, *l.DocumentEndpoint)
		}
	}
	if err := d.Set("read_endpoints", readEndpoints); err != nil {
		return fmt.Errorf("Error setting `read_endpoints`: %s", err)
	}

	writeEndpoints := make([]string, 0)
	if p := resp.WriteLocations; p != nil {
		for _, l := range *p {
			if l.DocumentEndpoint == nil {
				continue
			}

			writeEndpoints = append(writeEndpoints, *l.DocumentEndpoint)
		}
	}
	if err := d.Set("write_endpoints", writeEndpoints); err != nil {
		return fmt.Errorf("Error setting `write_endpoints`: %s", err)
	}

	// ListKeys returns a data structure containing a DatabaseAccountListReadOnlyKeysResult pointer
	// implying that it also returns the read only keys, however this appears to not be the case
	keys, err := client.ListKeys(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(keys.Response) {
			log.Printf("[DEBUG] Keys were not found for CosmosDB Account %q (Resource Group %q) - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("[ERROR] Unable to List Write keys for CosmosDB Account %s: %s", id.Name, err)
	}
	d.Set("primary_key", keys.PrimaryMasterKey)
	d.Set("secondary_key", keys.SecondaryMasterKey)
	d.Set("primary_master_key", keys.PrimaryMasterKey)
	d.Set("secondary_master_key", keys.SecondaryMasterKey)

	readonlyKeys, err := client.ListReadOnlyKeys(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(keys.Response) {
			log.Printf("[DEBUG] Read Only Keys were not found for CosmosDB Account %q (Resource Group %q) - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("[ERROR] Unable to List read-only keys for CosmosDB Account %s: %s", id.Name, err)
	}
	d.Set("primary_readonly_key", readonlyKeys.PrimaryReadonlyMasterKey)
	d.Set("secondary_readonly_key", readonlyKeys.SecondaryReadonlyMasterKey)
	d.Set("primary_readonly_master_key", readonlyKeys.PrimaryReadonlyMasterKey)
	d.Set("secondary_readonly_master_key", readonlyKeys.SecondaryReadonlyMasterKey)

	connStringResp, err := client.ListConnectionStrings(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(keys.Response) {
			log.Printf("[DEBUG] Connection Strings were not found for CosmosDB Account %q (Resource Group %q) - removing from state!", id.Name, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("[ERROR] Unable to List connection strings for CosmosDB Account %s: %s", id.Name, err)
	}

	var connStrings []string
	if connStringResp.ConnectionStrings != nil {
		connStrings = make([]string, len(*connStringResp.ConnectionStrings))
		for i, v := range *connStringResp.ConnectionStrings {
			connStrings[i] = *v.ConnectionString
		}
	}
	d.Set("connection_strings", connStrings)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceCosmosDbAccountDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.DatabaseClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.DatabaseAccountID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting CosmosDB Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of the CosmosDB Account %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

//func resourceCosmosDbAccountApiUpsert(client *documentdb.DatabaseAccountsClient, ctx context.Context, resourceGroup string, name string, account documentdb.DatabaseAccountCreateUpdateParameters, d *schema.ResourceData) (*documentdb.DatabaseAccountGetResults, error) {
//	future, err := client.CreateOrUpdate(ctx, resourceGroup, name, account)
//	if err != nil {
//		return nil, fmt.Errorf("Error creating/updating CosmosDB Account %q (Resource Group %q): %+v", name, resourceGroup, err)
//	}
//
//	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
//		return nil, fmt.Errorf("Error waiting for the CosmosDB Account %q (Resource Group %q) to finish creating/updating: %+v", name, resourceGroup, err)
//	}
//
//	// if a replication location is added or removed it can take some time to provision
//	stateConf := &resource.StateChangeConf{
//		Pending:    []string{"Creating", "Updating", "Deleting", "Initializing"},
//		Target:     []string{"Succeeded"},
//		MinTimeout: 30 * time.Second,
//		Delay:      30 * time.Second, // required because it takes some time before the 'creating' location shows up
//		Refresh: func() (interface{}, string, error) {
//			resp, err2 := client.Get(ctx, resourceGroup, name)
//			if err2 != nil || resp.StatusCode == http.StatusNotFound {
//				return nil, "", fmt.Errorf("Error reading CosmosDB Account %q after create/update (Resource Group %q): %+v", name, resourceGroup, err2)
//			}
//			status := "Succeeded"
//			if props := resp.DatabaseAccountGetProperties; props != nil {
//				locations := append(*props.ReadLocations, *props.WriteLocations...)
//				for _, l := range locations {
//					if status = *l.ProvisioningState; status == "Creating" || status == "Updating" || status == "Deleting" {
//						break // return the first non successful status.
//					}
//				}
//
//				for _, desiredLocation := range *account.Locations {
//					for index, l := range locations {
//						if azure.NormalizeLocation(*desiredLocation.LocationName) == azure.NormalizeLocation(*l.LocationName) {
//							break
//						}
//
//						if (index + 1) == len(locations) {
//							return resp, "Updating", nil
//						}
//					}
//				}
//			}
//
//			return resp, status, nil
//		},
//	}
//
//	if d.IsNewResource() {
//		stateConf.Timeout = d.Timeout(schema.TimeoutCreate)
//	} else {
//		stateConf.Timeout = d.Timeout(schema.TimeoutUpdate)
//	}
//
//	resp, err := stateConf.WaitForState()
//	if err != nil {
//		return nil, fmt.Errorf("Error waiting for the CosmosDB Account %q (Resource Group %q) to provision: %+v", name, resourceGroup, err)
//	}
//
//	r := resp.(documentdb.DatabaseAccountGetResults)
//	return &r, nil
//}

func expandAzureRmCosmosDBAccountConsistencyPolicy(d *schema.ResourceData) *documentdb.ConsistencyPolicy {
	i := d.Get("consistency_policy").([]interface{})
	if len(i) == 0 || i[0] == nil {
		return nil
	}
	input := i[0].(map[string]interface{})

	consistencyLevel := input["consistency_level"].(string)
	policy := documentdb.ConsistencyPolicy{
		DefaultConsistencyLevel: documentdb.DefaultConsistencyLevel(consistencyLevel),
	}

	if stalenessPrefix, ok := input["max_staleness_prefix"].(int); ok {
		if stalenessPrefix == 0 {
			stalenessPrefix = 100
		}
		policy.MaxStalenessPrefix = utils.Int64(int64(stalenessPrefix))
	}
	if maxInterval, ok := input["max_interval_in_seconds"].(int); ok {
		if maxInterval == 0 {
			maxInterval = 5
		}
		policy.MaxIntervalInSeconds = utils.Int32(int32(maxInterval))
	}

	return &policy
}

func expandAzureRmCosmosDBAccountGeoLocations(d *schema.ResourceData) ([]documentdb.Location, error) {
	locations := make([]documentdb.Location, 0)
	for _, l := range d.Get("geo_location").(*schema.Set).List() {
		data := l.(map[string]interface{})

		location := documentdb.Location{
			LocationName:     utils.String(azure.NormalizeLocation(data["location"].(string))),
			FailoverPriority: utils.Int32(int32(data["failover_priority"].(int))),
			IsZoneRedundant:  utils.Bool(data["zone_redundant"].(bool)),
		}

		locations = append(locations, location)
	}

	// TODO maybe this should be in a CustomizeDiff
	// all priorities & locations must be unique
	byPriorities := make(map[int]interface{}, len(locations))
	byName := make(map[string]interface{}, len(locations))
	for _, location := range locations {
		priority := int(*location.FailoverPriority)
		name := *location.LocationName

		if _, ok := byPriorities[priority]; ok {
			return nil, fmt.Errorf("Each `geo_location` needs to have a unique failover_prioroty. Multiple instances of '%d' found", priority)
		}

		if _, ok := byName[name]; ok {
			return nil, fmt.Errorf("Each `geo_location` needs to be in unique location. Multiple instances of '%s' found", name)
		}

		byPriorities[priority] = location
		byName[name] = location
	}

	// and must have one of 0 priority
	if _, ok := byPriorities[0]; !ok {
		return nil, fmt.Errorf("There needs to be a `geo_location` with a failover_priority of 0")
	}

	return locations, nil
}

func expandAzureRmCosmosDBAccountCapabilities(d *schema.ResourceData) *[]documentdb.Capability {
	capabilities := d.Get("capabilities").(*schema.Set).List()
	s := make([]documentdb.Capability, 0)

	for _, c := range capabilities {
		m := c.(map[string]interface{})
		s = append(s, documentdb.Capability{Name: utils.String(m["name"].(string))})
	}

	return &s
}

func expandAzureRmCosmosDBAccountVirtualNetworkRules(d *schema.ResourceData) *[]documentdb.VirtualNetworkRule {
	virtualNetworkRules := d.Get("virtual_network_rule").(*schema.Set).List()

	s := make([]documentdb.VirtualNetworkRule, len(virtualNetworkRules))
	for i, r := range virtualNetworkRules {
		m := r.(map[string]interface{})
		s[i] = documentdb.VirtualNetworkRule{
			ID:                               utils.String(m["id"].(string)),
			IgnoreMissingVNetServiceEndpoint: utils.Bool(m["ignore_missing_vnet_service_endpoint"].(bool)),
		}
	}
	return &s
}

func flattenAzureRmCosmosDBAccountConsistencyPolicy(policy *documentdb.ConsistencyPolicy) []interface{} {
	result := map[string]interface{}{}
	result["consistency_level"] = string(policy.DefaultConsistencyLevel)
	if policy.MaxIntervalInSeconds != nil {
		result["max_interval_in_seconds"] = int(*policy.MaxIntervalInSeconds)
	}
	if policy.MaxStalenessPrefix != nil {
		result["max_staleness_prefix"] = int(*policy.MaxStalenessPrefix)
	}

	return []interface{}{result}
}

func flattenAzureRmCosmosDBAccountGeoLocations(account *documentdb.DatabaseAccountGetProperties) *schema.Set {
	locationSet := schema.Set{
		F: resourceAzureRMCosmosDBAccountGeoLocationHash,
	}
	if account == nil {
		return &locationSet
	}

	for _, l := range *account.FailoverPolicies {
		id := *l.ID
		lb := map[string]interface{}{
			"id":                id,
			"location":          azure.NormalizeLocation(*l.LocationName),
			"failover_priority": int(*l.FailoverPriority),
			// there is not zone redundancy information in the FailoverPolicies currently, we have to search it by `id` in the Locations property.
			"zone_redundant": findZoneRedundant(account.Locations, id),
		}

		locationSet.Add(lb)
	}

	return &locationSet
}

func findZoneRedundant(locations *[]documentdb.Location, id string) bool {
	if locations == nil {
		return false
	}
	for _, location := range *locations {
		if location.ID != nil && *location.ID == id {
			if location.IsZoneRedundant != nil {
				return *location.IsZoneRedundant
			}
		}
	}
	return false
}

func isServerlessCapacityMode(accResp documentdb.DatabaseAccountGetResults) bool {
	if props := accResp.DatabaseAccountGetProperties; props != nil && props.Capabilities != nil {
		for _, v := range *props.Capabilities {
			if v.Name != nil && *v.Name == "EnableServerless" {
				return true
			}
		}
	}
	return false
}

func flattenAzureRmCosmosDBAccountCapabilities(capabilities *[]documentdb.Capability) *schema.Set {
	s := schema.Set{
		F: resourceAzureRMCosmosDBAccountCapabilitiesHash,
	}

	for _, c := range *capabilities {
		if v := c.Name; v != nil {
			e := map[string]interface{}{
				"name": *v,
			}
			s.Add(e)
		}
	}

	return &s
}

func flattenAzureRmCosmosDBAccountVirtualNetworkRules(rules *[]documentdb.VirtualNetworkRule) *schema.Set {
	results := schema.Set{
		F: resourceAzureRMCosmosDBAccountVirtualNetworkRuleHash,
	}

	if rules != nil {
		for _, r := range *rules {
			rule := map[string]interface{}{
				"id":                                   *r.ID,
				"ignore_missing_vnet_service_endpoint": *r.IgnoreMissingVNetServiceEndpoint,
			}
			results.Add(rule)
		}
	}

	return &results
}

func resourceAzureRMCosmosDBAccountGeoLocationHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		location := azure.NormalizeLocation(m["location"].(string))
		priority := int32(m["failover_priority"].(int))

		buf.WriteString(fmt.Sprintf("%s-%d", location, priority))
	}

	return schema.HashString(buf.String())
}

func resourceAzureRMCosmosDBAccountCapabilitiesHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	}

	return schema.HashString(buf.String())
}

func resourceAzureRMCosmosDBAccountVirtualNetworkRuleHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(strings.ToLower(m["id"].(string)))
	}

	return schema.HashString(buf.String())
}

func diffSuppressIgnoreKeyVaultKeyVersion(k, old, new string, d *schema.ResourceData) bool {
	oldKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(old)
	if err != nil {
		return false
	}
	newKey, err := keyVaultParse.ParseOptionallyVersionedNestedItemID(new)
	if err != nil {
		return false
	}

	return (oldKey.KeyVaultBaseUrl == newKey.KeyVaultBaseUrl) && (oldKey.Name == newKey.Name)
}

func expandCosmosdbAccountBackupPolicy(input []interface{}) (documentdb.BasicBackupPolicy, error) {
	if len(input) == 0 || input[0] == nil {
		return documentdb.BackupPolicy{
			Type: documentdb.TypeBackupPolicy,
		}, nil
	}
	attr := input[0].(map[string]interface{})

	switch attr["type"].(string) {
	case string(documentdb.TypeContinuous):
		return documentdb.ContinuousModeBackupPolicy{
			Type: documentdb.TypeContinuous,
		}, nil

	case string(documentdb.TypePeriodic):
		return documentdb.PeriodicModeBackupPolicy{
			Type: documentdb.TypePeriodic,
			PeriodicModeProperties: &documentdb.PeriodicModeProperties{
				BackupIntervalInMinutes:        utils.Int32(int32(attr["interval_in_minutes"].(int))),
				BackupRetentionIntervalInHours: utils.Int32(int32(attr["retention_in_hours"].(int))),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown `type` in `backup_policy`:%+v", attr["type"].(string))

	}
}

func flattenCosmosdbAccountBackupPolicy(input documentdb.BasicBackupPolicy) ([]interface{}, error) {
	if input == nil {
		return []interface{}{}, nil
	}

	switch input.(type) {
	case documentdb.ContinuousModeBackupPolicy:
		return []interface{}{
			map[string]interface{}{
				"type": string(documentdb.TypeContinuous),
			},
		}, nil

	case documentdb.PeriodicModeBackupPolicy:
		policy, ok := input.AsPeriodicModeBackupPolicy()
		if !ok {
			return nil, fmt.Errorf("can not transit %+v into `backup_policy` of `type` `Periodic`", input)
		}
		var interval, retention int32
		if v := policy.PeriodicModeProperties.BackupIntervalInMinutes; v != nil {
			interval = *v
		}
		if v := policy.PeriodicModeProperties.BackupRetentionIntervalInHours; v != nil {
			retention = *v
		}
		return []interface{}{
			map[string]interface{}{
				"type":                string(documentdb.TypePeriodic),
				"interval_in_minutes": int(interval),
				"retention_in_hours":  int(retention),
			},
		}, nil

	case documentdb.BackupPolicy:
		return []interface{}{}, nil

	default:
		return nil, fmt.Errorf("unknown `type` in `backup_policy`: %+v", input)

	}
}

func expandCosmosdbAccountIdentity(vs []interface{}) (*documentdb.ManagedServiceIdentity, error) {
	if len(vs) == 0 {
		return &documentdb.ManagedServiceIdentity{
			Type: documentdb.ResourceIdentityTypeNone,
		}, nil
	}

	v := vs[0].(map[string]interface{})
	managedServiceIdentity := documentdb.ManagedServiceIdentity{
		Type: documentdb.ResourceIdentityType(v["type"].(string)),
	}

	var identityIdSet []interface{}
	if identityIds, exists := v["identity_ids"]; exists {
		identityIdSet = identityIds.(*schema.Set).List()
	}

	// If type contains `UserAssigned`, `identity_ids` must be specified and have at least 1 element
	if managedServiceIdentity.Type == documentdb.ResourceIdentityTypeUserAssigned || managedServiceIdentity.Type == documentdb.ResourceIdentityTypeSystemAssignedUserAssigned {
		if len(identityIdSet) == 0 {
			return nil, fmt.Errorf("`identity_ids` must have at least 1 element when `type` includes `UserAssigned`")
		}

		userAssignedIdentities := make(map[string]*documentdb.ManagedServiceIdentityUserAssignedIdentitiesValue)
		for _, id := range identityIdSet {
			userAssignedIdentities[id.(string)] = &documentdb.ManagedServiceIdentityUserAssignedIdentitiesValue{}
		}

		managedServiceIdentity.UserAssignedIdentities = userAssignedIdentities
	} else if len(identityIdSet) > 0 {
		// If type does _not_ contain `UserAssigned` (i.e. is set to `SystemAssigned` or defaulted to `None`), `identity_ids` is not allowed
		return nil, fmt.Errorf("`identity_ids` can only be specified when `type` includes `UserAssigned`; but `type` is currently %q", managedServiceIdentity.Type)
	}

	return &managedServiceIdentity, nil
}

func flattenAzureRmdocumentdbMachineIdentity(identity *documentdb.ManagedServiceIdentity) ([]interface{}, error) {
	if identity == nil || identity.Type == documentdb.ResourceIdentityTypeNone {
		return make([]interface{}, 0), nil
	}

	var principalID, tenantID string
	if identity.PrincipalID != nil {
		principalID = *identity.PrincipalID
	}

	if identity.TenantID != nil {
		tenantID = *identity.TenantID
	}

	result := map[string]interface{}{
		"type":         string(identity.Type),
		"principal_id": principalID,
		"tenant_id":    tenantID,
	}

	identityIds := make([]interface{}, 0)
	if identity.UserAssignedIdentities != nil {
		for key := range identity.UserAssignedIdentities {
			parsedId, err := msiparse.UserAssignedIdentityID(key)
			if err != nil {
				return nil, err
			}
			identityIds = append(identityIds, parsedId.ID())
		}
		result["identity_ids"] = schema.NewSet(schema.HashString, identityIds)
	}

	return []interface{}{result}, nil
}
