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

func TestAccAzureRMStreamAnalyticsOutputSql_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_mssql", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsOutputSqlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStreamAnalyticsOutputSql_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsOutputSqlExists(data.ResourceName),
				),
			},
			data.ImportStep("password"),
		},
	})
}

func TestAccAzureRMStreamAnalyticsOutputSql_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_mssql", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsOutputSqlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStreamAnalyticsOutputSql_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsOutputSqlExists(data.ResourceName),
				),
			},
			{
				Config: testAccAzureRMStreamAnalyticsOutputSql_updated(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsOutputSqlExists(data.ResourceName),
				),
			},
			data.ImportStep("password"),
		},
	})
}

func TestAccAzureRMStreamAnalyticsOutputSql_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_output_mssql", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMStreamAnalyticsOutputSqlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMStreamAnalyticsOutputSql_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMStreamAnalyticsOutputSqlExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMStreamAnalyticsOutputSql_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_stream_analytics_output_mssql"),
			},
		},
	})
}

func testCheckAzureRMStreamAnalyticsOutputSqlExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acceptance.AzureProvider.Meta().(*clients.Client).StreamAnalytics.OutputsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		jobName := rs.Primary.Attributes["stream_analytics_job_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := conn.Get(ctx, resourceGroup, jobName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on streamAnalyticsOutputsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: Stream Output SQL %q (Stream Analytics Job %q / Resource Group %q) does not exist", name, jobName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMStreamAnalyticsOutputSqlDestroy(s *terraform.State) error {
	conn := acceptance.AzureProvider.Meta().(*clients.Client).StreamAnalytics.OutputsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_stream_analytics_output_mssql" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		jobName := rs.Primary.Attributes["stream_analytics_job_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		resp, err := conn.Get(ctx, resourceGroup, jobName, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Stream Analytics Output SQL still exists:\n%#v", resp.OutputProperties)
		}
	}

	return nil
}

func testAccAzureRMStreamAnalyticsOutputSql_basic(data acceptance.TestData) string {
	template := testAccAzureRMStreamAnalyticsOutputSql_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_mssql" "test" {
  name                      = "acctestoutput-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name

  server   = azurerm_sql_server.test.fully_qualified_domain_name
  user     = azurerm_sql_server.test.administrator_login
  password = azurerm_sql_server.test.administrator_login_password
  database = azurerm_sql_database.test.name
  table    = "AccTestTable"
}
`, template, data.RandomInteger)
}

func testAccAzureRMStreamAnalyticsOutputSql_updated(data acceptance.TestData) string {
	template := testAccAzureRMStreamAnalyticsOutputSql_template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_mssql" "test" {
  name                      = "acctestoutput-updated-%d"
  stream_analytics_job_name = azurerm_stream_analytics_job.test.name
  resource_group_name       = azurerm_stream_analytics_job.test.resource_group_name

  server   = azurerm_sql_server.test.fully_qualified_domain_name
  user     = azurerm_sql_server.test.administrator_login
  password = azurerm_sql_server.test.administrator_login_password
  database = azurerm_sql_database.test.name
  table    = "AccTestTable"
}
`, template, data.RandomInteger)
}

func testAccAzureRMStreamAnalyticsOutputSql_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMStreamAnalyticsOutputSql_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_output_mssql" "import" {
  name                      = azurerm_stream_analytics_output_mssql.test.name
  stream_analytics_job_name = azurerm_stream_analytics_output_mssql.test.stream_analytics_job_name
  resource_group_name       = azurerm_stream_analytics_output_mssql.test.resource_group_name

  server   = azurerm_sql_server.test.fully_qualified_domain_name
  user     = azurerm_sql_server.test.administrator_login
  password = azurerm_sql_server.test.administrator_login_password
  database = azurerm_sql_database.test.name
  table    = "AccTestTable"
}
`, template)
}

func testAccAzureRMStreamAnalyticsOutputSql_template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_sql_server" "test" {
  name                         = "acctestserver-%s"
  resource_group_name          = azurerm_resource_group.test.name
  location                     = azurerm_resource_group.test.location
  version                      = "12.0"
  administrator_login          = "acctestadmin"
  administrator_login_password = "t2RX8A76GrnE4EKC"
}

resource "azurerm_sql_database" "test" {
  name                             = "acctestdb"
  resource_group_name              = azurerm_resource_group.test.name
  location                         = azurerm_resource_group.test.location
  server_name                      = azurerm_sql_server.test.name
  requested_service_objective_name = "S0"
  collation                        = "SQL_LATIN1_GENERAL_CP1_CI_AS"
  max_size_bytes                   = "268435456000"
  create_mode                      = "Default"
}

resource "azurerm_stream_analytics_job" "test" {
  name                                     = "acctestjob-%s"
  resource_group_name                      = azurerm_resource_group.test.name
  location                                 = azurerm_resource_group.test.location
  compatibility_level                      = "1.0"
  data_locale                              = "en-GB"
  events_late_arrival_max_delay_in_seconds = 60
  events_out_of_order_max_delay_in_seconds = 50
  events_out_of_order_policy               = "Adjust"
  output_error_policy                      = "Drop"
  streaming_units                          = 3

  transformation_query = <<QUERY
    SELECT *
    INTO [YourOutputAlias]
    FROM [YourInputAlias]
QUERY

}
`, data.RandomInteger, data.Locations.Primary, data.RandomString, data.RandomString)
}
