package confluent_test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/confluent/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type ConfluentOrganizationResource struct {
}

func TestAccConfluentOrganization_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	if data.Client().IsServicePrincipal {
		t.Skip("Skipping due to API issue preventing authentication with service principal")
		return
	}
	r := ConfluentOrganizationResource{}

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

func TestAccConfluentOrganization_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	if data.Client().IsServicePrincipal {
		t.Skip("Skipping due to API issue preventing authentication with service principal")
		return
	}
	r := ConfluentOrganizationResource{}

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

func TestAccConfluentOrganization_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	if data.Client().IsServicePrincipal {
		t.Skip("Skipping due to API issue preventing authentication with service principal")
		return
	}
	r := ConfluentOrganizationResource{}

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

func TestAccConfluentOrganization_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	if data.Client().IsServicePrincipal {
		t.Skip("Skipping due to API issue preventing authentication with service principal")
		return
	}
	r := ConfluentOrganizationResource{}

	data.ResourceTest(t, r, []resource.TestStep{
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
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t ConfluentOrganizationResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.ConfluentOrganizationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Confluent.OrganizationClient.Get(ctx, id.ResourceGroup, id.OrganizationName)
	if err != nil {
		return nil, fmt.Errorf("reading Confluent Organization (%s): %+v", id, err)
	}

	return utils.Bool(resp.OrganizationResourcePropertiesModel != nil), nil
}

func (r ConfluentOrganizationResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-Confluent-%[1]d"
  location = "%[2]s"
}

resource "azurerm_confluent_organization" "test" {
  name = "acctest-CO-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location = azurerm_resource_group.test.location
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r ConfluentOrganizationResource) requiresImport(data acceptance.TestData) string {
	config := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_confluent_organization" "import" {
  name = azurerm_confluent_organization.test.name
  resource_group_name = azurerm_confluent_organization.test.resource_group_name
  location = azurerm_confluent_organization.test.location
}
`, config)
}

func (r ConfluentOrganizationResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-Confluent-%[1]d"
  location = "%[2]s"
}

resource "azurerm_confluent_organization" "test" {
  name = "acctest-CO-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location = azurerm_resource_group.test.location
  offer_detail {
    plan_id = "exmaplePlanId"
    plan_name = "examplePlanName"
    publisher_id = "examplePublisherId"
    term_unit = "exampleTermUnit"
  }

  user_detail {
    email_address = "contoso@microsoft.com"
    first_name = "Terraform"
    last_name = "Azure"
  }

  tags = {
    ENV = "Test"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r ConfluentOrganizationResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-Confluent-%[1]d"
  location = "%[2]s"
}

resource "azurerm_confluent_organization" "test" {
  name = "acctest-CO-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location = azurerm_resource_group.test.location
  offer_detail {
    plan_id = "exmaplePlanId"
    plan_name = "examplePlanName"
    publisher_id = "examplePublisherId"
    term_unit = "exampleTermUnit"
  }

  user_detail {
    email_address = "contoso@microsoft.com"
    first_name = "Terraform"
    last_name = "Azure"
  }

  tags = {
    Pro = "Stage"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
