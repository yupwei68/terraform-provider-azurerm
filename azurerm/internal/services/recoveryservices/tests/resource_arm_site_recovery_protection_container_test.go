package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
)

func TestAccAzureRMSiteRecoveryProtectionContainer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_site_recovery_protection_container", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMSiteRecoveryProtectionContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMSiteRecoveryProtectionContainer_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMSiteRecoveryProtectionContainerExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func testAccAzureRMSiteRecoveryProtectionContainer_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-recovery-%d"
  location = "%s"
}

resource "azurerm_recovery_services_vault" "test" {
  name                = "acctest-vault-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"

  soft_delete_enabled = false
}

resource "azurerm_site_recovery_fabric" "test" {
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  name                = "acctest-fabric-%d"
  location            = azurerm_resource_group.test.location
}

resource "azurerm_site_recovery_protection_container" "test" {
  resource_group_name  = azurerm_resource_group.test.name
  recovery_vault_name  = azurerm_recovery_services_vault.test.name
  recovery_fabric_name = azurerm_site_recovery_fabric.test.name
  name                 = "acctest-protection-cont-%d"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func testCheckAzureRMSiteRecoveryProtectionContainerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		state, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		resourceGroupName := state.Primary.Attributes["resource_group_name"]
		vaultName := state.Primary.Attributes["recovery_vault_name"]
		fabricName := state.Primary.Attributes["recovery_fabric_name"]
		protectionContainerName := state.Primary.Attributes["name"]

		client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ProtectionContainerClient(resourceGroupName, vaultName)

		resp, err := client.Get(ctx, fabricName, protectionContainerName)
		if err != nil {
			return fmt.Errorf("Bad: Get on RecoveryServices.ProtectionContainerClient: %+v", err)
		}

		if resp.Response.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Protection Container: %q does not exist", fabricName)
		}

		return nil
	}
}

func testCheckAzureRMSiteRecoveryProtectionContainerDestroy(s *terraform.State) error {
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_site_recovery_protection_container" {
			continue
		}

		resourceGroupName := rs.Primary.Attributes["resource_group_name"]
		vaultName := rs.Primary.Attributes["recovery_vault_name"]
		fabricName := rs.Primary.Attributes["recovery_fabric_name"]
		protectionContainerName := rs.Primary.Attributes["name"]

		client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ProtectionContainerClient(resourceGroupName, vaultName)

		resp, err := client.Get(ctx, fabricName, protectionContainerName)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Protection Container still exists:\n%#v", resp.Properties)
		}
	}

	return nil
}
