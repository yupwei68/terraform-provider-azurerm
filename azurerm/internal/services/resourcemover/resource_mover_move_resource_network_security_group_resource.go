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

func resourceResourceMoverMoveResourceNetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceMoverMoveResourceNetworkSecurityGroupCreateUpdate,
		Read:   resourceResourceMoverMoveResourceNetworkSecurityGroupRead,
		Update: resourceResourceMoverMoveResourceNetworkSecurityGroupCreateUpdate,
		Delete: resourceResourceMoverMoveResourceNetworkSecurityGroupDelete,

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

						"security_rule": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"access": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"description": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"destination_address_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"destination_port_range": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"direction": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"priority": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"protocol": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"source_address_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"source_port_range": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
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

func resourceResourceMoverMoveResourceNetworkSecurityGroupCreateUpdate(d *schema.ResourceData, meta interface{}) error {
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
			return tf.ImportAsExistsError("azurerm_resource_mover_move_resource_network_security_group", id)
		}
	}

	properties := resourcemover.MoveResource{
		Properties: &resourcemover.MoveResourceProperties{
			ResourceSettings:   expandArmMoveResourceNetworkSecurityGroupResourceSettings(d.Get("resource_setting").([]interface{})),
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

	return resourceResourceMoverMoveResourceNetworkSecurityGroupRead(d, meta)
}

func resourceResourceMoverMoveResourceNetworkSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
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
		nsgSetting, ok := props.ResourceSettings.AsNetworkSecurityGroupResourceSettings()
		if !ok {
			return fmt.Errorf("resource Mover Move Resource %q is not type `azurerm_resource_mover_move_resource_network_security_group`", d.Id())
		}

		if err := d.Set("resource_setting", flattenArmMoveResourceNetworkSecurityGroupResourceSettings(nsgSetting)); err != nil {
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

func resourceResourceMoverMoveResourceNetworkSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
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

func expandArmMoveResourceNetworkSecurityGroupResourceSettings(input []interface{}) resourcemover.BasicResourceSettings {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return resourcemover.NetworkSecurityGroupResourceSettings{
		ResourceType:       resourcemover.ResourceTypeMicrosoftNetworknetworkSecurityGroups,
		TargetResourceName: utils.String(v["target_resource_name"].(string)),
		SecurityRules:      expandArmMoveResourceNsgSecurityRuleArray(v["security_rule"].(*schema.Set).List()),
	}
}

func flattenArmMoveResourceNetworkSecurityGroupResourceSettings(input *resourcemover.NetworkSecurityGroupResourceSettings) []interface{} {
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
			"security_rule":        flattenArmMoveResourceNsgSecurityRuleArray(input.SecurityRules),
		},
	}
}

func expandArmMoveResourceNsgSecurityRuleArray(input []interface{}) *[]resourcemover.NsgSecurityRule {
	results := make([]resourcemover.NsgSecurityRule, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		results = append(results, resourcemover.NsgSecurityRule{
			Name:                     utils.String(v["name"].(string)),
			Access:                   utils.String(v["access"].(string)),
			Description:              utils.String(v["description"].(string)),
			DestinationAddressPrefix: utils.String(v["destination_address_prefix"].(string)),
			DestinationPortRange:     utils.String(v["destination_port_range"].(string)),
			Direction:                utils.String(v["direction"].(string)),
			Priority:                 utils.Int32(int32(v["priority"].(int))),
			Protocol:                 utils.String(v["protocol"].(string)),
			SourceAddressPrefix:      utils.String(v["source_address_prefix"].(string)),
			SourcePortRange:          utils.String(v["source_port_range"].(string)),
		})
	}
	return &results
}

func flattenArmMoveResourceNsgSecurityRuleArray(input *[]resourcemover.NsgSecurityRule) []interface{} {
	results := make([]interface{}, 0)
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}
		var access string
		if item.Access != nil {
			access = *item.Access
		}
		var description string
		if item.Description != nil {
			description = *item.Description
		}
		var destinationAddressPrefix string
		if item.DestinationAddressPrefix != nil {
			destinationAddressPrefix = *item.DestinationAddressPrefix
		}
		var destinationPortRange string
		if item.DestinationPortRange != nil {
			destinationPortRange = *item.DestinationPortRange
		}
		var direction string
		if item.Direction != nil {
			direction = *item.Direction
		}
		var priority int32
		if item.Priority != nil {
			priority = *item.Priority
		}
		var protocol string
		if item.Protocol != nil {
			protocol = *item.Protocol
		}
		var sourceAddressPrefix string
		if item.SourceAddressPrefix != nil {
			sourceAddressPrefix = *item.SourceAddressPrefix
		}
		var sourcePortRange string
		if item.SourcePortRange != nil {
			sourcePortRange = *item.SourcePortRange
		}
		results = append(results, map[string]interface{}{
			"name":                       name,
			"access":                     access,
			"description":                description,
			"destination_address_prefix": destinationAddressPrefix,
			"destination_port_range":     destinationPortRange,
			"direction":                  direction,
			"priority":                   priority,
			"protocol":                   protocol,
			"source_address_prefix":      sourceAddressPrefix,
			"source_port_range":          sourcePortRange,
		})
	}
	return results
}
