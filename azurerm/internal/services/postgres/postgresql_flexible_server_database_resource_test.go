package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/postgres/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type PostgresqlFlexibleServerDatabaseResource struct {
}

func TestAccPostgresqlFlexibleServerDatabase_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_postgresql_flexible_server_database", "test")
	r := PostgresqlFlexibleServerDatabaseResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccPostgresqlFlexibleServerDatabase_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_postgresql_flexible_server_database", "test")
	r := PostgresqlFlexibleServerDatabaseResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccPostgresqlFlexibleServerDatabase_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_postgresql_flexible_server_database", "test")
	r := PostgresqlFlexibleServerDatabaseResource{}
	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (PostgresqlFlexibleServerDatabaseResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.FlexibleServerDatabaseID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Postgres.FlexibleServerDatabasesClient.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("retrieving %q: %+v", id, err)
	}

	return utils.Bool(resp.DatabaseProperties != nil), nil
}

func (PostgresqlFlexibleServerDatabaseResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_postgresql_flexible_server_database" "test" {
  name      = "acctest-FSDB-%d"
  server_id = azurerm_postgresql_flexible_server.test.id
}
`, PostgresqlFlexibleServerResource{}.basic(data), data.RandomInteger)
}

func (r PostgresqlFlexibleServerDatabaseResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_postgresql_flexible_server_database" "import" {
  name      = azurerm_postgresql_flexible_server_database.test.name
  server_id = azurerm_postgresql_flexible_server_database.test.server_id
}
`, r.basic(data))
}

func (r PostgresqlFlexibleServerDatabaseResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_postgresql_flexible_server_database" "test" {
  name      = "acctest-FSFR-%d"
  server_id = azurerm_postgresql_flexible_server.test.id
  charset   = "utf8"
  collation = "en_US.utf8"
}
`, PostgresqlFlexibleServerResource{}.basic(data), data.RandomInteger)
}
