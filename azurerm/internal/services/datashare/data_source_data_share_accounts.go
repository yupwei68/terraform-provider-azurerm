package datashare

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceDataShareAccount() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmDataShareAccountRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmDataShareAccountRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).DataShare.AccountClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] DataShare %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("failure in retrieving DataShare Account %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if id := resp.ID; id != nil {
		d.SetId(*resp.ID)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	return tags.FlattenAndSet(d, resp.Tags)
}
