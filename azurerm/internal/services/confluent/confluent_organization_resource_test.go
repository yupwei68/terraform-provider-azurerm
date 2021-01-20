package confluent_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/confluent/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMconfluentOrganization_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMconfluentOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMconfluentOrganization_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMconfluentOrganization_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMconfluentOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMconfluentOrganization_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.RequiresImportErrorStep(testAccAzureRMconfluentOrganization_requiresImport),
		},
	})
}

func TestAccAzureRMconfluentOrganization_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMconfluentOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMconfluentOrganization_complete(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMconfluentOrganization_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_confluent_organization", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMconfluentOrganizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMconfluentOrganization_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMconfluentOrganization_complete(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMconfluentOrganization_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMconfluentOrganizationExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func testCheckAzureRMconfluentOrganizationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Confluent.OrganizationClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("confluent Organization not found: %s", resourceName)
		}
		id, err := parse.ConfluentOrganizationID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp, err := client.Get(ctx, id.ResourceGroup, id.OrganizationName); err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("bad: Confluent Organization %q does not exist", id.OrganizationName)
			}
			return fmt.Errorf("bad: Get on Confluent.OrganizationClient: %+v", err)
		}
		return nil
	}
}

func testCheckAzureRMconfluentOrganizationDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Confluent.OrganizationClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_confluent_organization" {
			continue
		}
		id, err := parse.ConfluentOrganizationID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp, err := client.Get(ctx, id.ResourceGroup, id.OrganizationName); err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("bad: Get on Confluent.OrganizationClient: %+v", err)
			}
		}
		return nil
	}
	return nil
}

func testAccAzureRMconfluentOrganization_template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-confluent-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func testAccAzureRMconfluentOrganization_basic(data acceptance.TestData) string {
	template := testAccAzureRMconfluentOrganization_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_confluent_organization" "test" {
  name = "acctest-co-%d"
  resource_group_name = azurerm_resource_group.test.name
  location = azurerm_resource_group.test.location
}
`, template, data.RandomInteger)
}

func testAccAzureRMconfluentOrganization_requiresImport(data acceptance.TestData) string {
	config := testAccAzureRMconfluentOrganization_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_confluent_organization" "import" {
  name = azurerm_confluent_organization.test.name
  resource_group_name = azurerm_confluent_organization.test.resource_group_name
  location = azurerm_confluent_organization.test.location
}
`, config)
}

func testAccAzureRMconfluentOrganization_complete(data acceptance.TestData) string {
	template := testAccAzureRMconfluentOrganization_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_confluent_organization" "test" {
  name = "acctest-co-%d"
  resource_group_name = azurerm_resource_group.test.name
  location = azurerm_resource_group.test.location
  offer_detail {
    plan_id = "string"
    plan_name = "string"
    publisher_id = "string"
    term_unit = "string"
  }

  user_detail {
    email_address = "contoso@microsoft.com"
    first_name = "string"
    last_name = "string"
  }

  tags = {
    ENV = "Test"
  }
}
`, template, data.RandomInteger)
}
