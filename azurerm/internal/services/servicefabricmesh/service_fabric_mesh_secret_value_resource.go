package servicefabricmesh

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/servicefabricmesh/mgmt/2018-09-01-preview/servicefabricmesh"
	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/servicefabricmesh/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmServiceFabricMeshSecretValue() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmServiceFabricMeshSecretValueCreateUpdate,
		Read:   resourceArmServiceFabricMeshSecretValueRead,
		Update: resourceArmServiceFabricMeshSecretValueCreateUpdate,
		Delete: resourceArmServiceFabricMeshSecretValueDelete,
		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ServiceFabricMeshSecretValueID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// Follow casing issue here https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validation.StringIsNotEmpty,
			},

			"service_fabric_mesh_secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// Follow casing issue here https://github.com/Azure/azure-rest-api-specs/issues/9330
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     azure.ValidateResourceID,
			},

			"location": azure.SchemaLocation(),

			"value": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"tags": tags.Schema(),
		},
	}
}

func resourceArmServiceFabricMeshSecretValueCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	location := location.Normalize(d.Get("location").(string))
	t := d.Get("tags").(map[string]interface{})

	secretID, err := parse.ServiceFabricMeshSecretID(d.Get("service_fabric_mesh_secret_id").(string))
	if err != nil {
		return err
	}

	if d.IsNewResource() {
		existing, err := client.Get(ctx, secretID.ResourceGroup, secretID.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Service Fabric Mesh Secret Value: %+v", err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_service_fabric_mesh_secret_value", *existing.ID)
		}
	}

	parameters := servicefabricmesh.SecretValueResourceDescription{
		SecretValueResourceProperties: &servicefabricmesh.SecretValueResourceProperties{
			Value: utils.String(d.Get("value").(string)),
		},
		Location: utils.String(location),
		Tags:     tags.Expand(t),
	}

	if _, err := client.Create(ctx, secretID.ResourceGroup, secretID.Name, name, parameters); err != nil {
		return fmt.Errorf("creating Service Fabric Mesh Secret Value %q (Resource Group %q / Secret %q): %+v", name, secretID.ResourceGroup, secretID.Name, err)
	}

	resp, err := client.Get(ctx, secretID.ResourceGroup, secretID.Name, name)
	if err != nil {
		return fmt.Errorf("retrieving Service Fabric Mesh Secret Value %q (Resource Group %q / Secret %q): %+v", name, secretID.ResourceGroup, secretID.Name, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("client returned a nil ID for Service Fabric Mesh Secret Value %q", name)
	}

	d.SetId(*resp.ID)

	return resourceArmServiceFabricMeshSecretValueRead(d, meta)
}

func resourceArmServiceFabricMeshSecretValueRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	secretClient := meta.(*clients.Client).ServiceFabricMesh.SecretClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceFabricMeshSecretValueID(d.Id())
	if err != nil {
		return err
	}

	secret, err := secretClient.Get(ctx, id.ResourceGroup, id.SecretName)
	if err != nil {
		if utils.ResponseWasNotFound(secret.Response) {
			log.Printf("[INFO] Unable to find Service Fabric Mesh Secret %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Service Fabric Mesh Secret: %+v", err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.SecretName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Unable to find Service Fabric Mesh Secret Value %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading Service Fabric Mesh Secret Value: %+v", err)
	}

	d.Set("name", id.Name)
	d.Set("service_fabric_mesh_secret_id", secret.ID)
	d.Set("location", location.NormalizeNilable(resp.Location))

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmServiceFabricMeshSecretValueDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ServiceFabricMesh.SecretValueClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ServiceFabricMeshSecretValueID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Delete(ctx, id.ResourceGroup, id.SecretName, id.Name)
	if err != nil {
		if !response.WasNotFound(resp.Response) {
			return fmt.Errorf("deleting Service Fabric Mesh Secret Value %q (Resource Group %q / Secret %q): %+v", id.Name, id.ResourceGroup, id.SecretName, err)
		}
	}

	return nil
}
