package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccDataSourceAzureRMDataShareAccount_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_data_share_account", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataShareAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDataShareAccount_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataShareAccountExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "tags.env", "Test"),
				),
			},
		},
	})
}

func testAccDataSourceDataShareAccount_basic(data acceptance.TestData) string {
	config := testAccAzureRMDataShareAccount_complete(data)
	return fmt.Sprintf(`
%s
data "azurerm_data_shareaccount" "test" {
  name                = azurerm_data_share_account.test.name
  resource_group_name = azurerm_resource_group.test.name
}
`, config)
}
