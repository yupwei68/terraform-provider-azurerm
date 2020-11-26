package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/advisor/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type AdvisorSuppressionResource struct{}

func TestAccAzureRMAdvisorSuppression_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_advisor_suppression", "test")
	r := AdvisorSuppressionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMAdvisorSuppression_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_advisor_suppression", "test")
	r := AdvisorSuppressionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})

}

func TestAccAzureRMAdvisorSuppression_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_advisor_suppression", "test")
	r := AdvisorSuppressionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccAzureRMAdvisorSuppression_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_advisor_suppression", "test")
	r := AdvisorSuppressionResource{}

	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r AdvisorSuppressionResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.AdvisorSuppressionID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := client.Advisor.SuppressionsClient.Get(ctx, id.ResourceUri, id.RecommendationName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Advisor Suppression %q (Recommendation Name: %q, Resource Uri:%q): %+v", id.Name, id.RecommendationName, id.ResourceUri, err)
	}

	return utils.Bool(true), nil
}

// we have a recommendation to create a service health alert that is available on all subscriptions that don’t have such an alert created already.
func (r AdvisorSuppressionResource) template() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_advisor_recommendations" "test" {}

locals{
 rec_id = [for x in data.azurerm_advisor_recommendations.test.recommendations: x.recommendation_id if x.recommendation_type_id == "c6ac1f03-bd58-4421-9522-23cffb64d8e1"]
}
`)
}

func (r AdvisorSuppressionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_advisor_suppression" "test" {
  name              = "acctest-sp-%d"
  recommendation_id = local.rec_id[0]
}
`, r.template(), data.RandomInteger)
}

func (r AdvisorSuppressionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_advisor_suppression" "import" {
  name              = azurerm_advisor_suppression.test.name
  recommendation_id = azurerm_advisor_suppression.test.recommendation_id
}
`, r.basic(data))
}

func (r AdvisorSuppressionResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_advisor_suppression" "test" {
  name              = "acctest-sp-%d"
  recommendation_id = local.rec_id[0]
  duration_in_days  = 1
}
`, r.template(), data.RandomInteger)
}

func (r AdvisorSuppressionResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_advisor_suppression" "test" {
  name              = "acctest-sp-%d"
  recommendation_id = local.rec_id[0]
  duration_in_days  = 2
}
`, r.template(), data.RandomInteger)
}
