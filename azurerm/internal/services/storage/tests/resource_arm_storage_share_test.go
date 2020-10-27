package tests

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
)

func TestAccAzureRMStorageShare_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "deleted"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "remaining_retention_days"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMStorageShare_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_storage_share"),
			},
		},
	})
}

func TestAccAzureRMStorageShare_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					testCheckAzureRMStorageShareDisappears(data.ResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMStorageShare_deleteAndRecreate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageShare_template(data),
			},
			{
				PreConfig: func() { time.Sleep(1 * time.Minute) },
				Config:    testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_metaData(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_metaData(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageShare_metaDataUpdated(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_acl(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_acl(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageShare_aclUpdated(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_aclGhostedRecall(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_aclGhostedRecall(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_updateQuota(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			{
				Config: testAccAzureRMStorageShare_updateQuota(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "quota", "5"),
				),
			},
		},
	})
}

func TestAccAzureRMStorageShare_largeQuota(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_largeQuota(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageShare_largeQuotaUpdate(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_NFS(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_NFS(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "deleted"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "remaining_retention_days"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_NFSUpdate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_NFS(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "deleted"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "remaining_retention_days"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMStorageShare_NFSUpdate(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "deleted"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "remaining_retention_days"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMStorageShare_SMB(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_storage_share", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStorageShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShare_SMB(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "deleted"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "remaining_retention_days"),
				),
			},
			data.ImportStep(),
		},
	})
}

func testCheckAzureRMStorageShareExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		storageClient := acceptance.AzureProvider.Meta().(*clients.Client).Storage
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		shareName := rs.Primary.Attributes["name"]
		accountName := rs.Primary.Attributes["storage_account_name"]

		account, err := storageClient.FindAccount(ctx, accountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q for Share %q: %s", accountName, shareName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
		}

		mgmtFileShareClient := storageClient.MgmtFileSharesClient

		if resp, err := mgmtFileShareClient.Get(ctx, account.ResourceGroup, accountName, shareName, ""); err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("bad: Storage File Share %q (Storage Account Name %q / Resource Group %q) does not exist", shareName, accountName, account.ResourceGroup)
			}
			return fmt.Errorf("bad: Get on Storage File Share Client: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMStorageShareDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		storageClient := acceptance.AzureProvider.Meta().(*clients.Client).Storage
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		shareName := rs.Primary.Attributes["name"]
		accountName := rs.Primary.Attributes["storage_account_name"]

		account, err := storageClient.FindAccount(ctx, accountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q for Share %q: %s", accountName, shareName, err)
		}
		if account == nil {
			return fmt.Errorf("Unable to locate Storage Account %q!", accountName)
		}

		client, err := storageClient.FileSharesClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building FileShare Client: %s", err)
		}

		if _, err := client.Delete(ctx, accountName, shareName, true); err != nil {
			return fmt.Errorf("Error deleting Share %q (Account %q): %v", shareName, accountName, err)
		}

		return nil
	}
}

func testCheckAzureRMStorageShareDestroy(s *terraform.State) error {
	storageClient := acceptance.AzureProvider.Meta().(*clients.Client).Storage
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_storage_share" {
			continue
		}

		shareName := rs.Primary.Attributes["name"]
		accountName := rs.Primary.Attributes["storage_account_name"]

		account, err := storageClient.FindAccount(ctx, accountName)
		if err != nil {
			return fmt.Errorf("Error retrieving Account %q for Share %q: %s", accountName, shareName, err)
		}

		// expected since it's been deleted
		if account == nil {
			return nil
		}

		client, err := storageClient.FileSharesClient(ctx, *account)
		if err != nil {
			return fmt.Errorf("Error building FileShare Client: %s", err)
		}

		props, err := client.GetProperties(ctx, accountName, shareName)
		if err != nil {
			return nil
		}

		return fmt.Errorf("Share still exists: %+v", props)
	}

	return nil
}

func testAccAzureRMStorageShare_basic(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_metaData(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name

  metadata = {
    hello = "world"
  }
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_metaDataUpdated(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name

  metadata = {
    hello = "world"
    happy = "birthday"
  }
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_acl(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name

  acl {
    id = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI"

    access_policy {
      permissions = "rwd"
      start       = "2019-07-02T09:38:21.0000000Z"
      expiry      = "2019-07-02T10:38:21.0000000Z"
    }
  }
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_aclGhostedRecall(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name

  acl {
    id = "GhostedRecall"
    access_policy {
      permissions = "r"
    }
  }
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_aclUpdated(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name

  acl {
    id = "AAAANDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI"

    access_policy {
      permissions = "rwd"
      start       = "2019-07-02T09:38:21.0000000Z"
      expiry      = "2019-07-02T10:38:21.0000000Z"
    }
  }
  acl {
    id = "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI"

    access_policy {
      permissions = "rwd"
      start       = "2019-07-02T09:38:21.0000000Z"
      expiry      = "2019-07-02T10:38:21.0000000Z"
    }
  }
}
`, template, data.RandomString)
}
func testAccAzureRMStorageShare_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "import" {
  name                 = azurerm_storage_share.test.name
  storage_account_name = azurerm_storage_share.test.storage_account_name
}
`, template)
}

func testAccAzureRMStorageShare_updateQuota(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 5
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_largeQuota(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-storageshare-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestshare%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  large_file_share_enabled = true

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 6000
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func testAccAzureRMStorageShare_largeQuotaUpdate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-storageshare-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestshare%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  large_file_share_enabled = true

  tags = {
    environment = "staging"
  }
}

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 10000
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}

func testAccAzureRMStorageShare_template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestacc%s"
  resource_group_name      = azurerm_resource_group.test.name
  location                 = azurerm_resource_group.test.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    environment = "staging"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func testAccAzureRMStorageShare_NFStemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-storage-%[1]d"
  location = "%[2]s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%[1]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctestsubnet%[1]d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
  service_endpoints    = ["Microsoft.Storage"]
}

resource "azurerm_storage_account" "test" {
  name                      = "unlikely23exst2acct%[3]s"
  resource_group_name       = azurerm_resource_group.test.name
  location                  = azurerm_resource_group.test.location
  account_kind              = "FileStorage"
  account_tier              = "Premium"
  account_replication_type  = "LRS"
  access_tier               = "Hot"
  enable_https_traffic_only = false

  network_rules {
    default_action             = "Deny"
    ip_rules                   = ["127.0.0.1"]
    virtual_network_subnet_ids = [azurerm_subnet.test.id]
  }

  tags = {
    environment = "production"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString)
}

func testAccAzureRMStorageShare_NFS(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_NFStemplate(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 4096
  enabled_protocol     = "NFS"
  root_squash          = "AllSquash"
  access_tier          = "Premium"
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_NFSUpdate(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_NFStemplate(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 5120
  enabled_protocol     = "NFS"
  root_squash          = "NoRootSquash"
  access_tier          = "Premium"
}
`, template, data.RandomString)
}

func testAccAzureRMStorageShare_SMB(data acceptance.TestData) string {
	template := testAccAzureRMStorageShare_NFStemplate(data)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share" "test" {
  name                 = "testshare%s"
  storage_account_name = azurerm_storage_account.test.name
  quota                = 1024
  enabled_protocol     = "SMB"
  access_tier          = "Premium"
}
`, template, data.RandomString)
}
