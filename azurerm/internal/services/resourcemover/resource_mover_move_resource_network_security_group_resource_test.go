package resourcemover_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type ResourceMoverMoveResourceNetworkSecurityGroupResource struct {
}

func TestAccResourceMoverMoveResourceNetworkSecurityGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_security_group", "test")
	r := ResourceMoverMoveResourceNetworkSecurityGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourceNetworkSecurityGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_security_group", "test")
	r := ResourceMoverMoveResourceNetworkSecurityGroupResource{}
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

func TestAccResourceMoverMoveResourceNetworkSecurityGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_security_group", "test")
	r := ResourceMoverMoveResourceNetworkSecurityGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourceNetworkSecurityGroup_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_security_group", "test")
	r := ResourceMoverMoveResourceNetworkSecurityGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourceNetworkSecurityGroup_existing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_security_group", "test")
	r := ResourceMoverMoveResourceNetworkSecurityGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.existing(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("CommitPending"),
			),
		},
		data.ImportStep(),
	})
}

func (ResourceMoverMoveResourceNetworkSecurityGroupResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.ResourceMoverMoveResourceID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ResourceMover.MoveResourceClient.Get(ctx, id.ResourceGroup, id.MoveCollectionName, id.MoveResourceName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Resource Mover Move Resource %q (resource group: %q): %+v", id.MoveResourceName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (r ResourceMoverMoveResourceNetworkSecurityGroupResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_network_security_group" "test" {
  name                = "acctest-nsg-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_network_security_rule" "test" {
  name                        = "acctest-nsg-rule-%[2]d"
  network_security_group_name = azurerm_network_security_group.test.name
  resource_group_name         = azurerm_resource_group.test.name
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  description                 = "acctest source rule"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
}
`, ResourceMoverMoveResourceResourceGroupResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkSecurityGroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_network_security_group" "test" {
  name               = "acctest-MR-NSG-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_security_group.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-nsg"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkSecurityGroupResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_network_security_group" "import" {
  name               = azurerm_resource_mover_move_resource_network_security_group.test.name
  move_collection_id = azurerm_resource_mover_move_resource_network_security_group.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_network_security_group.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_network_security_group.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourceNetworkSecurityGroupResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_network_security_group" "test" {
  name               = "acctest-MR-NSG-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_security_group.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-nsg"
    security_rule {
      name                       = "acctest-nsg-tar-rule"
      priority                   = "101"
      access                     = "Deny"
      description                = "acctest target rule"
      direction                  = "Inbound"
      protocol                   = "Icmp"
      source_address_prefix      = "10.0.0.0/8"
      destination_address_prefix = "10.0.1.0/24"
      source_port_range          = "1443"
      destination_port_range     = "443"
    }
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkSecurityGroupResource) existing(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = azurerm_resource_mover_move_collection.test.target_region
}

resource "azurerm_network_security_group" "target" {
  name                = "acctest-nsg-tar-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
}

resource "azurerm_network_security_rule" "target" {
  name                        = "acctest-nsg-rule-tar-%[2]d"
  network_security_group_name = azurerm_network_security_group.target.name
  resource_group_name         = azurerm_resource_group.target.name
  priority                    = 101
  direction                   = "Inbound"
  access                      = "Deny"
  description                 = "acctest target rule"
  protocol                    = "Icmp"
  source_port_range           = "1443"
  destination_port_range      = "443"
  source_address_prefix       = "10.0.0.0/8"
  destination_address_prefix  = "10.0.1.0/24"
}

resource "azurerm_resource_mover_move_resource_network_security_group" "test" {
  name               = "acctest-MR-NSG-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_security_group.test.id
  existing_target_id = azurerm_network_security_group.target.id
  depends_on_override {
    id        = azurerm_resource_group.test.id
    target_id = azurerm_resource_group.target.id
  }
  resource_setting {
    target_resource_name = "acctestRG-target-%[2]d"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}
