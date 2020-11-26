package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/datalakestore/paths"
)

func TestAccAzureRMStorageDataLakeGen2Path_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageDataLakeGen2Path_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.RequiresImportErrorStep(testAccAzureRMStorageDataLakeGen2Path_requiresImport),
		},
	})
}

func TestAccAzureRMStorageDataLakeGen2Path_withSimpleACL(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_withSimpleACL(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_withSimpleACLUpdated(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageDataLakeGen2Path_withSimpleACLAndUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_withSimpleACL(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageDataLakeGen2Path_withACLWithSpecificUserAndDefaults(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_withACLWithSpecificUserAndDefaults(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageDataLakeGen2Path_withOwner(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_data_lake_gen2_path", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageDataLakeGen2PathDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageDataLakeGen2Path_withOwner(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageDataLakeGen2PathExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}
func testCheckAzureRMStorageDataLakeGen2PathExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Storage.ADLSGen2PathsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		fileSystemName := rs.Primary.Attributes["filesystem_name"]
		path := rs.Primary.Attributes["path"]
		storageID, err := parse.AccountID(rs.Primary.Attributes["storage_account_id"])
		if err != nil {
			return err
		}

		resp, err := client.GetProperties(ctx, storageID.Name, fileSystemName, path, paths.GetPropertiesActionGetStatus)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Path %q in File System %q (Account %q) does not exist", path, fileSystemName, storageID.Name)
			}

			return fmt.Errorf("Bad: Get on ADLSGen2PathsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMStorageDataLakeGen2PathDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Storage.ADLSGen2PathsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_storage_data_lake_gen2_path" {
			continue
		}

		fileSystemName := rs.Primary.Attributes["filesystem_name"]
		path := rs.Primary.Attributes["path"]
		storageID, err := parse.AccountID(rs.Primary.Attributes["storage_account_id"])
		if err != nil {
			return err
		}

		_, err = client.GetProperties(ctx, storageID.Name, fileSystemName, path, paths.GetPropertiesActionGetStatus)
		if err != nil {
			return nil
		}

		return fmt.Errorf("Path still exists: %q", path)
	}

	return nil
}

func testAccAzureRMStorageDataLakeGen2Path_basic(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_data_lake_gen2_path" "test" {
  storage_account_id = azurerm_storage_account.test.id
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.test.name
  path               = "testpath"
  resource           = "directory"
}
`, template)
}

func testAccAzureRMStorageDataLakeGen2Path_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_data_lake_gen2_path" "import" {
  path               = azurerm_storage_data_lake_gen2_path.test.path
  filesystem_name    = azurerm_storage_data_lake_gen2_path.test.filesystem_name
  storage_account_id = azurerm_storage_data_lake_gen2_path.test.storage_account_id
  resource           = azurerm_storage_data_lake_gen2_path.test.resource
}
`, template)
}

func testAccAzureRMStorageDataLakeGen2Path_withSimpleACL(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_role_assignment" "storage_blob_owner" {
  role_definition_name = "Storage Blob Data Owner"
  scope                = azurerm_resource_group.test.id
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azurerm_storage_data_lake_gen2_path" "test" {
  storage_account_id = azurerm_storage_account.test.id
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.test.name
  path               = "testpath"
  resource           = "directory"
  ace {
    type        = "user"
    permissions = "r-x"
  }
  ace {
    type        = "group"
    permissions = "-wx"
  }
  ace {
    type        = "other"
    permissions = "--x"
  }
}
`, template)
}
func testAccAzureRMStorageDataLakeGen2Path_withSimpleACLUpdated(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_role_assignment" "storage_blob_owner" {
  role_definition_name = "Storage Blob Data Owner"
  scope                = azurerm_resource_group.test.id
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azurerm_storage_data_lake_gen2_path" "test" {
  storage_account_id = azurerm_storage_account.test.id
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.test.name
  path               = "testpath"
  resource           = "directory"
  ace {
    type        = "user"
    permissions = "rwx"
  }
  ace {
    type        = "group"
    permissions = "-wx"
  }
  ace {
    type        = "other"
    permissions = "--x"
  }
}
`, template)
}

func testAccAzureRMStorageDataLakeGen2Path_withACLWithSpecificUserAndDefaults(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_template(data)
	return fmt.Sprintf(`
%s

resource "azuread_application" "test" {
  name = "acctestspa%[2]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azurerm_storage_data_lake_gen2_path" "test" {
  storage_account_id = azurerm_storage_account.test.id
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.test.name
  path               = "testpath"
  resource           = "directory"
  ace {
    type        = "user"
    permissions = "r-x"
  }
  ace {
    type        = "user"
    id          = azuread_service_principal.test.object_id
    permissions = "r-x"
  }
  ace {
    type        = "group"
    permissions = "-wx"
  }
  ace {
    type        = "mask"
    permissions = "--x"
  }
  ace {
    type        = "other"
    permissions = "--x"
  }
  ace {
    scope       = "default"
    type        = "user"
    permissions = "r-x"
  }
  ace {
    scope       = "default"
    type        = "user"
    id          = azuread_service_principal.test.object_id
    permissions = "r-x"
  }
  ace {
    scope       = "default"
    type        = "group"
    permissions = "-wx"
  }
  ace {
    scope       = "default"
    type        = "mask"
    permissions = "--x"
  }
  ace {
    scope       = "default"
    type        = "other"
    permissions = "--x"
  }
}
`, template, data.RandomInteger)
}

func testAccAzureRMStorageDataLakeGen2Path_withOwner(data acceptance.TestData) string {
	template := testAccAzureRMStorageDataLakeGen2Path_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_role_assignment" "storage_blob_owner" {
  role_definition_name = "Storage Blob Data Owner"
  scope                = azurerm_resource_group.test.id
  principal_id         = data.azurerm_client_config.current.object_id
}

resource "azuread_application" "test" {
  name = "acctestspa%[2]d"
}

resource "azuread_service_principal" "test" {
  application_id = azuread_application.test.application_id
}

resource "azurerm_storage_data_lake_gen2_path" "test" {
  storage_account_id = azurerm_storage_account.test.id
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.test.name
  path               = "testpath"
  resource           = "directory"
  owner              = azuread_service_principal.test.object_id
}
`, template, data.RandomInteger)
}

func testAccAzureRMStorageDataLakeGen2Path_template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%[1]d"
  location = "%[2]s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%[3]s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_kind             = "BlobStorage"
  account_tier             = "Standard"
  account_replication_type = "LRS"
  is_hns_enabled           = true
}

data "azurerm_client_config" "current" {
}

resource "azurerm_role_assignment" "storageAccountRoleAssignment" {
  scope                = azurerm_storage_account.test.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = data.azurerm_client_config.current.object_id
}


resource "azurerm_storage_data_lake_gen2_filesystem" "test" {
  name               = "fstest"
  storage_account_id = azurerm_storage_account.test.id
  depends_on = [
    azurerm_role_assignment.storageAccountRoleAssignment
  ]
}

`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}
