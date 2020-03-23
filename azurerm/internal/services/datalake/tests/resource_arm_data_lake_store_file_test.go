package tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/datalake"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestValidateAzureDataLakeStoreRemoteFilePath(t *testing.T) {
	cases := []struct {
		Value  string
		Errors int
	}{
		{
			Value:  "bad",
			Errors: 1,
		},
		{
			Value:  "/good/file/path",
			Errors: 0,
		},
	}

	for _, tc := range cases {
		_, errors := datalake.ValidateDataLakeStoreRemoteFilePath()(tc.Value, "unittest")

		if len(errors) != tc.Errors {
			t.Fatalf("Expected validateDataLakeStoreRemoteFilePath to trigger '%d' errors for '%s' - got '%d'", tc.Errors, tc.Value, len(errors))
		}
	}
}

func TestAccAzureRMDataLakeStoreFile_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataLakeStoreFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataLakeStoreFile_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataLakeStoreFileExists(data.ResourceName),
				),
			},
			data.ImportStep("local_file_path"),
		},
	})
}

func TestAccAzureRMDataLakeStoreFile_largefiles(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")

	//"large" in this context is anything greater than 4 megabytes
	largeSize := 12 * 1024 * 1024 //12 mb
	bytes := make([]byte, largeSize)
	rand.Read(bytes) //fill with random data

	tmpfile, err := ioutil.TempFile("", "azurerm-acc-datalake-file-large")
	if err != nil {
		t.Errorf("Unable to open a temporary file.")
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(bytes); err != nil {
		t.Errorf("Unable to write to temporary file %q: %v", tmpfile.Name(), err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("Unable to close temporary file %q: %v", tmpfile.Name(), err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataLakeStoreFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataLakeStoreFile_largefiles(data, tmpfile.Name()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataLakeStoreFileExists(data.ResourceName),
				),
			},
			data.ImportStep("local_file_path"),
		},
	})
}

func TestAccAzureRMDataLakeStoreFile_requiresimport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_data_lake_store_file", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMDataLakeStoreFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataLakeStoreFile_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMDataLakeStoreFileExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMDataLakeStoreFile_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_data_lake_store_file"),
			},
		},
	})
}

func testCheckAzureRMDataLakeStoreFileExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acceptance.AzureProvider.Meta().(*clients.Client).Datalake.StoreFilesClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		remoteFilePath := rs.Primary.Attributes["remote_file_path"]
		accountName := rs.Primary.Attributes["account_name"]

		resp, err := conn.GetFileStatus(ctx, accountName, remoteFilePath, utils.Bool(true))
		if err != nil {
			return fmt.Errorf("Bad: Get on dataLakeStoreFileClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Date Lake Store File Rule %q (Account %q) does not exist", remoteFilePath, accountName)
		}

		return nil
	}
}

func testCheckAzureRMDataLakeStoreFileDestroy(s *terraform.State) error {
	conn := acceptance.AzureProvider.Meta().(*clients.Client).Datalake.StoreFilesClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_data_lake_store_file" {
			continue
		}

		remoteFilePath := rs.Primary.Attributes["remote_file_path"]
		accountName := rs.Primary.Attributes["account_name"]

		resp, err := conn.GetFileStatus(ctx, accountName, remoteFilePath, utils.Bool(true))
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				return nil
			}

			return err
		}

		return fmt.Errorf("Data Lake Store File still exists:\n%#v", resp)
	}

	return nil
}

func testAccAzureRMDataLakeStoreFile_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = "%s"
  firewall_state      = "Disabled"
}

resource "azurerm_data_lake_store_file" "test" {
  remote_file_path = "/test/application_gateway_test.cer"
  account_name     = azurerm_data_lake_store.test.name
  local_file_path  = "./testdata/application_gateway_test.cer"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.Locations.Primary)
}

func testAccAzureRMDataLakeStoreFile_largefiles(data acceptance.TestData, file string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_data_lake_store" "test" {
  name                = "unlikely23exst2acct%s"
  resource_group_name = azurerm_resource_group.test.name
  location            = "%s"
  firewall_state      = "Disabled"
}

resource "azurerm_data_lake_store_file" "test" {
  remote_file_path = "/test/testAccAzureRMDataLakeStoreFile_largefiles.bin"
  account_name     = azurerm_data_lake_store.test.name
  local_file_path  = "%s"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.Locations.Primary, file)
}

func testAccAzureRMDataLakeStoreFile_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMDataLakeStoreFile_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_data_lake_store_file" "import" {
  remote_file_path = azurerm_data_lake_store_file.test.remote_file_path
  account_name     = azurerm_data_lake_store_file.test.name
  local_file_path  = "./testdata/application_gateway_test.cer"
}
`, template)
}
