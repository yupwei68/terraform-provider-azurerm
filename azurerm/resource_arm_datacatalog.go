package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/datacatalog/mgmt/2016-03-30/datacatalog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmDatacatalog() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmDatacatalogCreateUpdate,
		Read:   resourceArmDatacatalogRead,
		Update: resourceArmDatacatalogCreateUpdate,
		Delete: resourceArmDatacatalogDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku": {
				Type:     schema.TypeString,
				Required: true,
				//ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(datacatalog.Free),
					string(datacatalog.Standard),
				}, false),
			},

			"units": {
				Type:     schema.TypeInt,
				Optional: true,
				// todo ValidateFunc: validate.NoEmptyStrings,
			},

			"enable_automatic_unit_adjustment": {
				Type:     schema.TypeBool,
				Optional: true,
				// default? todo
			},

			"admin": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"upn": {
							Type:     schema.TypeString,
							Optional: true,
							//todo validate email
						},
						"object_id": {
							Type:     schema.TypeString,
							Optional: true,
							//todo validate gUID
						},
					},
				},
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmDatacatalogCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	ctx := meta.(*ArmClient).StopContext

	log.Printf("[INFO] preparing arguments for Azure Data Catalog creation.")

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	client := meta.(*ArmClient).dataCatalog.CatalogsClient(name)

	if features.ShouldResourcesBeImported() && d.IsNewResource() {
		existing, err := client.Get(ctx, resourceGroup)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Azure Data Catalog %q (Resource Group %q): %s", name, resourceGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_datacatalog", *existing.ID)
		}
	}

	datacatalog := datacatalog.ADCCatalog{
		Location: utils.String(azure.NormalizeLocation(d.Get("location").(string))),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
		ADCCatalogProperties: &datacatalog.ADCCatalogProperties{
			Sku: datacatalog.SkuType(d.Get("sku").(string)),
			Admins: &[]datacatalog.Principals{
				{
					Upn: utils.String(d.Get("admin.0.upn").(string)),
				},
			},
		},
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, datacatalog); err != nil {
		return err
	}

	read, err := client.Get(ctx, resourceGroup)
	if err != nil {
		return err
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read Azure Data Catalog %q (resource group %s) ID", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmDatacatalogRead(d, meta)
}

func resourceArmDatacatalogRead(d *schema.ResourceData, meta interface{}) error {
	ctx := meta.(*ArmClient).StopContext

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["catalogs"]

	client := meta.(*ArmClient).dataCatalog.CatalogsClient(name)

	resp, err := client.Get(ctx, resourceGroup)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on Azure CDN Profile %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.ADCCatalogProperties; props != nil {
		d.Set("sku", string(props.Sku))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmDatacatalogDelete(d *schema.ResourceData, meta interface{}) error {
	ctx := meta.(*ArmClient).StopContext

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resourceGroup := id.ResourceGroup
	name := id.Path["catalogs"]

	client := meta.(*ArmClient).dataCatalog.CatalogsClient(name)

	future, err := client.Delete(ctx, resourceGroup)
	if err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error issuing delete request for Azure CDN Profile %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if response.WasNotFound(future.Response()) {
			return nil
		}
		return fmt.Errorf("Error waiting for Azure CDN Profile %q (Resource Group %q) to be deleted: %+v", name, resourceGroup, err)
	}

	return err
}
