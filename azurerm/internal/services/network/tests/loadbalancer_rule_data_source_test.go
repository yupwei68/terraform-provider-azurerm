package tests

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAzureRMDataSourceLoadBalancerRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_lb_rule", "test")
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataSourceLoadBalancerRule_basic(data, lbRuleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(data.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "frontend_ip_configuration_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "protocol"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "frontend_port"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "backend_port"),
				),
			},
		},
	})
}

func TestAccAzureRMDataSourceLoadBalancerRule_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_lb_rule", "test")
	backendPoolName := fmt.Sprintf("LbPool-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	lbRuleName := fmt.Sprintf("LbRule-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	probeName := fmt.Sprintf("LbProbe-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMDataSourceLoadBalancerRule_complete(data, lbRuleName, backendPoolName, probeName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(data.ResourceName, "id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "frontend_ip_configuration_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "protocol"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "frontend_port"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "backend_port"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "backend_address_pool_id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "probe_id"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "enable_floating_ip"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "enable_tcp_reset"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "disable_outbound_snat"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "idle_timeout_in_minutes"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "load_distribution"),
				),
			},
		},
	})
}

func testAccAzureRMDataSourceLoadBalancerRule_basic(data acceptance.TestData, name string) string {
	resource := testAccAzureRMLoadBalancerRule_basic(data, name, "Basic")
	return fmt.Sprintf(`
%s

data "azurerm_lb_rule" "test" {
  name                = azurerm_lb_rule.test.name
  resource_group_name = azurerm_lb_rule.test.resource_group_name
  loadbalancer_id     = azurerm_lb_rule.test.loadbalancer_id
}
`, resource)
}

func testAccAzureRMDataSourceLoadBalancerRule_complete(data acceptance.TestData, lbRuleName string, backendPoolName string, probeName string) string {
	return fmt.Sprintf(`
%s
resource "azurerm_lb_backend_address_pool" "test" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id
}

resource "azurerm_lb_probe" "test" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id
  protocol            = "Tcp"
  port                = 443
}

resource "azurerm_lb_rule" "test" {
  name                = "%s"
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id

  protocol      = "Tcp"
  frontend_port = 3389
  backend_port  = 3389

  disable_outbound_snat   = true
  enable_floating_ip      = true
  enable_tcp_reset        = true
  idle_timeout_in_minutes = 10

  backend_address_pool_id = azurerm_lb_backend_address_pool.test.id
  probe_id                = azurerm_lb_probe.test.id

  frontend_ip_configuration_name = azurerm_lb.test.frontend_ip_configuration.0.name
}

data "azurerm_lb_rule" "test" {
  name                = azurerm_lb_rule.test.name
  resource_group_name = azurerm_lb_rule.test.resource_group_name
  loadbalancer_id     = azurerm_lb_rule.test.loadbalancer_id
}
`, testAccAzureRMLoadBalancerRule_template(data, "Standard"), backendPoolName, probeName, lbRuleName)
}
