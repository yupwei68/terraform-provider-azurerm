package synapse

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/synapse/mgmt/2019-06-01-preview/synapse"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	mssqlParse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mssql/parse"
	mssqlValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mssql/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/synapse/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/synapse/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

const (
	DefaultCreateMode            = "Default"
	RecoveryCreateMode           = "Recovery"
	PointInTimeRestoreCreateMode = "PointInTimeRestore"
)

func resourceArmSynapseSqlPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSynapseSqlPoolCreate,
		Read:   resourceArmSynapseSqlPoolRead,
		Update: resourceArmSynapseSqlPoolUpdate,
		Delete: resourceArmSynapseSqlPoolDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if _, err := parse.SynapseSqlPoolID(d.Id()); err != nil {
					return []*schema.ResourceData{d}, err
				}

				d.Set("create_mode", DefaultCreateMode)
				if v, ok := d.GetOk("create_mode"); ok && v.(string) != "" {
					d.Set("create_mode", v)
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SqlPoolName,
			},

			"synapse_workspace_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SynapseWorkspaceID,
			},

			"sku_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"DW100c",
					"DW200c",
					"DW300c",
					"DW400c",
					"DW500c",
					"DW1000c",
					"DW1500c",
					"DW2000c",
					"DW2500c",
					"DW3000c",
					"DW5000c",
					"DW6000c",
					"DW7500c",
					"DW10000c",
					"DW15000c",
					"DW30000c",
				}, false),
			},

			"create_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultCreateMode,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					DefaultCreateMode,
					RecoveryCreateMode,
					PointInTimeRestoreCreateMode,
				}, false),
			},

			"collation": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: mssqlValidate.MsSqlDBCollation(),
			},

			"recovery_database_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"restore"},
				ValidateFunc: validation.Any(
					validate.SqlPoolID,
					mssqlValidate.MsSqlDatabaseID,
				),
			},

			"restore": {
				Type:          schema.TypeList,
				ForceNew:      true,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"recovery_database_id"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"point_in_time": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},

						"source_database_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.Any(
								validate.SqlPoolID,
								mssqlValidate.MsSqlDatabaseID,
							),
						},
					},
				},
			},

			"data_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmSynapseSqlPoolCreate(d *schema.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	workspaceClient := meta.(*clients.Client).Synapse.WorkspaceClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	workspaceId, err := parse.SynapseWorkspaceID(d.Get("synapse_workspace_id").(string))
	if err != nil {
		return err
	}

	existing, err := sqlClient.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Synapse Sql Pool %q (workspace %q / Resource Group %q): %+v", name, workspaceId.Name, workspaceId.ResourceGroup, err)
		}
	}
	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_synapse_sql_pool", *existing.ID)
	}

	workspace, err := workspaceClient.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name)
	if err != nil {
		return fmt.Errorf("retrieving Synapse workspace %q (Resource Group %q): %+v", workspaceId.Name, workspaceId.ResourceGroup, err)
	}

	mode := d.Get("create_mode").(string)
	sqlPoolInfo := synapse.SQLPool{
		Location: workspace.Location,
		SQLPoolResourceProperties: &synapse.SQLPoolResourceProperties{
			CreateMode: utils.String(mode),
		},
		Sku: &synapse.Sku{
			Name: utils.String(d.Get("sku_name").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	switch mode {
	case DefaultCreateMode:
		sqlPoolInfo.SQLPoolResourceProperties.Collation = utils.String(d.Get("collation").(string))
	case RecoveryCreateMode:
		recoveryDatabaseId := constructSourceDatabaseId(d.Get("recovery_database_id").(string), subscriptionId)
		if recoveryDatabaseId == "" {
			return fmt.Errorf("`recovery_database_id` must be set when `create_mode` is %q", RecoveryCreateMode)
		}
		sqlPoolInfo.SQLPoolResourceProperties.RecoverableDatabaseID = utils.String(recoveryDatabaseId)
	case PointInTimeRestoreCreateMode:
		restore := d.Get("restore").([]interface{})
		if len(restore) == 0 || restore[0] == nil {
			return fmt.Errorf("`restore` block must be set when `create_mode` is %q", PointInTimeRestoreCreateMode)
		}
		v := restore[0].(map[string]interface{})
		sourceDatabaseId := constructSourceDatabaseId(v["source_database_id"].(string), subscriptionId)
		restorePointInTime, err := time.Parse(time.RFC3339, v["point_in_time"].(string))
		if err != nil {
			return err
		}
		sqlPoolInfo.SQLPoolResourceProperties.RestorePointInTime = &date.Time{Time: restorePointInTime}
		sqlPoolInfo.SQLPoolResourceProperties.SourceDatabaseID = utils.String(sourceDatabaseId)
	}

	future, err := sqlClient.Create(ctx, workspaceId.ResourceGroup, workspaceId.Name, name, sqlPoolInfo)
	if err != nil {
		return fmt.Errorf("creating Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}
	if err = future.WaitForCompletionRef(ctx, sqlClient.Client); err != nil {
		return fmt.Errorf("waiting on creating future for Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}

	if d.Get("data_encrypted").(bool) {
		parameter := synapse.TransparentDataEncryption{
			TransparentDataEncryptionProperties: &synapse.TransparentDataEncryptionProperties{
				Status: synapse.TransparentDataEncryptionStatusEnabled,
			},
		}
		if _, err := sqlPTDEClient.CreateOrUpdate(ctx, workspaceId.ResourceGroup, workspaceId.Name, name, parameter); err != nil {
			return fmt.Errorf("setting `data_encrypted`: %+v", err)
		}
	}

	resp, err := sqlClient.Get(ctx, workspaceId.ResourceGroup, workspaceId.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", name, workspaceId.ResourceGroup, workspaceId.Name, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Synapse SqlPool %q (Resource Group %q / workspaceName %q) ID", name, workspaceId.ResourceGroup, workspaceId.Name)
	}

	d.SetId(*resp.ID)
	return resourceArmSynapseSqlPoolRead(d, meta)
}

func resourceArmSynapseSqlPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SynapseSqlPoolID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("data_encrypted") {
		status := synapse.TransparentDataEncryptionStatusDisabled
		if d.Get("data_encrypted").(bool) {
			status = synapse.TransparentDataEncryptionStatusEnabled
		}

		parameter := synapse.TransparentDataEncryption{
			TransparentDataEncryptionProperties: &synapse.TransparentDataEncryptionProperties{
				Status: status,
			},
		}
		if _, err := sqlPTDEClient.CreateOrUpdate(ctx, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name, parameter); err != nil {
			return fmt.Errorf("updating `data_encrypted`: %+v", err)
		}
	}

	if d.HasChanges("sku_name", "tags") {
		sqlPoolInfo := synapse.SQLPoolPatchInfo{
			Sku: &synapse.Sku{
				Name: utils.String(d.Get("sku_name").(string)),
			},
			Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
		}

		if _, err := sqlClient.Update(ctx, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name, sqlPoolInfo); err != nil {
			return fmt.Errorf("updating Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", id.Name, id.Workspace.ResourceGroup, id.Workspace.Name, err)
		}

		// wait for sku scale completion
		if d.HasChange("sku_name") {
			stateConf := &resource.StateChangeConf{
				Pending: []string{
					"Scaling",
				},
				Target: []string{
					"Online",
				},
				Refresh:                   synapseSqlPoolScaleStateRefreshFunc(ctx, sqlClient, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name),
				MinTimeout:                5 * time.Second,
				ContinuousTargetOccurence: 3,
				Timeout:                   d.Timeout(schema.TimeoutUpdate),
			}

			if _, err := stateConf.WaitForState(); err != nil {
				return fmt.Errorf("waiting for scaling of Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", id.Name, id.Workspace.ResourceGroup, id.Workspace.Name, err)
			}
		}
	}

	return resourceArmSynapseSqlPoolRead(d, meta)
}

func resourceArmSynapseSqlPoolRead(d *schema.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	sqlPTDEClient := meta.(*clients.Client).Synapse.SqlPoolTransparentDataEncryptionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SynapseSqlPoolID(d.Id())
	if err != nil {
		return err
	}

	resp, err := sqlClient.Get(ctx, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Synapse SQL Pool %q (Workspace %q / Resource Group %q) does not exist - removing from state", id.Name, id.Workspace.Name, id.Workspace.ResourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Synapse SqlPool %q (workspace %q / Resource Group %q): %+v", id.Name, id.Workspace.Name, id.Workspace.ResourceGroup, err)
	}

	transparentDataEncryption, err := sqlPTDEClient.Get(ctx, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name)
	if err != nil {
		return fmt.Errorf("retrieving Transparent Data Encryption settings of Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", id.Name, id.Workspace.ResourceGroup, id.Workspace.Name, err)
	}

	d.Set("name", id.Name)
	d.Set("synapse_workspace_id", id.Workspace.String())
	if resp.Sku != nil {
		d.Set("sku_name", resp.Sku.Name)
	}
	if props := resp.SQLPoolResourceProperties; props != nil {
		d.Set("collation", props.Collation)
	}
	if props := transparentDataEncryption.TransparentDataEncryptionProperties; props != nil {
		d.Set("data_encrypted", props.Status == synapse.TransparentDataEncryptionStatusEnabled)
	}

	// whole "restore" block is not returned. to avoid conflict, so set it from the old state
	d.Set("restore", d.Get("restore").([]interface{}))

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmSynapseSqlPoolDelete(d *schema.ResourceData, meta interface{}) error {
	sqlClient := meta.(*clients.Client).Synapse.SqlPoolClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SynapseSqlPoolID(d.Id())
	if err != nil {
		return err
	}

	future, err := sqlClient.Delete(ctx, id.Workspace.ResourceGroup, id.Workspace.Name, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", id.Name, id.Workspace.ResourceGroup, id.Workspace.Name, err)
	}

	if err = future.WaitForCompletionRef(ctx, sqlClient.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Synapse SqlPool %q (Resource Group %q / workspaceName %q): %+v", id.Name, id.Workspace.ResourceGroup, id.Workspace.Name, err)
	}
	return nil
}

func synapseSqlPoolScaleStateRefreshFunc(ctx context.Context, client *synapse.SQLPoolsClient, resourceGroup, workspaceName, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := client.Get(ctx, resourceGroup, workspaceName, name)
		if err != nil {
			return resp, "failed", err
		}
		if resp.SQLPoolResourceProperties == nil || resp.SQLPoolResourceProperties.Status == nil {
			return resp, "failed", nil
		}
		return resp, *resp.SQLPoolResourceProperties.Status, nil
	}
}

// sqlPool backend service is a proxy to sql database
// backend service restore and backup only accept id format of sql database
// so if the id is sqlPool, we need to construct the corresponding sql database id
func constructSourceDatabaseId(id string, subscriptionId string) string {
	sqlPoolId, err := parse.SynapseSqlPoolID(id)
	if err != nil {
		return id
	}
	return mssqlParse.NewMsSqlDatabaseID(sqlPoolId.Workspace.ResourceGroup, sqlPoolId.Workspace.Name, sqlPoolId.Name).ID(subscriptionId)
}
