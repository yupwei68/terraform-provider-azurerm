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

type MysqlFlexibleServerFirewallRuleResource struct {
}

func TestAccMysqlFlexibleServerFirewallRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_firewall_rule", "test")
	r := MysqlFlexibleServerFirewallRuleResource{}
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

func TestAccMysqlFlexibleServerFirewallRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_firewall_rule", "test")
	r := MysqlFlexibleServerFirewallRuleResource{}
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

func (MysqlFlexibleServerFirewallRuleResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.FlexibleServerFirewallRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.MySQL.FlexibleServerFirewallRulesClient.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.FirewallRuleName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Mysql Flexible Server Firewall Rule %q (server name: %q / resource group: %q): %+v", id.FirewallRuleName, id.FlexibleServerName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.FirewallRuleProperties != nil), nil
}

func TestAccMysqlFlexibleServerFirewallRule_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_mysql_flexible_server_firewall_rule", "test")
	r := MysqlFlexibleServerFirewallRuleResource{}
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

func (r MysqlFlexibleServerFirewallRuleResource) basic(data acceptance.TestData) string {
	fs := MysqlFlexibleServerResource{}
	return fmt.Sprintf(`
%s

resource "azurerm_mysql_flexible_server_firewall_rule" "test" {
  name               = "acctest-fs-fr-%d"
  flexible_server_id = azurerm_mysql_flexible_server.test.id
  start_ip_address   = "0.0.0.0"
  end_ip_address     = "255.255.255.255"
}
`, fs.basic(data), data.RandomInteger)
}

func (r MysqlFlexibleServerFirewallRuleResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_mysql_flexible_server_firewall_rule" "import" {
  name               = azurerm_mysql_flexible_server_firewall_rule.test.name
  flexible_server_id = azurerm_mysql_flexible_server_firewall_rule.test.flexible_server_id
  start_ip_address   = azurerm_mysql_flexible_server_firewall_rule.test.start_ip_address
  end_ip_address     = azurerm_mysql_flexible_server_firewall_rule.test.end_ip_address
}
`, r.basic(data))
}

func (r MysqlFlexibleServerFirewallRuleResource) update(data acceptance.TestData) string {
	fs := MysqlFlexibleServerResource{}
	return fmt.Sprintf(`
%s

resource "azurerm_mysql_flexible_server_firewall_rule" "test" {
  name               = "acctest-fs-fr-%d"
  flexible_server_id = azurerm_mysql_flexible_server.test.id
  start_ip_address   = "196.97.0.0"
  end_ip_address     = "196.97.10.0"
}
`, fs.basic(data), data.RandomInteger)
}
