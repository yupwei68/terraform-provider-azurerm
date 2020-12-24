package resourcemover

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"time"
)

func dataSourceResourceMoverMoveCollection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmResourceMoverMoveCollectionRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ResourceMoverMoveCollectionName,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"source_region": azure.SchemaLocationForDataSource(),

			"target_region": azure.SchemaLocationForDataSource(),

			"identity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"principal_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"tenant_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmResourceMoverMoveCollectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceMover.MoveCollectionClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewResourceMoverMoveCollectionID(subscriptionId, resourceGroup, name).ID()

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("resource Mover Move Collection %q does not exist", name)
		}
		return fmt.Errorf("retrieving Resource Mover Move Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if props := resp.Properties; props != nil {
		d.Set("source_region", props.SourceRegion)
		d.Set("target_region", props.TargetRegion)
	}
	if err := d.Set("identity", flattenArmMoveCollectionIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	d.SetId(id)

	return tags.FlattenAndSet(d, resp.Tags)
}
