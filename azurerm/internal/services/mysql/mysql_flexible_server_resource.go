package mysql

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/mysql/mgmt/2020-07-01-preview/mysqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mysql/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"time"
)

func resourceMysqlFlexibleServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceMysqlFlexibleServerCreate,
		Read:   resourceMysqlFlexibleServerRead,
		Update: resourceMysqlFlexibleServerUpdate,
		Delete: resourceMysqlFlexibleServerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.FlexibleServerID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"administrator_login": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"administrator_login_password": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// will be supported in M6
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					"1",
					"2",
					"3",
				}, false),
			},

			"create_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(mysqlflexibleservers.Default),
					string(mysqlflexibleservers.PointInTimeRestore),
					string(mysqlflexibleservers.Replica),
				}, false),
			},

			"identity": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(mysqlflexibleservers.SystemAssigned),
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
					},
				},
			},

			"restore_point_in_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},

			"source_flexible_server_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(mysqlflexibleservers.FiveFullStopSeven),
				}, false),
			},

			//There is a bug, fix it in the future
			"delegated_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ha_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"maintenance_window": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},

						"start_hour": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},

						"start_minute": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"replication_role": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			// will be optional and computed in the M6
			"sku": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"tier": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(mysqlflexibleservers.Burstable),
								string(mysqlflexibleservers.GeneralPurpose),
								string(mysqlflexibleservers.MemoryOptimized),
							}, false),
						},
					},
				},
			},

			"storage_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_retention_days": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},

						// will be supported in GA
						"storage_autogrow_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"storage_iops": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},

						"storage_mb": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"byok_enforcement_enabled": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ha_state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"public_network_access_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"replica_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			// will be supported in M6
			"standby_availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}
func resourceMysqlFlexibleServerCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).MySQL.FlexibleServersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewFlexibleServerID(subscriptionId, resourceGroup, name).ID()

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Mysqlflexibleservers Server %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_mysql_flexible_server", id)
	}

	properties := mysqlflexibleservers.Server{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		Identity: expandArmServerIdentity(d.Get("identity").([]interface{})),
		Sku:      expandArmServerSku(d.Get("sku").([]interface{})),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	createMode := mysqlflexibleservers.Default
	if v, ok := d.GetOk("create_mode"); ok {
		createMode = mysqlflexibleservers.CreateMode(v.(string))
	}

	if createMode == mysqlflexibleservers.Default {
		if properties.Sku == nil {
			return fmt.Errorf("`Sku` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `Default`", name, resourceGroup)
		}

		if _, ok := d.GetOk("administrator_login"); !ok {
			return fmt.Errorf("`administrator_login` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `Default`", name, resourceGroup)
		}

		if _, ok := d.GetOk("administrator_login_password"); !ok {
			return fmt.Errorf("`administrator_login` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `Default`", name, resourceGroup)
		}

		haEnabled := mysqlflexibleservers.Disabled
		if d.Get("ha_enabled").(bool) {
			haEnabled = mysqlflexibleservers.Enabled
		}

		properties.ServerProperties = &mysqlflexibleservers.ServerProperties{
			AdministratorLogin:         utils.String(d.Get("administrator_login").(string)),
			AvailabilityZone:           utils.String(d.Get("availability_zone").(string)),
			CreateMode:                 mysqlflexibleservers.Default,
			Version:                    mysqlflexibleservers.ServerVersion(d.Get("version").(string)),
			AdministratorLoginPassword: utils.String(d.Get("administrator_login_password").(string)),
			HaEnabled:                  haEnabled,
			ReplicationRole:            utils.String(d.Get("replication_role").(string)),
			StorageProfile:             expandArmServerStorageProfile(d.Get("storage_profile").([]interface{})),
		}

		if v, ok := d.GetOk("delegated_subnet_id"); ok {
			properties.ServerProperties.DelegatedSubnetArguments = &mysqlflexibleservers.DelegatedSubnetArguments{
				SubnetArmResourceID: utils.String(v.(string)),
			}
		}
	}

	if createMode == mysqlflexibleservers.PointInTimeRestore {
		if _, ok := d.GetOk("restore_point_in_time"); createMode == mysqlflexibleservers.PointInTimeRestore && !ok {
			return fmt.Errorf("`restore_point_in_time` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `PointInTimeRestore`", name, resourceGroup)
		}

		if _, ok := d.GetOk("source_flexible_server_id"); !ok {
			return fmt.Errorf("`source_flexible_server_id` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `PointInTimeRestore`", name, resourceGroup)
		}

		restorePointInTime, _ := time.Parse(time.RFC3339, d.Get("restore_point_in_time").(string))

		properties.ServerProperties = &mysqlflexibleservers.ServerProperties{
			CreateMode:         mysqlflexibleservers.PointInTimeRestore,
			RestorePointInTime: &date.Time{Time: restorePointInTime},
			SourceServerID:     utils.String(d.Get("source_flexible_server_id").(string)),
			ReplicationRole:    utils.String(d.Get("replication_role").(string)),
			StorageProfile:     expandArmServerStorageProfile(d.Get("storage_profile").([]interface{})),
		}
	}

	if createMode == mysqlflexibleservers.Replica {
		if _, ok := d.GetOk("source_flexible_server_id"); !ok {
			return fmt.Errorf("`source_flexible_server_id` is required for Mysql Flexible Server %q (Resource Group %q) when `create_mode` is `Replica`", name, resourceGroup)
		}

		properties.ServerProperties = &mysqlflexibleservers.ServerProperties{
			CreateMode:      mysqlflexibleservers.Replica,
			SourceServerID:  utils.String(d.Get("source_flexible_server_id").(string)),
			ReplicationRole: utils.String(d.Get("replication_role").(string)),
			StorageProfile:  expandArmServerStorageProfile(d.Get("storage_profile").([]interface{})),
		}
	}

	future, err := client.Create(ctx, resourceGroup, name, properties)
	if err != nil {
		return fmt.Errorf("creating Mysql Flexible Server %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of the Mysql Flexible Server %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id)

	mwProps := mysqlflexibleservers.ServerForUpdate{
		ServerPropertiesForUpdate: &mysqlflexibleservers.ServerPropertiesForUpdate{
			MaintenanceWindow: expandArmServerMaintenanceWindow(d.Get("maintenance_window").([]interface{})),
		},
	}

	mwfuture, err := client.Update(ctx, resourceGroup, name, mwProps)
	if err != nil {
		return fmt.Errorf("updating Mysql Flexible Server %q (Resource Group %q) Maintenance Window: %+v", name, resourceGroup, err)
	}

	if err := mwfuture.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation of the Mysql Flexible Server %q (Resource Group %q) Maintenance Window: %+v", name, resourceGroup, err)
	}

	return resourceMysqlFlexibleServerRead(d, meta)
}

func resourceMysqlFlexibleServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MySQL.FlexibleServersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Mysql Flexible Server %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Mysql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if err := d.Set("identity", flattenArmServerIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}
	if props := resp.ServerProperties; props != nil {
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("availability_zone", props.AvailabilityZone)
		if sn := props.DelegatedSubnetArguments; sn != nil {
			d.Set("delegated_subnet_id", sn.SubnetArmResourceID)
		}

		if err := d.Set("maintenance_window", flattenArmServerMaintenanceWindow(props.MaintenanceWindow)); err != nil {
			return fmt.Errorf("setting `maintenance_window`: %+v", err)
		}

		d.Set("ha_enabled", props.HaEnabled == mysqlflexibleservers.Enabled)
		d.Set("replication_role", props.ReplicationRole)
		if rpit := props.RestorePointInTime; rpit != nil {
			d.Set("restore_point_in_time", (*rpit).Format(time.RFC3339))
		}
		d.Set("source_flexible_server_id", props.SourceServerID)
		if err := d.Set("storage_profile", flattenArmServerStorageProfile(props.StorageProfile)); err != nil {
			return fmt.Errorf("setting `storage_profile`: %+v", err)
		}
		d.Set("version", props.Version)
		d.Set("byok_enforcement_enabled", props.ByokEnforcement)
		d.Set("fqdn", props.FullyQualifiedDomainName)
		d.Set("ha_state", props.HaState)
		d.Set("public_network_access_enabled", props.PublicNetworkAccess == mysqlflexibleservers.PublicNetworkAccessEnumEnabled)
		d.Set("replica_capacity", props.ReplicaCapacity)
		d.Set("standby_availability_zone", props.StandbyAvailabilityZone)
	}
	if err := d.Set("sku", flattenArmServerSku(resp.Sku)); err != nil {
		return fmt.Errorf("setting `sku`: %+v", err)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceMysqlFlexibleServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MySQL.FlexibleServersClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	parameters := mysqlflexibleservers.ServerForUpdate{
		ServerPropertiesForUpdate: &mysqlflexibleservers.ServerPropertiesForUpdate{},
	}

	if d.HasChange("storage_profile") {
		parameters.ServerPropertiesForUpdate.StorageProfile = expandArmServerStorageProfile(d.Get("storage_profile").([]interface{}))
	}

	if d.HasChange("administrator_login_password") {
		parameters.ServerPropertiesForUpdate.AdministratorLoginPassword = utils.String(d.Get("administrator_login_password").(string))
	}

	if d.HasChange("delegated_subnet_id") {
		parameters.ServerPropertiesForUpdate.DelegatedSubnetArguments = &mysqlflexibleservers.DelegatedSubnetArguments{
			SubnetArmResourceID: utils.String(d.Get("delegated_subnet_id").(string)),
		}
	}

	if d.HasChange("ha_enabled") {
		haEnabled := mysqlflexibleservers.Disabled
		if d.Get("ha_enabled").(bool) {
			haEnabled = mysqlflexibleservers.Enabled
		}
		parameters.ServerPropertiesForUpdate.HaEnabled = haEnabled
	}

	if d.HasChange("maintenance_window") {
		parameters.ServerPropertiesForUpdate.MaintenanceWindow = expandArmServerMaintenanceWindow(d.Get("maintenance_window").([]interface{}))
	}

	if d.HasChange("replication_role") {
		parameters.ServerPropertiesForUpdate.ReplicationRole = utils.String(d.Get("replication_role").(string))
	}

	if d.HasChange("sku") {
		parameters.Sku = expandArmServerSku(d.Get("sku").([]interface{}))
	}

	if d.HasChange("tags") {
		parameters.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("updating Mysql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of the Mysql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return resourceMysqlFlexibleServerRead(d, meta)
}

func resourceMysqlFlexibleServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MySQL.FlexibleServersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Mysql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of the Mysql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return nil
}

func expandArmServerIdentity(input []interface{}) *mysqlflexibleservers.Identity {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &mysqlflexibleservers.Identity{
		Type: mysqlflexibleservers.ResourceIdentityType(v["type"].(string)),
	}
}

func expandArmServerMaintenanceWindow(input []interface{}) *mysqlflexibleservers.MaintenanceWindow {
	if len(input) == 0 {
		return &mysqlflexibleservers.MaintenanceWindow{
			CustomWindow: utils.String("Disabled"),
		}
	}
	v := input[0].(map[string]interface{})
	return &mysqlflexibleservers.MaintenanceWindow{
		CustomWindow: utils.String("Enabled"),
		StartHour:    utils.Int32(int32(v["start_hour"].(int))),
		StartMinute:  utils.Int32(int32(v["start_minute"].(int))),
		DayOfWeek:    utils.Int32(int32(v["day_of_week"].(int))),
	}
}

func expandArmServerStorageProfile(input []interface{}) *mysqlflexibleservers.StorageProfile {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	storageAutogrow := mysqlflexibleservers.StorageAutogrowDisabled
	if v["storage_autogrow_enabled"].(bool) {
		storageAutogrow = mysqlflexibleservers.StorageAutogrowEnabled
	}
	return &mysqlflexibleservers.StorageProfile{
		BackupRetentionDays: utils.Int32(int32(v["backup_retention_days"].(int))),
		StorageMB:           utils.Int32(int32(v["storage_mb"].(int))),
		StorageIops:         utils.Int32(int32(v["storage_iops"].(int))),
		StorageAutogrow:     storageAutogrow,
	}
}

func expandArmServerSku(input []interface{}) *mysqlflexibleservers.Sku {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &mysqlflexibleservers.Sku{
		Name: utils.String(v["name"].(string)),
		Tier: mysqlflexibleservers.SkuTier(v["tier"].(string)),
	}
}

func flattenArmServerIdentity(input *mysqlflexibleservers.Identity) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var principalId string
	if input.PrincipalID != nil {
		principalId = *input.PrincipalID
	}
	var tenantId string
	if input.TenantID != nil {
		tenantId = *input.TenantID
	}
	return []interface{}{
		map[string]interface{}{
			"type":         string(input.Type),
			"principal_id": principalId,
			"tenant_id":    tenantId,
		},
	}
}

func flattenArmServerMaintenanceWindow(input *mysqlflexibleservers.MaintenanceWindow) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var dayOfWeek int32
	if input.DayOfWeek != nil {
		dayOfWeek = *input.DayOfWeek
	}
	var startHour int32
	if input.StartHour != nil {
		startHour = *input.StartHour
	}
	var startMinute int32
	if input.StartMinute != nil {
		startMinute = *input.StartMinute
	}
	return []interface{}{
		map[string]interface{}{
			"day_of_week":  dayOfWeek,
			"start_hour":   startHour,
			"start_minute": startMinute,
		},
	}
}

func flattenArmServerStorageProfile(input *mysqlflexibleservers.StorageProfile) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var backupRetentionDays int32
	if input.BackupRetentionDays != nil {
		backupRetentionDays = *input.BackupRetentionDays
	}

	var storageIops int32
	if input.StorageIops != nil {
		storageIops = *input.StorageIops
	}

	var storageMb int32
	if input.StorageMB != nil {
		storageMb = *input.StorageMB
	}
	return []interface{}{
		map[string]interface{}{
			"backup_retention_days":    backupRetentionDays,
			"storage_autogrow_enabled": input.StorageAutogrow == mysqlflexibleservers.StorageAutogrowEnabled,
			"storage_iops":             storageIops,
			"storage_mb":               storageMb,
		},
	}
}

func flattenArmServerSku(input *mysqlflexibleservers.Sku) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var name string
	if input.Name != nil {
		name = *input.Name
	}
	var tier mysqlflexibleservers.SkuTier
	if input.Tier != "" {
		tier = input.Tier
	}
	return []interface{}{
		map[string]interface{}{
			"name": name,
			"tier": tier,
		},
	}
}
