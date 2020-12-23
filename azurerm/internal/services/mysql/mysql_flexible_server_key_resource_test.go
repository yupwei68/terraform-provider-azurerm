package mysql_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mysql/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"testing"
)

type MysqlFlexibleServerKeyResource struct {
}

func TestAccMysqlFlexibleServerKey_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_key", "test")
	r := MysqlFlexibleServerKeyResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccMysqlFlexibleServerKey_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_key", "test")
	r := MysqlFlexibleServerKeyResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func (MysqlFlexibleServerKeyResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.FlexibleServerKeyID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.MySQL.FlexibleServerKeysClient.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.KeyName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Mysql Flexible Server Key %q (server name: %q / resource group: %q): %+v", id.KeyName, id.FlexibleServerName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.ServerKeyProperties != nil), nil
}

func TestAccMysqlFlexibleServerKey_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_key", "test")
	r := MysqlFlexibleServerKeyResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r MysqlFlexibleServerKeyResource) template(data acceptance.TestData) string {
	fs := MysqlFlexibleServerResource{}
	return fmt.Sprintf(`
%s

data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "test" {
  name                     = "acctestkv%s"
  location                 = azurerm_resource_group.test.location
  resource_group_name      = azurerm_resource_group.test.name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  sku_name                 = "standard"
  soft_delete_enabled      = true
  purge_protection_enabled = true
}

resource "azurerm_key_vault_access_policy" "flexible_server" {
  key_vault_id = azurerm_key_vault.test.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_mysql_flexible_server.test.identity.0.principal_id

  key_permissions    = ["get", "create", "list", "restore", "recover", "unwrapkey", "wrapkey", "purge", "encrypt", "decrypt", "sign", "verify"]
  secret_permissions = ["get"]
}

resource "azurerm_key_vault_access_policy" "client" {
  key_vault_id = azurerm_key_vault.test.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = data.azurerm_client_config.current.object_id

  key_permissions    = ["get", "create", "delete", "list", "restore", "recover", "unwrapkey", "wrapkey", "purge", "encrypt", "decrypt", "sign", "verify"]
  secret_permissions = ["get"]
}

resource "azurerm_key_vault_key" "first" {
  name         = "first"
  key_vault_id = azurerm_key_vault.test.id
  key_type     = "RSA"
  key_size     = 2048
  key_opts     = ["decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"]

  depends_on = [
    azurerm_key_vault_access_policy.client,
    azurerm_key_vault_access_policy.flexible_server,
  ]
}

`, fs.complete(data), data.RandomString)
}

func (r MysqlFlexibleServerKeyResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mysql_flexible_server_key" "test" {
  name               = "acctest-fs-key-%d"
  flexible_server_id = azurerm_mysql_flexible_server.test.id
  server_key_type    = "AzureKeyVault"
  key_vault_key_id   = azurerm_key_vault_key.first.id
}
`, r.template(data), data.RandomInteger)
}

func (r MysqlFlexibleServerKeyResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mysql_flexible_server_key" "import" {
  name               = azurerm_mysql_flexible_server_key.test.name
  flexible_server_id = azurerm_mysql_flexible_server_key.test.flexible_server_id
  server_key_type    = azurerm_mysql_flexible_server_key.test.server_key_type
  key_vault_key_id   = azurerm_mysql_flexible_server_key.test.key_vault_key_id
}
`, r.basic(data))
}

func (r MysqlFlexibleServerKeyResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_key_vault_key" "second" {
  name         = "first"
  key_vault_id = azurerm_key_vault.test.id
  key_type     = "RSA"
  key_size     = 2048
  key_opts     = ["decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"]

  depends_on = [
    azurerm_key_vault_access_policy.client,
    azurerm_key_vault_access_policy.flexible_server,
  ]
}

resource "azurerm_mysql_flexible_server_key" "test" {
  name               = "acctest-fs-key-%d"
  flexible_server_id = azurerm_mysql_flexible_server.test.id
  server_key_type    = "AzureKeyVault"
  key_vault_key_id   = azurerm_key_vault_key.second.id
}
`, r.template(data), data.RandomInteger)
}
