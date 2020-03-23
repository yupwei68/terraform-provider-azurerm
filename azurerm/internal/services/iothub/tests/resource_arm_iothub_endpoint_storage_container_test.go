package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMIotHubEndpointStorageContainer_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_iothub_endpoint_storage_container", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testAccAzureRMIotHubEndpointStorageContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMIotHubEndpointStorageContainer_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testAccAzureRMIotHubEndpointStorageContainerExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "file_name_format", "{iothub}/{partition}_{YYYY}_{MM}_{DD}_{HH}_{mm}"),
					resource.TestCheckResourceAttr(data.ResourceName, "batch_frequency_in_seconds", "60"),
					resource.TestCheckResourceAttr(data.ResourceName, "max_chunk_size_in_bytes", "10485760"),
					resource.TestCheckResourceAttr(data.ResourceName, "encoding", "JSON"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMIotHubEndpointStorageContainer_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}
	data := acceptance.BuildTestData(t, "azurerm_iothub_endpoint_storage_container", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testAccAzureRMIotHubEndpointStorageContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMIotHubEndpointStorageContainer_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testAccAzureRMIotHubEndpointStorageContainerExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMIotHubEndpointStorageContainer_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_iothub_endpoint_storage_container"),
			},
		},
	})
}

func testAccAzureRMIotHubEndpointStorageContainer_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-iothub-%[1]d"
  location = "%[2]s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acc%[1]d"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "test" {
  name                  = "acctestcont"
  storage_account_name  = azurerm_storage_account.test.name
  container_access_type = "private"
}

resource "azurerm_iothub" "test" {
  name                = "acctestIoTHub-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

  sku {
    name     = "B1"
    capacity = "1"
  }

  tags = {
    purpose = "testing"
  }
}

resource "azurerm_iothub_endpoint_storage_container" "test" {
  resource_group_name = azurerm_resource_group.test.name
  iothub_name         = azurerm_iothub.test.name
  name                = "acctest"

  container_name    = "acctestcont"
  connection_string = azurerm_storage_account.test.primary_blob_connection_string

  file_name_format           = "{iothub}/{partition}_{YYYY}_{MM}_{DD}_{HH}_{mm}"
  batch_frequency_in_seconds = 60
  max_chunk_size_in_bytes    = 10485760
  encoding                   = "JSON"
}
`, data.RandomInteger, data.Locations.Primary)
}

func testAccAzureRMIotHubEndpointStorageContainer_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMIotHubEndpointStorageContainer_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_iothub_endpoint_storage_container" "import" {
  resource_group_name = azurerm_resource_group.test.name
  iothub_name         = azurerm_iothub.test.name
  name                = "acctest"

  container_name    = "acctestcont"
  connection_string = azurerm_storage_account.test.primary_blob_connection_string

  file_name_format           = "{iothub}/{partition}_{YYYY}_{MM}_{DD}_{HH}_{mm}"
  batch_frequency_in_seconds = 60
  max_chunk_size_in_bytes    = 10485760
  encoding                   = "JSON"
}
`, template)
}

func testAccAzureRMIotHubEndpointStorageContainerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).IoTHub.ResourceClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		parsedIothubId, err := azure.ParseAzureResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		iothubName := parsedIothubId.Path["IotHubs"]
		endpointName := parsedIothubId.Path["Endpoints"]
		resourceGroup := parsedIothubId.ResourceGroup
		iothub, err := client.Get(ctx, resourceGroup, iothubName)
		if err != nil {
			if utils.ResponseWasNotFound(iothub.Response) {
				return fmt.Errorf("IotHub %q (Resource Group %q) was not found", iothubName, resourceGroup)
			}

			return fmt.Errorf("Error loading IotHub %q (Resource Group %q): %+v", iothubName, resourceGroup, err)
		}

		if iothub.Properties == nil || iothub.Properties.Routing == nil || iothub.Properties.Routing.Endpoints == nil {
			return fmt.Errorf("Bad: No endpoint %s defined for IotHub %s", endpointName, iothubName)
		}
		endpoints := iothub.Properties.Routing.Endpoints.StorageContainers

		if endpoints == nil {
			return fmt.Errorf("Bad: No Storage Container endpoint %s defined for IotHub %s", endpointName, iothubName)
		}

		for _, endpoint := range *endpoints {
			if existingEndpointName := endpoint.Name; existingEndpointName != nil {
				if strings.EqualFold(*existingEndpointName, endpointName) {
					return nil
				}
			}
		}

		return fmt.Errorf("Bad: No Storage Container endpoint %s defined for IotHub %s", endpointName, iothubName)
	}
}

func testAccAzureRMIotHubEndpointStorageContainerDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).IoTHub.ResourceClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_iothub_endpoint_storage_container" {
			continue
		}

		endpointName := rs.Primary.Attributes["name"]
		iothubName := rs.Primary.Attributes["iothub_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		iothub, err := client.Get(ctx, resourceGroup, iothubName)
		if err != nil {
			if utils.ResponseWasNotFound(iothub.Response) {
				return nil
			}
			return fmt.Errorf("Bad: Get on iothubResourceClient: %+v", err)
		}
		if iothub.Properties == nil || iothub.Properties.Routing == nil || iothub.Properties.Routing.Endpoints == nil {
			return nil
		}

		endpoints := iothub.Properties.Routing.Endpoints.StorageContainers
		if endpoints == nil {
			return nil
		}

		for _, endpoint := range *endpoints {
			if existingEndpointName := endpoint.Name; existingEndpointName != nil {
				if strings.EqualFold(*existingEndpointName, endpointName) {
					return fmt.Errorf("Bad: Storage Container endpoint %s still exists on IoTHb %s", endpointName, iothubName)
				}
			}
		}
	}
	return nil
}
