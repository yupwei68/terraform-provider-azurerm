package postgres

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2021-06-01/postgresqlflexibleservers"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	networkValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/postgres/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/postgres/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourcePostgresqlFlexibleServer() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePostgresqlFlexibleServerCreate,
		Read:   resourcePostgresqlFlexibleServerRead,
		Update: resourcePostgresqlFlexibleServerUpdate,
		Delete: resourcePostgresqlFlexibleServerDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(1 * time.Hour),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(1 * time.Hour),
			Delete: pluginsdk.DefaultTimeout(1 * time.Hour),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FlexibleServerID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FlexibleServerName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"administrator_login": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"administrator_password": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"sku_name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validate.FlexibleServerSkuName,
			},

			"storage_mb": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntInSlice([]int{32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608, 16777216, 33554432}),
			},

			"version": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(postgresqlflexibleservers.ServerVersionOneOne),
					string(postgresqlflexibleservers.ServerVersionOneTwo),
					string(postgresqlflexibleservers.ServerVersionOneThree),
				}, false),
			},

			"zone": azure.SchemaZoneComputed(),

			"create_mode": {
				Type:     pluginsdk.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(postgresqlflexibleservers.CreateModeDefault),
					string(postgresqlflexibleservers.CreateModePointInTimeRestore),
				}, false),
			},

			"delegated_subnet_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: networkValidate.SubnetID,
			},

			"point_in_time_restore_time_in_utc": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},

			"source_server_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validate.FlexibleServerID,
			},

			"maintenance_window": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"day_of_week": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 6),
						},

						"start_hour": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},

						"start_minute": {
							Type:         pluginsdk.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},

			"backup_retention_days": {
				Type:         pluginsdk.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(7, 35),
			},

			"geo_redundant_backup_enabled": {
				Type:     pluginsdk.TypeBool,
				Optional: true,
				Default:  false,
			},

			"high_availability": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"standby_availability_zone": azure.SchemaZoneComputed(),
					},
				},
			},

			"cmk_enabled": {
				Type:       pluginsdk.TypeString,
				Computed:   true,
				Deprecated: "This field is deprecated and will be removed in version 3.0 of the Azure Provider",
			},

			"fqdn": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"public_network_access_enabled": {
				Type:     pluginsdk.TypeBool,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}
func resourcePostgresqlFlexibleServerCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Postgres.FlexibleServersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewFlexibleServerID(subscriptionId, resourceGroup, name)

	existing, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_postgresql_flexible_server", id.ID())
	}

	createMode := d.Get("create_mode").(string)

	if postgresqlflexibleservers.CreateMode(createMode) == postgresqlflexibleservers.CreateModePointInTimeRestore {
		if _, ok := d.GetOk("source_server_id"); !ok {
			return fmt.Errorf("`source_server_id` is required when `create_mode` is `PointInTimeRestore`")
		}
		if _, ok := d.GetOk("point_in_time_restore_time_in_utc"); !ok {
			return fmt.Errorf("`point_in_time_restore_time_in_utc` is required when `create_mode` is `PointInTimeRestore`")
		}
	}

	if createMode == "" || postgresqlflexibleservers.CreateMode(createMode) == postgresqlflexibleservers.CreateModeDefault {
		if _, ok := d.GetOk("administrator_login"); !ok {
			return fmt.Errorf("`administrator_login` is required when `create_mode` is `Default`")
		}
		if _, ok := d.GetOk("administrator_password"); !ok {
			return fmt.Errorf("`administrator_password` is required when `create_mode` is `Default`")
		}
		if _, ok := d.GetOk("sku_name"); !ok {
			return fmt.Errorf("`sku_name` is required when `create_mode` is `Default`")
		}
		if _, ok := d.GetOk("version"); !ok {
			return fmt.Errorf("`version` is required when `create_mode` is `Default`")
		}
		if _, ok := d.GetOk("storage_mb"); !ok {
			return fmt.Errorf("`storage_mb` is required when `create_mode` is `Default`")
		}
	}

	sku, err := expandFlexibleServerSku(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("expanding `sku_name` for PostgreSQL Flexible Server %s (Resource Group %q): %v", id.Name, id.ResourceGroup, err)
	}

	geoRedundantBackup := postgresqlflexibleservers.GeoRedundantBackupEnumDisabled
	if d.Get("geo_redundant_backup_enabled").(bool) {
		geoRedundantBackup = postgresqlflexibleservers.GeoRedundantBackupEnumEnabled
	}

	parameters := postgresqlflexibleservers.Server{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		ServerProperties: &postgresqlflexibleservers.ServerProperties{
			CreateMode: postgresqlflexibleservers.CreateMode(d.Get("create_mode").(string)),
			Network: &postgresqlflexibleservers.Network{
				DelegatedSubnetResourceID: utils.String(d.Get("delegated_subnet_id").(string)),
			},
			Version: postgresqlflexibleservers.ServerVersion(d.Get("version").(string)),
			Storage: &postgresqlflexibleservers.Storage{
				StorageSizeGB: utils.Int32(int32(d.Get("storage_mb").(int) / 1024)),
			},
			HighAvailability: expandFlexibleServerHighAvailability(d.Get("high_availability").([]interface{})),
			Backup: &postgresqlflexibleservers.Backup{
				GeoRedundantBackup: geoRedundantBackup,
			},
		},
		Sku:  sku,
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("administrator_login"); ok && v.(string) != "" {
		parameters.ServerProperties.AdministratorLogin = utils.String(v.(string))
	}

	if v, ok := d.GetOk("administrator_password"); ok && v.(string) != "" {
		parameters.ServerProperties.AdministratorLoginPassword = utils.String(v.(string))
	}

	if v, ok := d.GetOk("zone"); ok && v.(string) != "" {
		parameters.ServerProperties.AvailabilityZone = utils.String(v.(string))
	}

	if v, ok := d.GetOk("source_server_id"); ok && v.(string) != "" {
		parameters.ServerProperties.SourceServerResourceID = utils.String(v.(string))
	}

	if v, ok := d.GetOk("backup_retention_days"); ok {
		parameters.ServerProperties.Backup.BackupRetentionDays = utils.Int32(int32(v.(int)))
	}

	pointInTimeUTC := d.Get("point_in_time_restore_time_in_utc").(string)
	if pointInTimeUTC != "" {
		v, err := time.Parse(time.RFC3339, pointInTimeUTC)
		if err != nil {
			return fmt.Errorf("unable to parse `point_in_time_restore_time_in_utc` value")
		}
		parameters.ServerProperties.PointInTimeUTC = &date.Time{Time: v}
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("creating Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of the Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	// `maintenance_window` could only be updated with, could not be created with
	if v, ok := d.GetOk("maintenance_window"); ok {
		mwParams := postgresqlflexibleservers.ServerForUpdate{
			ServerPropertiesForUpdate: &postgresqlflexibleservers.ServerPropertiesForUpdate{
				MaintenanceWindow: expandArmServerMaintenanceWindow(v.([]interface{})),
			},
		}
		mwFuture, err := client.Update(ctx, id.ResourceGroup, id.Name, mwParams)
		if err != nil {
			return fmt.Errorf("updating Postgresql Flexible Server %q maintenance window (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}

		if err := mwFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("waiting for the update of the Postgresql Flexible Server %q maintenance window (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
		}
	}

	d.SetId(id.ID())

	return resourcePostgresqlFlexibleServerRead(d, meta)
}

func resourcePostgresqlFlexibleServerRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.FlexibleServersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Postgresql Flexibleserver %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if props := resp.ServerProperties; props != nil {
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("zone", props.AvailabilityZone)
		d.Set("version", props.Version)
		d.Set("fqdn", props.FullyQualifiedDomainName)

		if props.Network != nil {
			d.Set("public_network_access_enabled", props.Network.PublicNetworkAccess == postgresqlflexibleservers.ServerPublicNetworkAccessStateEnabled)
			d.Set("delegated_subnet_id", props.Network.DelegatedSubnetResourceID)
		}

		if err := d.Set("maintenance_window", flattenArmServerMaintenanceWindow(props.MaintenanceWindow)); err != nil {
			return fmt.Errorf("setting `maintenance_window`: %+v", err)
		}

		if storage := props.Storage; storage != nil {
			d.Set("storage_mb", (*storage.StorageSizeGB)*1024)
		}

		if backup := props.Backup; backup != nil {
			d.Set("backup_retention_days", backup.BackupRetentionDays)
			d.Set("geo_redundant_backup_enabled", backup.GeoRedundantBackup == postgresqlflexibleservers.GeoRedundantBackupEnumEnabled)
		}

		if err := d.Set("high_availability", flattenFlexibleServerHighAvailability(props.HighAvailability)); err != nil {
			return fmt.Errorf("setting `high_availability`: %+v", err)
		}
	}

	sku, err := flattenFlexibleServerSku(resp.Sku)
	if err != nil {
		return fmt.Errorf("flattening `sku_name` for PostgreSQL Flexible Server %s (Resource Group %q): %v", id.Name, id.ResourceGroup, err)
	}

	d.Set("sku_name", sku)

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourcePostgresqlFlexibleServerUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.FlexibleServersClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	parameters := postgresqlflexibleservers.ServerForUpdate{
		Location:                  utils.String(location.Normalize(d.Get("location").(string))),
		ServerPropertiesForUpdate: &postgresqlflexibleservers.ServerPropertiesForUpdate{},
	}

	if d.HasChange("administrator_password") {
		parameters.ServerPropertiesForUpdate.AdministratorLoginPassword = utils.String(d.Get("administrator_password").(string))
	}

	if d.HasChange("storage_mb") {
		parameters.ServerPropertiesForUpdate.Storage = &postgresqlflexibleservers.Storage{
			StorageSizeGB: utils.Int32(int32(d.Get("storage_mb").(int) / 1024)),
		}
	}

	if d.HasChange("backup_retention_days") || d.HasChange("geo_redundant_backup_enabled") {
		geoRedundantBackup := postgresqlflexibleservers.GeoRedundantBackupEnumDisabled
		if d.Get("geo_redundant_backup_enabled").(bool) {
			geoRedundantBackup = postgresqlflexibleservers.GeoRedundantBackupEnumEnabled
		}
		parameters.ServerPropertiesForUpdate.Backup = &postgresqlflexibleservers.Backup{
			BackupRetentionDays: utils.Int32(int32(d.Get("backup_retention_days").(int))),
			GeoRedundantBackup:  geoRedundantBackup,
		}
	}

	if d.HasChange("maintenance_window") {
		parameters.ServerPropertiesForUpdate.MaintenanceWindow = expandArmServerMaintenanceWindow(d.Get("maintenance_window").([]interface{}))
	}

	if d.HasChange("sku_name") {
		sku, err := expandFlexibleServerSku(d.Get("sku_name").(string))
		if err != nil {
			return fmt.Errorf("expanding `sku_name` for PostgreSQL Flexible Server %s (Resource Group %q): %v", id.Name, id.ResourceGroup, err)
		}
		parameters.Sku = sku
	}

	if d.HasChange("tags") {
		parameters.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if d.HasChange("high_availability") {
		parameters.HighAvailability = expandFlexibleServerHighAvailability(d.Get("high_availability").([]interface{}))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("updating Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the update of the Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return resourcePostgresqlFlexibleServerRead(d, meta)
}

func resourcePostgresqlFlexibleServerDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.FlexibleServersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of the Postgresql Flexible Server %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	return nil
}

func expandArmServerMaintenanceWindow(input []interface{}) *postgresqlflexibleservers.MaintenanceWindow {
	if len(input) == 0 {
		return &postgresqlflexibleservers.MaintenanceWindow{
			CustomWindow: utils.String("Disabled"),
		}
	}
	v := input[0].(map[string]interface{})

	maintenanceWindow := postgresqlflexibleservers.MaintenanceWindow{
		CustomWindow: utils.String(string("Enabled")),
		StartHour:    utils.Int32(int32(v["start_hour"].(int))),
		StartMinute:  utils.Int32(int32(v["start_minute"].(int))),
		DayOfWeek:    utils.Int32(int32(v["day_of_week"].(int))),
	}

	return &maintenanceWindow
}

func expandFlexibleServerSku(name string) (*postgresqlflexibleservers.Sku, error) {
	if name == "" {
		return nil, nil
	}
	parts := strings.SplitAfterN(name, "_", 2)

	var tier postgresqlflexibleservers.SkuTier
	switch strings.TrimSuffix(parts[0], "_") {
	case "B":
		tier = postgresqlflexibleservers.SkuTierBurstable
	case "GP":
		tier = postgresqlflexibleservers.SkuTierGeneralPurpose
	case "MO":
		tier = postgresqlflexibleservers.SkuTierMemoryOptimized
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", name, parts[0])
	}

	return &postgresqlflexibleservers.Sku{
		Name: utils.String(parts[1]),
		Tier: tier,
	}, nil
}

func flattenFlexibleServerSku(sku *postgresqlflexibleservers.Sku) (string, error) {
	if sku == nil || sku.Name == nil || sku.Tier == "" {
		return "", nil
	}

	var tier string
	switch sku.Tier {
	case postgresqlflexibleservers.SkuTierBurstable:
		tier = "B"
	case postgresqlflexibleservers.SkuTierGeneralPurpose:
		tier = "GP"
	case postgresqlflexibleservers.SkuTierMemoryOptimized:
		tier = "MO"
	default:
		return "", fmt.Errorf("sku_name has unknown sku tier %s", sku.Tier)
	}

	return strings.Join([]string{tier, *sku.Name}, "_"), nil
}

func flattenArmServerMaintenanceWindow(input *postgresqlflexibleservers.MaintenanceWindow) []interface{} {
	if input == nil || input.CustomWindow == nil || *input.CustomWindow == "Disabled" {
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

func expandFlexibleServerHighAvailability(inputs []interface{}) *postgresqlflexibleservers.HighAvailability {
	if len(inputs) == 0 {
		return &postgresqlflexibleservers.HighAvailability{
			Mode: postgresqlflexibleservers.HighAvailabilityModeDisabled,
		}
	}
	result := postgresqlflexibleservers.HighAvailability{
		Mode: postgresqlflexibleservers.HighAvailabilityModeZoneRedundant,
	}

	if inputs[0] != nil {
		input := inputs[0].(map[string]interface{})

		if v, ok := input["standby_availability_zone"]; ok && v.(string) != "" {
			result.StandbyAvailabilityZone = utils.String(v.(string))
		}
	}

	return &result
}

func flattenFlexibleServerHighAvailability(ha *postgresqlflexibleservers.HighAvailability) []interface{} {
	if ha == nil || (*ha).Mode == postgresqlflexibleservers.HighAvailabilityModeDisabled {
		return []interface{}{}
	}

	var zone string
	if ha.StandbyAvailabilityZone != nil {
		zone = *ha.StandbyAvailabilityZone
	}

	return []interface{}{
		map[string]interface{}{
			"standby_availability_zone": zone,
		},
	}
}
