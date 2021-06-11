package apimanagement

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2020-12-01/apimanagement"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/apimanagement/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/apimanagement/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"strings"
	"time"
)

func resourceApiManagementCache() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceApiManagementCacheCreate,
		Read:   resourceApiManagementCacheRead,
		Update: resourceApiManagementCacheUpdate,
		Delete: resourceApiManagementCacheDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.CacheID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"api_management_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ApiManagementID,
			},

			"connection_string": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"redis_cache_id": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"description": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"use_from_location": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Default:          "default",
				ValidateFunc:     validate.CacheUseFromLocation,
				DiffSuppressFunc: location.DiffSuppressFunc,
			},
		},
	}
}
func resourceApiManagementCacheCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	apimId, err := parse.ApiManagementID(d.Get("api_management_id").(string))
	if err != nil {
		return err
	}
	id := parse.NewCacheID(subscriptionId, apimId.ResourceGroup, apimId.ServiceName, name)

	existing, err := client.Get(ctx, apimId.ResourceGroup, apimId.ServiceName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for existing %q: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_api_management_cache", id.ID())
	}

	parameters := apimanagement.CacheContract{
		CacheContractProperties: &apimanagement.CacheContractProperties{
			ConnectionString: utils.String(d.Get("connection_string").(string)),
			UseFromLocation:  utils.String(location.Normalize(d.Get("use_from_location").(string))),
		},
	}

	if v, ok := d.GetOk("description"); ok && v.(string) != "" {
		parameters.CacheContractProperties.Description = utils.String(v.(string))
	}

	if v, ok := d.GetOk("redis_cache_id"); ok && v.(string) != "" {
		parameters.CacheContractProperties.ResourceID = utils.String(meta.(*clients.Client).Account.Environment.ResourceManagerEndpoint + v.(string))
	}

	if _, err := client.CreateOrUpdate(ctx, apimId.ResourceGroup, apimId.ServiceName, name, parameters, ""); err != nil {
		return fmt.Errorf("creating %q: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceApiManagementCacheRead(d, meta)
}

func resourceApiManagementCacheRead(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CacheID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ServiceName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] apimanagement %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %q: %+v", id, err)
	}
	d.Set("name", id.Name)
	d.Set("api_management_id", parse.NewApiManagementID(subscriptionId, id.ResourceGroup, id.ServiceName).ID())
	if props := resp.CacheContractProperties; props != nil {
		d.Set("description", props.Description)

		cacheId := ""
		if props.ResourceID != nil {
			cacheId = strings.TrimPrefix(*props.ResourceID, meta.(*clients.Client).Account.Environment.ResourceManagerEndpoint)
		}
		d.Set("redis_cache_id", cacheId)
		d.Set("use_from_location", props.UseFromLocation)
	}
	return nil
}

func resourceApiManagementCacheUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CacheID(d.Id())
	if err != nil {
		return err
	}

	parameters := apimanagement.CacheUpdateParameters{
		CacheUpdateProperties: &apimanagement.CacheUpdateProperties{},
	}
	if d.HasChange("description") {
		parameters.CacheUpdateProperties.Description = nil
		if v := d.Get("description").(string); v != "" {
			parameters.CacheUpdateProperties.Description = utils.String(v)
		}
	}
	if d.HasChange("connection_string") {
		parameters.CacheUpdateProperties.ConnectionString = utils.String(d.Get("connection_string").(string))
	}
	if d.HasChange("use_from_location") {
		parameters.CacheUpdateProperties.UseFromLocation = utils.String(d.Get("use_from_location").(string))
	}
	if d.HasChange("redis_cache_id") {
		parameters.CacheUpdateProperties.ResourceID = utils.String(meta.(*clients.Client).Account.Environment.ResourceManagerEndpoint + d.Get("redis_cache_id").(string))
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.ServiceName, id.Name, parameters, "*"); err != nil {
		return fmt.Errorf("updating %q: %+v", id, err)
	}
	return resourceApiManagementCacheRead(d, meta)
}

func resourceApiManagementCacheDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ApiManagement.CacheClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.CacheID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.ServiceName, id.Name, "*"); err != nil {
		return fmt.Errorf("deleting %q: %+v", id, err)
	}
	return nil
}
