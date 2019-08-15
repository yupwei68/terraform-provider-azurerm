package azurerm

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
)

func TestAccAzureRMStorageShareFile_basic(t *testing.T) {
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(5))
	location := testLocation()
	resourceName := "azurerm_storage_share_file.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShareFile_basic(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMStorageShareFile_requiresImport(t *testing.T) {
	if !requireResourcesToBeImported {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(5))
	location := testLocation()
	resourceName := "azurerm_storage_share_file.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShareFile_basic(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				Config:      testAccAzureRMStorageShareFile_requiresImport(ri, rs, location),
				ExpectError: testRequiresImportError("azurerm_storage_share_file"),
			},
		},
	})
}

func TestAccAzureRMStorageShareFile_complete(t *testing.T) {
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(5))
	location := testLocation()
	resourceName := "azurerm_storage_share_file.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShareFile_complete(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMStorageShareFile_update(t *testing.T) {
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(5))
	location := testLocation()
	resourceName := "azurerm_storage_share_file.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShareFile_complete(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAzureRMStorageShareFile_updated(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMStorageShareFile_topLevel(t *testing.T) {
	ri := tf.AccRandTimeInt()
	rs := strings.ToLower(acctest.RandString(5))
	location := testLocation()
	resourceName := "azurerm_storage_share_file.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMStorageShareFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStorageShareFile_topLevel(ri, rs, location),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStorageShareFileExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckAzureRMStorageShareFileExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		shareName := rs.Primary.Attributes["share_name"]
		directoryName := rs.Primary.Attributes["share_directory_name"]
		accountName := rs.Primary.Attributes["storage_account_name"]

		storageClient := testAccProvider.Meta().(*ArmClient).storage
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resourceGroup, err := storageClient.FindResourceGroup(ctx, accountName)
		if err != nil {
			return fmt.Errorf("Error locating Resource Group for Storage Share File %q (Share %s, Account %s): %s", name, shareName, accountName, err)
		}
		if resourceGroup == nil {
			return fmt.Errorf("Unable to locate Resource Group for Storage Share File %q (Share %s, Account %s) ", name, shareName, accountName)
		}

		client, err := storageClient.FilesClient(ctx, *resourceGroup, accountName)
		if err != nil {
			return fmt.Errorf("Error building Files Client: %s", err)
		}

		resp, err := client.GetProperties(ctx, accountName, shareName, directoryName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on FilesClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: File %q (File Share %q / Account %q / Resource Group %q) does not exist", name, shareName, accountName, *resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMStorageShareFileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_storage_share_file" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		shareName := rs.Primary.Attributes["share_name"]
		directoryName := rs.Primary.Attributes["share_directory_name"]
		accountName := rs.Primary.Attributes["storage_account_name"]

		storageClient := testAccProvider.Meta().(*ArmClient).storage
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resourceGroup, err := storageClient.FindResourceGroup(ctx, accountName)
		if err != nil {
			return fmt.Errorf("Error locating Resource Group for Storage Share File %q (Share %s, Account %s): %s", name, shareName, accountName, err)
		}

		// not found, the account's gone
		if resourceGroup == nil {
			return nil
		}

		client, err := storageClient.FilesClient(ctx, *resourceGroup, accountName)
		if err != nil {
			return fmt.Errorf("Error building File Client: %s", err)
		}

		resp, err := client.GetProperties(ctx, accountName, shareName, directoryName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on FileShareDirectoriesClient: %+v", err)
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("File Share still exists:\n%#v", resp)
		}
	}

	return nil
}

func testAccAzureRMStorageShareFile_basic(rInt int, rString string, location string) string {
	template := testAccAzureRMStorageShareFile_template(rInt, rString, location)
	return fmt.Sprintf(`
%s


resource "azurerm_storage_share_file" "test" {
  name                 = "README.md"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = "${azurerm_storage_share_directory.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"
}
`, template)
}

func testAccAzureRMStorageShareFile_requiresImport(rInt int, rString string, location string) string {
	template := testAccAzureRMStorageShareFile_basic(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share_file" "test" {
  name                 = "README.md"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = "${azurerm_storage_share_directory.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"
}
`, template)
}

func testAccAzureRMStorageShareFile_complete(rInt int, rString string, location string) string {
	template := testAccAzureRMStorageShareFile_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share_file" "test" {
  name                 = "README.md"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = "${azurerm_storage_share_directory.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"

  content_length = 100

  metadata = {
    hello = "world"
  }
}
`, template)
}

func testAccAzureRMStorageShareFile_updated(rInt int, rString string, location string) string {
	template := testAccAzureRMStorageShareFile_template(rInt, rString, location)
	return fmt.Sprintf(`
%s

resource "azurerm_storage_share_file" "test" {
  name                 = "README.md"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = "${azurerm_storage_share_directory.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"

  content_length = 200

  metadata = {
    hello    = "world"
    sunshine = "at dawn"
  }
}
`, template)
}

func testAccAzureRMStorageShareFile_topLevel(rInt int, rString string, location string) string {
	template := testAccAzureRMStorageShareFile_template(rInt, rString, location)
	return fmt.Sprintf(`
%s


resource "azurerm_storage_share_file" "test" {
  name                 = "README.md"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = ""
  storage_account_name = "${azurerm_storage_account.test.name}"
}
`, template)
}

func testAccAzureRMStorageShareFile_template(rInt int, rString string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctestsa%s"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  location                 = "${azurerm_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_share" "test" {
  name                 = "fileshare"
  storage_account_name = "${azurerm_storage_account.test.name}"
  quota                = 50
}

resource "azurerm_storage_share_directory" "test" {
  name                 = "dir"
  share_name           = "${azurerm_storage_share.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"
}
`, rInt, location, rString)
}
