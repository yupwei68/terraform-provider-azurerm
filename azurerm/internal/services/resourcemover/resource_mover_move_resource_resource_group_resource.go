package resourcemover

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/resourcemover/mgmt/2019-10-01-preview/resourcemover"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/parse"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"time"
)

func resourceResourceMoverMoveResourceResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceMoverMoveResourceResourceGroupCreateUpdate,
		Read:   resourceResourceMoverMoveResourceResourceGroupRead,
		Update: resourceResourceMoverMoveResourceResourceGroupCreateUpdate,
		Delete: resourceResourceMoverMoveResourceResourceGroupDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ResourceMoverMoveResourceID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"move_collection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_setting": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_resource_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"source_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"existing_target_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"error": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Type:     schema.TypeString,
							Computed: true,
						},

						//"detail": {
						//	Type:     schema.TypeString,
						//	Computed: true,
						//},

						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"move_status": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"error": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code": {
										Type:     schema.TypeString,
										Computed: true,
									},

									//"detail": {
									//	Type:     schema.TypeString,
									//	Computed: true,
									//},

									"message": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"target": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"job_status": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"job_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"job_progress": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"move_state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"target_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
func resourceResourceMoverMoveResourceResourceGroupCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ResourceMover.MoveResourceClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	moveCollectionId, err := parse.ResourceMoverMoveCollectionID(d.Get("move_collection_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewResourceMoverMoveResourceID(subscriptionId, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, name).ID()

	if d.IsNewResource() {
		existing, err := client.Get(ctx, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", name, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_resource_mover_move_resource_resource_group", id)
		}
	}

	properties := resourcemover.MoveResource{
		Properties: &resourcemover.MoveResourceProperties{
			ResourceSettings: expandArmMoveResourceResourceGroupResourceSettings(d.Get("resource_setting").([]interface{})),
			SourceID:         utils.String(d.Get("source_id").(string)),
		},
	}

	if v, ok := d.GetOk("existing_target_id"); ok {
		properties.Properties.ExistingTargetID = utils.String(v.(string))
	}
	future, err := client.Create(ctx, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, name, &properties)
	if err != nil {
		return fmt.Errorf("creating/updating Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", name, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation/update of the Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", name, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, err)
	}

	if _, err := client.Get(ctx, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, name); err != nil {
		return fmt.Errorf("retrieving Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", name, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName, err)
	}

	d.SetId(id)

	return resourceResourceMoverMoveResourceResourceGroupRead(d, meta)
}

func resourceResourceMoverMoveResourceResourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ResourceMover.MoveResourceClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceMoverMoveResourceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.MoveCollectionName, id.MoveResourceName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Resource Mover Move Resource %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", id.MoveResourceName, id.ResourceGroup, id.MoveCollectionName, err)
	}
	d.Set("name", id.MoveResourceName)
	d.Set("move_collection_id", parse.NewResourceMoverMoveCollectionID(subscriptionId, id.ResourceGroup, id.MoveCollectionName).ID())
	if props := resp.Properties; props != nil {
		d.Set("existing_target_id", props.ExistingTargetID)

		resgpSetting, ok := props.ResourceSettings.AsResourceGroupResourceSettings()
		if !ok {
			return fmt.Errorf("resource Mover Move Resource %q is not type `azurerm_resource_mover_move_resource_resource_group`", d.Id())
		}
		if err := d.Set("resource_setting", flattenArmMoveResourceResourceGroupResourceSettings(resgpSetting)); err != nil {
			return fmt.Errorf("setting `resource_setting`: %+v", err)
		}
		d.Set("source_id", props.SourceID)
		if err := d.Set("error", flattenArmMoveResourceMoveResourcePropertiesError(props.Errors)); err != nil {
			return fmt.Errorf("setting `error`: %+v", err)
		}
		if err := d.Set("move_status", flattenArmMoveResourceMoveResourcePropertiesMoveStatus(props.MoveStatus)); err != nil {
			return fmt.Errorf("setting `move_status`: %+v", err)
		}
		d.Set("target_id", props.TargetID)
	}
	return nil
}

func resourceResourceMoverMoveResourceResourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceMover.MoveResourceClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceMoverMoveResourceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.MoveCollectionName, id.MoveResourceName)
	if err != nil {
		return fmt.Errorf("deleting Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", id.MoveResourceName, id.ResourceGroup, id.MoveCollectionName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", id.MoveResourceName, id.ResourceGroup, id.MoveCollectionName, err)
	}
	return nil
}

func expandArmMoveResourceResourceGroupResourceSettings(input []interface{}) resourcemover.BasicResourceSettings {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return resourcemover.ResourceGroupResourceSettings{
		ResourceType:       resourcemover.ResourceTypeResourceGroups,
		TargetResourceName: utils.String(v["target_resource_name"].(string)),
	}
}

func flattenArmMoveResourceResourceGroupResourceSettings(input *resourcemover.ResourceGroupResourceSettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var targetResourceName string
	if input.TargetResourceName != nil {
		targetResourceName = *input.TargetResourceName
	}
	return []interface{}{
		map[string]interface{}{
			"target_resource_name": targetResourceName,
		},
	}
}

func flattenArmMoveResourceMoveResourcePropertiesError(input *resourcemover.MoveResourcePropertiesErrors) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	props := input.Properties
	var code string
	if props.Code != nil {
		code = *props.Code
	}
	var message string
	if props.Message != nil {
		message = *props.Message
	}
	var target string
	if props.Target != nil {
		target = *props.Target
	}
	return []interface{}{
		map[string]interface{}{
			"code":    code,
			"message": message,
			"target":  target,
		},
	}
}

func flattenArmMoveResourceMoveResourceError(input *resourcemover.MoveResourceError) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	props := input.Properties
	var code string
	if props.Code != nil {
		code = *props.Code
	}
	var message string
	if props.Message != nil {
		message = *props.Message
	}
	var target string
	if props.Target != nil {
		target = *props.Target
	}
	return []interface{}{
		map[string]interface{}{
			"code":    code,
			"message": message,
			"target":  target,
		},
	}
}

func flattenArmMoveResourceMoveResourcePropertiesMoveStatus(input *resourcemover.MoveResourcePropertiesMoveStatus) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var moveState resourcemover.MoveState
	if input.MoveState != "" {
		moveState = input.MoveState
	}
	return []interface{}{
		map[string]interface{}{
			"error":      flattenArmMoveResourceMoveResourceError(input.Errors),
			"job_status": flattenArmMoveResourceJobStatus(input.JobStatus),
			"move_state": moveState,
		},
	}
}

//func flattenArmMoveResourceMoveResourceErrorBodyArray(input *[]resourcemover.MoveResourceErrorBody) []interface{} {
//	results := make([]interface{}, 0)
//	if input == nil {
//		return results
//	}
//
//	for _, item := range *input {
//		results = append(results, map[string]interface{}{
//			"detail": item.Details,
//		})
//	}
//	return results
//}

func flattenArmMoveResourceJobStatus(input *resourcemover.JobStatus) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var jobName resourcemover.JobName
	if input.JobName != "" {
		jobName = input.JobName
	}
	var jobProgress string
	if input.JobProgress != nil {
		jobProgress = *input.JobProgress
	}
	return []interface{}{
		map[string]interface{}{
			"job_name":     jobName,
			"job_progress": jobProgress,
		},
	}
}
