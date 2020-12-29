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

func resourceResourceMoverMoveResourcePublicIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceMoverMoveResourcePublicIPCreateUpdate,
		Read:   resourceResourceMoverMoveResourcePublicIPRead,
		Update: resourceResourceMoverMoveResourcePublicIPCreateUpdate,
		Delete: resourceResourceMoverMoveResourcePublicIPDelete,

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

						"domain_name_label": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"f_qdn": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"public_ip_allocation_method": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"sku": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"zones": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"source_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"depends_on_override": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"existing_target_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"dependency": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatic_resolution": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"move_resource_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"dependency_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"is_optional": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"manual_resolution": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"resolution_status": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"resolution_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
func resourceResourceMoverMoveResourcePublicIPCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ResourceMover.MoveResourceClient
	collectionClient := meta.(*clients.Client).ResourceMover.MoveCollectionClient
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
			return tf.ImportAsExistsError("azurerm_resource_mover_move_resource_public_ip", id)
		}
	}

	properties := resourcemover.MoveResource{
		Properties: &resourcemover.MoveResourceProperties{
			ResourceSettings:   expandArmMoveResourcePublicIPAddressResourceSettings(d.Get("resource_setting").([]interface{})),
			SourceID:           utils.String(d.Get("source_id").(string)),
			DependsOnOverrides: expandArmMoveResourceMoveResourceDependencyOverrideArray(d.Get("depends_on_override").(*schema.Set)),
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

	dependencyFuture, err := collectionClient.ResolveDependencies(ctx, moveCollectionId.ResourceGroup, moveCollectionId.MoveCollectionName)
	if err != nil {
		return fmt.Errorf("generating Resource Mover Move Collection %q Resolve Dependency (Resource Group %q ): %+v", moveCollectionId.MoveCollectionName, moveCollectionId.ResourceGroup, err)
	}

	if err := dependencyFuture.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the generation of the Resource Mover Move Collection %q Resolve Dependency (Resource Group %q ): %+v", moveCollectionId.MoveCollectionName, moveCollectionId.ResourceGroup, err)
	}

	return resourceResourceMoverMoveResourcePublicIPRead(d, meta)
}

func resourceResourceMoverMoveResourcePublicIPRead(d *schema.ResourceData, meta interface{}) error {
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
			log.Printf("[INFO] resourcemover %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Resource Mover Move Resource %q (Resource Group %q / moveCollectionName %q): %+v", id.MoveResourceName, id.ResourceGroup, id.MoveCollectionName, err)
	}
	d.Set("name", id.MoveResourceName)
	d.Set("move_collection_id", parse.NewResourceMoverMoveCollectionID(subscriptionId, id.ResourceGroup, id.MoveCollectionName).ID())
	if props := resp.Properties; props != nil {
		pipSetting, ok := props.ResourceSettings.AsPublicIPAddressResourceSettings()
		if !ok {
			return fmt.Errorf("resource Mover Move Resource %q is not type `azurerm_resource_mover_move_resource_public_ip`", d.Id())
		}

		if err := d.Set("resource_setting", flattenArmMoveResourcePublicIPAddressResourceSettings(pipSetting)); err != nil {
			return fmt.Errorf("setting `resource_setting`: %+v", err)
		}

		if err := d.Set("depends_on_override", flattenArmMoveResourceMoveResourceDependencyOverrideArray(props.DependsOnOverrides)); err != nil {
			return fmt.Errorf("setting `depends_on_override`: %+v", err)
		}
		d.Set("existing_target_id", props.ExistingTargetID)
		d.Set("source_id", props.SourceID)
		if err := d.Set("dependency", flattenArmMoveResourceMoveResourceDependencyArray(props.DependsOn)); err != nil {
			return fmt.Errorf("setting `dependency`: %+v", err)
		}
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

func resourceResourceMoverMoveResourcePublicIPDelete(d *schema.ResourceData, meta interface{}) error {
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

func expandArmMoveResourcePublicIPAddressResourceSettings(input []interface{}) resourcemover.BasicResourceSettings {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return resourcemover.PublicIPAddressResourceSettings{
		ResourceType:             resourcemover.ResourceTypeMicrosoftNetworkpublicIPAddresses,
		TargetResourceName:       utils.String(v["target_resource_name"].(string)),
		DomainNameLabel:          utils.String(v["domain_name_label"].(string)),
		FQDN:                     utils.String(v["f_qdn"].(string)),
		PublicIPAllocationMethod: utils.String(v["public_ip_allocation_method"].(string)),
		Sku:                      utils.String(v["sku"].(string)),
		Zones:                    utils.String(v["zones"].(string)),
	}
}

func expandArmMoveResourceMoveResourceDependencyOverrideArray(input *schema.Set) *[]resourcemover.MoveResourceDependencyOverride {
	results := make([]resourcemover.MoveResourceDependencyOverride, 0)
	for _, item := range input.List() {
		v := item.(map[string]interface{})
		results = append(results, resourcemover.MoveResourceDependencyOverride{
			ID:       utils.String(v["id"].(string)),
			TargetID: utils.String(v["target_id"].(string)),
		})
	}
	return &results
}

func flattenArmMoveResourceMoveResourceDependencyOverrideArray(input *[]resourcemover.MoveResourceDependencyOverride) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var id, targetId string
		if item.ID != nil {
			id = *item.ID
		}
		if item.TargetID != nil {
			targetId = *item.TargetID
		}
		results = append(results, map[string]interface{}{
			"id":        id,
			"target_id": targetId,
		})
	}
	return results
}

func flattenArmMoveResourcePublicIPAddressResourceSettings(input *resourcemover.PublicIPAddressResourceSettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var targetResourceName string
	if input.TargetResourceName != nil {
		targetResourceName = *input.TargetResourceName
	}
	var domainNameLabel string
	if input.DomainNameLabel != nil {
		domainNameLabel = *input.DomainNameLabel
	}
	var fQdn string
	if input.FQDN != nil {
		fQdn = *input.FQDN
	}
	var publicIpAllocationMethod string
	if input.PublicIPAllocationMethod != nil {
		publicIpAllocationMethod = *input.PublicIPAllocationMethod
	}
	var sku string
	if input.Sku != nil {
		sku = *input.Sku
	}
	var zones string
	if input.Zones != nil {
		zones = *input.Zones
	}
	return []interface{}{
		map[string]interface{}{
			"target_resource_name":        targetResourceName,
			"domain_name_label":           domainNameLabel,
			"f_qdn":                       fQdn,
			"public_ip_allocation_method": publicIpAllocationMethod,
			"sku":                         sku,
			"zones":                       zones,
		},
	}
}

func flattenArmMoveResourceMoveResourceDependencyArray(input *[]resourcemover.MoveResourceDependency) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var dependencyType resourcemover.DependencyType
		if item.DependencyType != "" {
			dependencyType = item.DependencyType
		}
		var isOptional string
		if item.IsOptional != nil {
			isOptional = *item.IsOptional
		}
		var resolutionStatus string
		if item.ResolutionStatus != nil {
			resolutionStatus = *item.ResolutionStatus
		}
		var resolutionType resourcemover.ResolutionType
		if item.ResolutionType != "" {
			resolutionType = item.ResolutionType
		}
		results = append(results, map[string]interface{}{
			"automatic_resolution": flattenArmMoveResourceAutomaticResolutionProperties(item.AutomaticResolution),
			"dependency_type":      dependencyType,
			"is_optional":          isOptional,
			"manual_resolution":    flattenArmMoveResourceManualResolutionProperties(item.ManualResolution),
			"resolution_status":    resolutionStatus,
			"resolution_type":      resolutionType,
		})
	}
	return results
}

func flattenArmMoveResourceAutomaticResolutionProperties(input *resourcemover.AutomaticResolutionProperties) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var moveResourceId string
	if input.MoveResourceID != nil {
		moveResourceId = *input.MoveResourceID
	}
	return []interface{}{
		map[string]interface{}{
			"move_resource_id": moveResourceId,
		},
	}
}

func flattenArmMoveResourceManualResolutionProperties(input *resourcemover.ManualResolutionProperties) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var targetId string
	if input.TargetID != nil {
		targetId = *input.TargetID
	}
	return []interface{}{
		map[string]interface{}{
			"target_id": targetId,
		},
	}
}
