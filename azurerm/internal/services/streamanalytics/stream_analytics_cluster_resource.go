package streamanalytics

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/streamanalytics/mgmt/2020-03-01-preview/streamanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/streamanalytics/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"time"
)

func resourceStreamAnalyticsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceStreamAnalyticsClusterCreate,
		Read:   resourceStreamAnalyticsClusterRead,
		Update: resourceStreamAnalyticsClusterUpdate,
		Delete: resourceStreamAnalyticsClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ClusterID(id)
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

			"sku": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(streamanalytics.Default),
							}, false),
						},

						"capacity": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(36, 216),
						},
					},
				},
			},

			"capacity_allocated": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"capacity_assigned": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}
func resourceStreamAnalyticsClusterCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).StreamAnalytics.ClustersClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewClusterID(subscriptionId, resourceGroup, name).ID()

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Streamanalytics Cluster %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_stream_analytics_cluster", id)
	}

	cluster := streamanalytics.Cluster{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		Sku:      expandArmClusterClusterSku(d.Get("sku").([]interface{})),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
	}
	future, err := client.CreateOrUpdate(ctx, cluster, resourceGroup, name, "", "")
	if err != nil {
		return fmt.Errorf("creating Streamanalytics Cluster %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creating future for Streamanalytics Cluster %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id)
	return resourceStreamAnalyticsClusterRead(d, meta)
}

func resourceStreamAnalyticsClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.ClustersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] streamanalytics %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if err := d.Set("sku", flattenArmClusterClusterSku(resp.Sku)); err != nil {
		return fmt.Errorf("setting `sku`: %+v", err)
	}
	if props := resp.Properties; props != nil {
		d.Set("capacity_allocated", props.CapacityAllocated)
		d.Set("capacity_assigned", props.CapacityAssigned)
		d.Set("cluster_id", props.ClusterID)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceStreamAnalyticsClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.ClustersClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	cluster := streamanalytics.Cluster{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
	}
	if d.HasChange("name") {
		cluster.Name = utils.String(d.Get("name").(string))
	}
	if d.HasChange("sku") {
		cluster.Sku = expandArmClusterClusterSku(d.Get("sku").([]interface{}))
	}
	if d.HasChange("tags") {
		cluster.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	future, err := client.Update(ctx, cluster, id.ResourceGroup, id.Name, "")
	if err != nil {
		return fmt.Errorf("updating Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for updating future for Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return resourceStreamAnalyticsClusterRead(d, meta)
}

func resourceStreamAnalyticsClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).StreamAnalytics.ClustersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deleting future for Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return nil
}

func expandArmClusterClusterSku(input []interface{}) *streamanalytics.ClusterSku {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &streamanalytics.ClusterSku{
		Name:     streamanalytics.ClusterSkuName(v["name"].(string)),
		Capacity: utils.Int32(int32(v["capacity"].(int))),
	}
}

func flattenArmClusterClusterSku(input *streamanalytics.ClusterSku) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var name streamanalytics.ClusterSkuName
	if input.Name != "" {
		name = input.Name
	}
	var capacity int32
	if input.Capacity != nil {
		capacity = *input.Capacity
	}
	return []interface{}{
		map[string]interface{}{
			"name":     name,
			"capacity": capacity,
		},
	}
}
