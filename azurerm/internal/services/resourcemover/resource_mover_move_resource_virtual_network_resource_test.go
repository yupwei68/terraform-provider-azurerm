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

type ResourceMoverMoveResourceVirtualNetworkResource struct {
}

func TestAccResourceMoverMoveResourceVirtualNetwork_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_network", "test")
	r := ResourceMoverMoveResourceVirtualNetworkResource{}
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

func TestAccResourceMoverMoveResourceVirtualNetwork_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_network", "test")
	r := ResourceMoverMoveResourceVirtualNetworkResource{}
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

func TestAccResourceMoverMoveResourceVirtualNetwork_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_network", "test")
	r := ResourceMoverMoveResourceVirtualNetworkResource{}
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

func TestAccResourceMoverMoveResourceVirtualNetwork_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_network", "test")
	r := ResourceMoverMoveResourceVirtualNetworkResource{}
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

func TestAccResourceMoverMoveResourceVirtualNetwork_existing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_network", "test")
	r := ResourceMoverMoveResourceVirtualNetworkResource{}
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

func (ResourceMoverMoveResourceVirtualNetworkResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
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

func (r ResourceMoverMoveResourceVirtualNetworkResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_virtual_network" "test" {
  name                = "acctestvirtnet%[2]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
  dns_servers         = ["10.7.7.2", "10.7.7.7", "10.7.7.1"]
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.1.0/24"
}
`, ResourceMoverMoveResourceResourceGroupResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualNetworkResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_virtual_network" "test" {
  name               = "acctest-MR-VN-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_network.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-vn"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualNetworkResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_virtual_network" "import" {
  name               = azurerm_resource_mover_move_resource_virtual_network.test.name
  move_collection_id = azurerm_resource_mover_move_resource_virtual_network.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_virtual_network.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_virtual_network.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourceVirtualNetworkResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_virtual_network" "test" {
  name               = "acctest-MR-VN-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_network.test.id
  resource_setting {
    target_resource_name   = "acctestRG-target-vn"
    address_spaces         = ["10.1.0.0/16"]
    dns_servers            = ["10.10.10.1"]
    enable_ddos_protection = true
    subnet {
      name           = "acctest-target-SN"
      address_prefix = "10.1.1.0/24"
    }
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualNetworkResource) existing(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = azurerm_resource_mover_move_collection.test.target_region
}

resource "azurerm_network_ddos_protection_plan" "target" {
  name                = "acctestddospplan-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
}

resource "azurerm_virtual_network" "target" {
  name                = "acctestvirtnettar%[2]d"
  address_space       = ["10.1.0.0/16"]
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
  dns_servers         = ["10.10.10.1"]
  ddos_protection_plan {
    id     = azurerm_network_ddos_protection_plan.target.id
    enable = true
  }
}

resource "azurerm_subnet" "target" {
  name                 = "internaltarget"
  resource_group_name  = azurerm_resource_group.target.name
  virtual_network_name = azurerm_virtual_network.target.name
  address_prefix       = "10.1.1.0/24"
}

resource "azurerm_resource_mover_move_resource_virtual_network" "test" {
  name               = "acctest-MR-VN-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_network.test.id
  existing_target_id = azurerm_virtual_network.target.id
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
