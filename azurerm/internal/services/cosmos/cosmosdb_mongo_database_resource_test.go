package cosmos_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/preview/cosmos-db/mgmt/2020-04-01-preview/documentdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMCosmosDbMongoDatabase_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_mongo_database", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMCosmosDbMongoDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_basic(data),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMCosmosDbMongoDatabase_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_mongo_database", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMCosmosDbMongoDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_complete(data),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMCosmosDbMongoDatabase_autoscale(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_mongo_database", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMCosmosDbMongoDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_autoscale(data, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "autoscale_settings.0.max_throughput", "4000"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_autoscale(data, 5000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "autoscale_settings.0.max_throughput", "5000"),
				),
			},
			data.ImportStep(),
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_autoscale(data, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "autoscale_settings.0.max_throughput", "4000"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMCosmosDbMongoDatabase_serverless(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_mongo_database", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMCosmosDbMongoDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMCosmosDbMongoDatabase_serverless(data),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckAzureRMCosmosDbMongoDatabaseExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func testCheckAzureRMCosmosDbMongoDatabaseDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Cosmos.MongoDbClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_cosmosdb_mongo_database" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		account := rs.Primary.Attributes["account_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.GetMongoDBDatabase(ctx, resourceGroup, account, name)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Error checking destroy for Cosmos Mongo Database %s (account %s) still exists:\n%v", name, account, err)
			}
		}

		if !utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("Cosmos Mongo Database %s (account %s) still exists:\n%#v", name, account, resp)
		}
	}

	return nil
}

func testCheckAzureRMCosmosDbMongoDatabaseExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Cosmos.MongoDbClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		account := rs.Primary.Attributes["account_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.GetMongoDBDatabase(ctx, resourceGroup, account, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on cosmosAccountsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Cosmos database '%s' (account: '%s') does not exist", name, account)
		}

		return nil
	}
}

func testAccAzureRMCosmosDbMongoDatabase_basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_mongo_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
}
`, testAccAzureRMCosmosDBAccount_basic(data, documentdb.MongoDB, documentdb.Strong), data.RandomInteger)
}

func testAccAzureRMCosmosDbMongoDatabase_complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_mongo_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
  throughput          = 700
}
`, testAccAzureRMCosmosDBAccount_basic(data, documentdb.MongoDB, documentdb.Strong), data.RandomInteger)
}

func testAccAzureRMCosmosDbMongoDatabase_autoscale(data acceptance.TestData, maxThroughput int) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_mongo_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
  autoscale_settings {
    max_throughput = %[3]d
  }
}
`, testAccAzureRMCosmosDBAccount_basic(data, documentdb.MongoDB, documentdb.Strong), data.RandomInteger, maxThroughput)
}

func testAccAzureRMCosmosDbMongoDatabase_serverless(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_cosmosdb_mongo_database" "test" {
  name                = "acctest-%[2]d"
  resource_group_name = azurerm_cosmosdb_account.test.resource_group_name
  account_name        = azurerm_cosmosdb_account.test.name
}
`, testAccAzureRMCosmosDBAccount_capabilities(data, documentdb.MongoDB, []string{"EnableServerless", "mongoEnableDocLevelTTL", "EnableMongo"}), data.RandomInteger)
}
