package confluent

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/confluent/mgmt/2020-03-01-preview/confluent"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/location"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/confluent/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/confluent/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmConfluentOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmConfluentOrganizationCreate,
		Read:   resourceArmConfluentOrganizationRead,
		Update: resourceArmConfluentOrganizationUpdate,
		Delete: resourceArmConfluentOrganizationDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ConfluentOrganizationID(id)
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

			"offer_detail": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"plan_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationPlanID,
						},

						"plan_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationPlanName,
						},

						"publisher_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationPublisherID,
						},

						"term_unit": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationTermUnit,
						},
					},
				},
			},

			"user_detail": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationEmailAddress,
						},

						"first_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationFirstName,
						},

						"last_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validate.ConfluentOrganizationLastName,
						},
					},
				},
			},

			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sso_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tags.Schema(),
		},
	}
}
func resourceArmConfluentOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Confluent.OrganizationClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	id := parse.NewConfluentOrganizationID(subscriptionId, resourceGroup, name).ID()

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Confluent Organization %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_confluent_organization", id)
	}

	properties := confluent.OrganizationResource{
		Location: utils.String(location.Normalize(d.Get("location").(string))),
		OrganizationResourcePropertiesModel: &confluent.OrganizationResourcePropertiesModel{
			OfferDetail: expandArmOrganizationOrganizationResourcePropertiesOfferDetail(d.Get("offer_detail").([]interface{})),
			UserDetail:  expandArmOrganizationOrganizationResourcePropertiesUserDetail(d.Get("user_detail").([]interface{})),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	future, err := client.Create(ctx, resourceGroup, name, &properties)
	if err != nil {
		return fmt.Errorf("creating Confluent Organization %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of the Confluent Organization %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if _, err := client.Get(ctx, resourceGroup, name); err != nil {
		return fmt.Errorf("retrieving Confluent Organization %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(id)

	return resourceArmConfluentOrganizationRead(d, meta)
}

func resourceArmConfluentOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Confluent.OrganizationClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ConfluentOrganizationID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.OrganizationName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] confluent %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Confluent Organization %q (Resource Group %q): %+v", id.OrganizationName, id.ResourceGroup, err)
	}
	d.Set("name", id.OrganizationName)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", location.NormalizeNilable(resp.Location))
	if props := resp.OrganizationResourcePropertiesModel; props != nil {
		if err := d.Set("offer_detail", flattenArmOrganizationOrganizationResourcePropertiesOfferDetail(props.OfferDetail)); err != nil {
			return fmt.Errorf("setting `offer_detail`: %+v", err)
		}
		if err := d.Set("user_detail", flattenArmOrganizationOrganizationResourcePropertiesUserDetail(props.UserDetail)); err != nil {
			return fmt.Errorf("setting `user_detail`: %+v", err)
		}
		d.Set("organization_id", props.OrganizationID)
		d.Set("sso_url", props.SsoURL)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmConfluentOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Confluent.OrganizationClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ConfluentOrganizationID(d.Id())
	if err != nil {
		return err
	}

	properties := confluent.OrganizationResourceUpdate{}
	if d.HasChange("tags") {
		properties.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.OrganizationName, &properties); err != nil {
		return fmt.Errorf("updating Confluent Organization %q (Resource Group %q): %+v", id.OrganizationName, id.ResourceGroup, err)
	}
	return resourceArmConfluentOrganizationRead(d, meta)
}

func resourceArmConfluentOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Confluent.OrganizationClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ConfluentOrganizationID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.OrganizationName)
	if err != nil {
		return fmt.Errorf("deleting Confluent Organization %q (Resource Group %q): %+v", id.OrganizationName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deleting future for Confluent Organization %q (Resource Group %q): %+v", id.OrganizationName, id.ResourceGroup, err)
	}
	return nil
}

func expandArmOrganizationOrganizationResourcePropertiesOfferDetail(input []interface{}) *confluent.OrganizationResourcePropertiesOfferDetail {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &confluent.OrganizationResourcePropertiesOfferDetail{
		PublisherID: utils.String(v["publisher_id"].(string)),
		PlanID:      utils.String(v["plan_id"].(string)),
		ID:          utils.String(v["id"].(string)),
		PlanName:    utils.String(v["plan_name"].(string)),
		TermUnit:    utils.String(v["term_unit"].(string)),
	}
}

func expandArmOrganizationOrganizationResourcePropertiesUserDetail(input []interface{}) *confluent.OrganizationResourcePropertiesUserDetail {
	if len(input) == 0 {
		return nil
	}
	v := input[0].(map[string]interface{})
	return &confluent.OrganizationResourcePropertiesUserDetail{
		FirstName:    utils.String(v["first_name"].(string)),
		LastName:     utils.String(v["last_name"].(string)),
		EmailAddress: utils.String(v["email_address"].(string)),
	}
}

func flattenArmOrganizationOrganizationResourcePropertiesOfferDetail(input *confluent.OrganizationResourcePropertiesOfferDetail) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var planId string
	if input.PlanID != nil {
		planId = *input.PlanID
	}
	var planName string
	if input.PlanName != nil {
		planName = *input.PlanName
	}
	var publisherId string
	if input.PublisherID != nil {
		publisherId = *input.PublisherID
	}
	var termUnit string
	if input.TermUnit != nil {
		termUnit = *input.TermUnit
	}
	var id string
	if input.ID != nil {
		id = *input.ID
	}
	return []interface{}{
		map[string]interface{}{
			"plan_id":      planId,
			"plan_name":    planName,
			"publisher_id": publisherId,
			"term_unit":    termUnit,
			"id":           id,
		},
	}
}

func flattenArmOrganizationOrganizationResourcePropertiesUserDetail(input *confluent.OrganizationResourcePropertiesUserDetail) []interface{} {
	if input == nil {
		return make([]interface{}, 0)
	}

	var emailAddress string
	if input.EmailAddress != nil {
		emailAddress = *input.EmailAddress
	}
	var firstName string
	if input.FirstName != nil {
		firstName = *input.FirstName
	}
	var lastName string
	if input.LastName != nil {
		lastName = *input.LastName
	}
	return []interface{}{
		map[string]interface{}{
			"email_address": emailAddress,
			"first_name":    firstName,
			"last_name":     lastName,
		},
	}
}
