package resourcemover

import (
	"bytes"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/resourcemover/mgmt/2019-10-01-preview/resourcemover"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
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

func resourceResourceMoverMoveResourceVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceMoverMoveResourceVirtualNetworkCreateUpdate,
		Read:   resourceResourceMoverMoveResourceVirtualNetworkRead,
		Update: resourceResourceMoverMoveResourceVirtualNetworkCreateUpdate,
		Delete: resourceResourceMoverMoveResourceVirtualNetworkDelete,

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

						"address_spaces": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"dns_servers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},

						"enable_ddos_protection": {
							Type:     schema.TypeBool,
							Optional: true,
						},

						"subnet": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"address_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							Set: resourceMoverMoveResourceSubnetHash,
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

func resourceResourceMoverMoveResourceVirtualNetworkCreateUpdate(d *schema.ResourceData, meta interface{}) error {
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
			return tf.ImportAsExistsError("azurerm_resource_mover_move_resource_virtual_network", id)
		}
	}

	properties := resourcemover.MoveResource{
		Properties: &resourcemover.MoveResourceProperties{
			ResourceSettings:   expandArmMoveResourceVirtualNetworkResourceSettings(d.Get("resource_setting").([]interface{})),
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

	return resourceResourceMoverMoveResourceVirtualNetworkRead(d, meta)
}

func resourceResourceMoverMoveResourceVirtualNetworkRead(d *schema.ResourceData, meta interface{}) error {
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
		vnSetting, ok := props.ResourceSettings.AsVirtualNetworkResourceSettings()
		if !ok {
			return fmt.Errorf("resource Mover Move Resource %q is not type `azurerm_resource_mover_move_resource_virtual_network`", d.Id())
		}

		if err := d.Set("resource_setting", flattenArmMoveResourceVirtualNetworkResourceSettings(vnSetting)); err != nil {
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

func resourceResourceMoverMoveResourceVirtualNetworkDelete(d *schema.ResourceData, meta interface{}) error {
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

func expandArmMoveResourceVirtualNetworkResourceSettings(input []interface{}) resourcemover.BasicResourceSettings {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return resourcemover.VirtualNetworkResourceSettings{
		ResourceType:         resourcemover.ResourceTypeMicrosoftNetworkvirtualNetworks,
		TargetResourceName:   utils.String(v["target_resource_name"].(string)),
		EnableDdosProtection: utils.Bool(v["enable_ddos_protection"].(bool)),
		AddressSpace:         utils.ExpandStringSlice(v["address_spaces"].(*schema.Set).List()),
		DNSServers:           utils.ExpandStringSlice(v["dns_servers"].(*schema.Set).List()),
		Subnets:              expandArmMoveResourceSubnetResourceSettingsArray(v["subnet"].(*schema.Set).List()),
	}
}

func flattenArmMoveResourceVirtualNetworkResourceSettings(input *resourcemover.VirtualNetworkResourceSettings) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var targetResourceName string
	if input.TargetResourceName != nil {
		targetResourceName = *input.TargetResourceName
	}
	var enableDdosProtection bool
	if input.EnableDdosProtection != nil {
		enableDdosProtection = *input.EnableDdosProtection
	}
	return []interface{}{
		map[string]interface{}{
			"target_resource_name":   targetResourceName,
			"address_spaces":         utils.FlattenStringSlice(input.AddressSpace),
			"dns_servers":            utils.FlattenStringSlice(input.DNSServers),
			"enable_ddos_protection": enableDdosProtection,
			"subnet":                 flattenArmMoveResourceSubnetResourceSettingsArray(input.Subnets),
		},
	}
}

func flattenArmMoveResourceSubnetResourceSettingsArray(input *[]resourcemover.SubnetResourceSettings) *schema.Set {
	results := &schema.Set{
		F: resourceMoverMoveResourceSubnetHash,
	}
	if input == nil {
		return results
	}

	for _, item := range *input {
		var name string
		if item.Name != nil {
			name = *item.Name
		}
		var addressPrefix string
		if item.AddressPrefix != nil {
			addressPrefix = *item.AddressPrefix
		}
		results.Add(map[string]interface{}{
			"name":           name,
			"address_prefix": addressPrefix,
		})
	}
	return results
}

func expandArmMoveResourceSubnetResourceSettingsArray(input []interface{}) *[]resourcemover.SubnetResourceSettings {
	results := make([]resourcemover.SubnetResourceSettings, 0)
	for _, item := range input {
		v := item.(map[string]interface{})
		results = append(results, resourcemover.SubnetResourceSettings{
			Name:          utils.String(v["name"].(string)),
			AddressPrefix: utils.String(v["address_prefix"].(string)),
		})
	}
	return &results
}

func resourceMoverMoveResourceSubnetHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		if v, ok := m["name"]; ok {
			buf.WriteString(v.(string))
		}
		if v, ok := m["address_prefix"]; ok {
			buf.WriteString(v.(string))
		}
	}

	return hashcode.String(buf.String())
}
