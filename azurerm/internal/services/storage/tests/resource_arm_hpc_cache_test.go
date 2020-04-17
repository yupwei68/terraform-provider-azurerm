package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
)

func TestAccAzureRMHPCCache_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_hpc_cache", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMHPCCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHPCCache_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHPCCacheExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "mount_addresses.#"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMHPCCache_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_hpc_cache", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMHPCCacheDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMHPCCache_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMHPCCacheExists(data.ResourceName),
				),
			},
			data.RequiresImportErrorStep(testAccAzureRMHPCCahce_requiresImport),
		},
	})
}

func testCheckAzureRMHPCCacheExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acceptance.AzureProvider.Meta().(*clients.Client).Storage.CachesClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on storageCacheCachesClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: HPC Cache %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMHPCCacheDestroy(s *terraform.State) error {
	conn := acceptance.AzureProvider.Meta().(*clients.Client).Storage.CachesClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_hpc_cache" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		resp, err := conn.Get(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("HPC Cache still exists:\n%#v", resp)
		}
	}

	return nil
}

func testAccAzureRMHPCCache_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-storage-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctest-VN-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctestsub-%d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_hpc_cache" "test" {
  name                = "acctest-HPCC-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  cache_size_in_gb    = 3072
  subnet_id           = azurerm_subnet.test.id
  sku_name            = "Standard_2G"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMHPCCahce_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMHPCCache_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_hpc_cache" "import" {
  name                = azurerm_hpc_cache.test.name
  resource_group_name = azurerm_hpc_cache.test.resource_group_name
  location            = azurerm_hpc_cache.test.location
  cache_size_in_gb    = azurerm_hpc_cache.test.cache_size_in_gb
  subnet_id           = azurerm_hpc_cache.test.subnet_id
  sku_name            = azurerm_hpc_cache.test.sku_name
}
`, template)
}
