package appplatform

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/appplatform/mgmt/2019-05-01-preview/appplatform"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/appplatform/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/appplatform/validate"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmSpringCloudCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSpringCloudCertificateCreate,
		Read:   resourceArmSpringCloudCertificateRead,
		Delete: resourceArmSpringCloudCertificateDelete,

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.SpringCloudCertificateID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"service_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.SpringCloudServiceName,
			},

			"key_vault_certificate_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateKeyVaultChildId,
			},
		},
	}
}

func resourceArmSpringCloudCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.CertificatesClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)
	serviceName := d.Get("service_name").(string)

	resourceId := parse.NewSpringCloudCertificateID(subscriptionId, resourceGroup, serviceName, name).ID("")
	existing, err := client.Get(ctx, resourceGroup, serviceName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Spring Cloud Service Certificate %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_spring_cloud_certificate", resourceId)
	}

	keyVaultCertificateId, _ := azure.ParseKeyVaultChildID(d.Get("key_vault_certificate_id").(string))
	cert := appplatform.CertificateResource{
		Properties: &appplatform.CertificateProperties{
			VaultURI:         &keyVaultCertificateId.KeyVaultBaseUrl,
			KeyVaultCertName: &keyVaultCertificateId.Name,
		},
	}

	if _, err := client.CreateOrUpdate(ctx, resourceGroup, serviceName, name, cert); err != nil {
		return fmt.Errorf("creating Spring Cloud Certificate %q (Spring Cloud Service %q / Resource Group %q): %+v", name, serviceName, resourceGroup, err)
	}

	d.SetId(resourceId)
	return resourceArmSpringCloudCertificateRead(d, meta)
}

func resourceArmSpringCloudCertificateRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.CertificatesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudCertificateID(d.Id())
	if err != nil {
		return err
	}
	resp, err := client.Get(ctx, id.ResourceGroup, id.SpringName, id.CertificateName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Spring Cloud Certificate %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("reading Spring Cloud Certificate %q (Spring Cloud Service %q / Resource Group %q): %+v", id.CertificateName, id.SpringName, id.ResourceGroup, err)
	}

	d.Set("name", id.CertificateName)
	d.Set("service_name", id.SpringName)
	d.Set("resource_group_name", id.ResourceGroup)

	return nil
}

func resourceArmSpringCloudCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppPlatform.CertificatesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SpringCloudCertificateID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.SpringName, id.CertificateName); err != nil {
		return fmt.Errorf("deleting Spring Cloud Certificate %q (Spring Cloud Service %q / Resource Group %q): %+v", id.CertificateName, id.SpringName, id.ResourceGroup, err)
	}

	return nil
}
