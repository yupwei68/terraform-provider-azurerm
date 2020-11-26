package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccDataSourceArmManagementGroup_basicByName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmManagementGroup_basicByName(data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestmg-%d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "subscription_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceArmManagementGroup_basicByDisplayName(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: acceptance.SupportedProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArmManagementGroup_basicByDisplayName(data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctest Management Group %d", data.RandomInteger)),
					resource.TestCheckResourceAttr(data.ResourceName, "subscription_ids.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceArmManagementGroup_basicByName(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  display_name = "acctestmg-%d"
}

data "azurerm_management_group" "test" {
  name = azurerm_management_group.test.name
}
`, data.RandomInteger)
}

func testAccDataSourceArmManagementGroup_basicByDisplayName(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  display_name = "acctest Management Group %d"
}

data "azurerm_management_group" "test" {
  display_name = azurerm_management_group.test.display_name
}
`, data.RandomInteger)
}
