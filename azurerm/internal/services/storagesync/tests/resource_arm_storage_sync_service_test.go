package tests

import (
    "fmt"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
    "testing"

    "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)


func testCheckAzureRMStorageSyncServiceExists(resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return fmt.Errorf("Storage Sync Service not found: %s", resourceName)
        }

        name := rs.Primary.Attributes["name"]
        resourceGroupName := rs.Primary.Attributes["resource_group"]


        client := acceptance.AzureProvider.Meta().(*clients.Client).StorageSync.storageSyncServicesClient
        ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

        if resp, err := client.Get(ctx, resourceGroupName, name); err != nil {
            if utils.ResponseWasNotFound(resp.Response) {
                return fmt.Errorf("Bad: Storage Sync Service (Storage Sync Service Name %q / Resource Group %q) does not exist", name, resourceGroupName)
            }
            return fmt.Errorf("Bad: Get on storageSyncServicesClient: %+v", err)
        }

        return nil
    }
}

func testCheckAzureRMStorageSyncServiceDestroy(s *terraform.State) error {
    client := acceptance.AzureProvider.Meta().(*clients.Client).StorageSync.storageSyncServicesClient
    ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "azurerm_storage_sync_service" {
            continue
        }

        resourceGroupName := rs.Primary.Attributes["resource_group"]
        name := rs.Primary.Attributes["storage_sync_service_name"]

        if resp, err := client.Get(ctx, resourceGroupName, name); err != nil {
            if !utils.ResponseWasNotFound(resp.Response) {
                return fmt.Errorf("Bad: Get on storageSyncServicesClient: %+v", err)
            }
        }

        return nil
    }
    return nil
}

func TestAccAzureRMStorageSyncService_basic(t *testing.T) {
    data := acceptance.BuildTestData(t, "azurerm_storage_sync_service", "test")

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:     func() { acceptance.PreCheck(t) },
        Providers:    acceptance.SupportedProviders,
        CheckDestroy: testCheckAzureRMStorageSyncServiceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccAzureRMStorageSyncService_basic(data),
                Check: resource.ComposeTestCheckFunc(
                    testCheckAzureRMStorageSyncServiceExists(data.ResourceName),
                ),
            },
            data.ImportStep(),
        },
    })
}

func testAccAzureRMStorageSyncService_basic(data acceptance.TestData) string {
    return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_storage_sync_service" "test" {
  name                         = "acctest-storagesync-%d"
  resource_group_name          = "${azurerm_resource_group.test.name}"
  location                     = "${azurerm_resource_group.test.location}"
  tags = {
    purpose = "testing"
  }
}

`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
