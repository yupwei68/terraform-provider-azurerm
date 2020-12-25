package resourcemover

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/resourcemover/mgmt/2019-10-01-preview/resourcemover"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"time"
)

func resourceResourceMoverMoveCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceMoverMoveCollectionCreate,
		Read:   resourceResourceMoverMoveCollectionRead,
		Update: resourceResourceMoverMoveCollectionUpdate,
		Delete: resourceResourceMoverMoveCollectionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ResourceMoverMoveCollectionID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ResourceMoverMoveCollectionName,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"location": azure.SchemaLocation(),

			"source_region": azure.SchemaResourceGroupName(),

			"target_region": azure.SchemaResourceGroupName(),

			"identity": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"principal_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsUUID,
						},

						"tenant_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsUUID,
						},

						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(resourcemover.SystemAssigned),
								string(resourcemover.UserAssigned),
							}, false),
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}
func resourceResourceMoverMoveCollectionCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ResourceMover.MoveCollectionClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewResourceMoverMoveCollectionID(subscriptionId, resourceGroup, name).ID()

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Resource Mover Move Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}

	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_resource_mover_move_collection", id)
	}

	properties := resourcemover.MoveCollection{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		Properties: &resourcemover.MoveCollectionProperties{
			SourceRegion: utils.String(d.Get("source_region").(string)),
			TargetRegion: utils.String(d.Get("target_region").(string)),
		},
		Identity: expandArmMoveCollectionIdentity(d.Get("identity").([]interface{})),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if _, err := client.Create(ctx, resourceGroup, name, &properties); err != nil {
		return fmt.Errorf("creating Resource Mover Move Collection %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id)

	return resourceResourceMoverMoveCollectionRead(d, meta)
}

func resourceResourceMoverMoveCollectionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceMover.MoveCollectionClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceMoverMoveCollectionID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.MoveCollectionName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Resource Mover Move Collection %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Resource Mover Move Collection %q (Resource Group %q): %+v", id.MoveCollectionName, id.ResourceGroup, err)
	}
	d.Set("name", id.MoveCollectionName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if err := d.Set("identity", flattenArmMoveCollectionIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}
	if props := resp.Properties; props != nil {
		d.Set("source_region", props.SourceRegion)
		d.Set("target_region", props.TargetRegion)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceResourceMoverMoveCollectionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceMover.MoveCollectionClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceMoverMoveCollectionID(d.Id())
	if err != nil {
		return err
	}

	properties := resourcemover.UpdateMoveCollectionRequest{}

	if d.HasChange("identity") {
		properties.Identity = expandArmMoveCollectionIdentity(d.Get("identity").([]interface{}))
	}

	if d.HasChange("tags") {
		properties.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.MoveCollectionName, &properties); err != nil {
		return fmt.Errorf("updating Resource Mover Move Collection %q (Resource Group %q): %+v", id.MoveCollectionName, id.ResourceGroup, err)
	}

	return resourceResourceMoverMoveCollectionRead(d, meta)
}

func resourceResourceMoverMoveCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceMover.MoveCollectionClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceMoverMoveCollectionID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.MoveCollectionName)
	if err != nil {
		return fmt.Errorf("deleting Resource Mover Move Collection %q (Resource Group %q): %+v", id.MoveCollectionName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deletion of the Resource Mover Move Collection %q (Resource Group %q): %+v", id.MoveCollectionName, id.ResourceGroup, err)
	}
	return nil
}

func expandArmMoveCollectionIdentity(input []interface{}) *resourcemover.Identity {
	if len(input) == 0 {
		return &resourcemover.Identity{
			Type: resourcemover.None,
		}
	}
	v := input[0].(map[string]interface{})

	identity := resourcemover.Identity{
		Type: resourcemover.ResourceIdentityType(v["type"].(string)),
	}

	if p, ok := v["principal_id"]; ok && p.(string) != "" {
		identity.PrincipalID = utils.String(p.(string))
	}

	if t, ok := v["tenant_id"]; ok && t.(string) != "" {
		identity.TenantID = utils.String(t.(string))
	}

	return &identity
}

func flattenArmMoveCollectionIdentity(input *resourcemover.Identity) []interface{} {
	if input == nil || input.Type == resourcemover.None {
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
			"principal_id": principalId,
			"tenant_id":    tenantId,
			"type":         input.Type,
		},
	}
}
