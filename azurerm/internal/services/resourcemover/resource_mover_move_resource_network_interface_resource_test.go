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

type ResourceMoverMoveResourceNetworkInterfaceResource struct {
}

func TestAccResourceMoverMoveResourceNetworkInterface_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_interface", "test")
	r := ResourceMoverMoveResourceNetworkInterfaceResource{}
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

func TestAccResourceMoverMoveResourceNetworkInterface_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_interface", "test")
	r := ResourceMoverMoveResourceNetworkInterfaceResource{}
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

func TestAccResourceMoverMoveResourceNetworkInterface_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_interface", "test")
	r := ResourceMoverMoveResourceNetworkInterfaceResource{}
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

func TestAccResourceMoverMoveResourceNetworkInterface_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_interface", "test")
	r := ResourceMoverMoveResourceNetworkInterfaceResource{}
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

func TestAccResourceMoverMoveResourceNetworkInterface_existing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_network_interface", "test")
	r := ResourceMoverMoveResourceNetworkInterfaceResource{}
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

func (ResourceMoverMoveResourceNetworkInterfaceResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
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

func (r ResourceMoverMoveResourceNetworkInterfaceResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_network_interface" "test" {
  name                          = "acctestni-%[2]d"
  location                      = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name           = azurerm_resource_group.test.name
  enable_accelerated_networking = false

  ip_configuration {
    name                          = "primary"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    primary                       = true
  }
}
`, ResourceMoverMoveResourceVirtualNetworkResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkInterfaceResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_network_interface" "test" {
  name               = "acctest-MR-NI-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_interface.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-ni"
  }

  depends_on = [azurerm_resource_mover_move_resource_virtual_network.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkInterfaceResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_network_interface" "import" {
  name               = azurerm_resource_mover_move_resource_network_interface.test.name
  move_collection_id = azurerm_resource_mover_move_resource_network_interface.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_network_interface.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_network_interface.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourceNetworkInterfaceResource) lbTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_virtual_network" "test" {
  name                = "acctestvn-%[2]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "testsubnet"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_public_ip" "test" {
  name                = "test-ip-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurerm_lb" "test" {
  name                = "acctestlb-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name

  frontend_ip_configuration {
    name                 = "primary"
    public_ip_address_id = azurerm_public_ip.test.id
  }
}

resource "azurerm_lb_backend_address_pool" "test" {
  resource_group_name = azurerm_resource_group.test.name
  loadbalancer_id     = azurerm_lb.test.id
  name                = "acctestpool"
}

resource "azurerm_network_interface" "test" {
  name                = "acctestni-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface_backend_address_pool_association" "test" {
  network_interface_id    = azurerm_network_interface.test.id
  ip_configuration_name   = "testconfiguration1"
  backend_address_pool_id = azurerm_lb_backend_address_pool.test.id
}

resource "azurerm_resource_mover_move_resource_virtual_network" "test" {
  name               = "acctest-MR-VN-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_network.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-vn"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, ResourceMoverMoveResourceResourceGroupResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkInterfaceResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_network_interface" "test" {
  name               = "acctest-MR-NI-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_interface.test.id

  resource_setting {
    target_resource_name          = "acctest-target-ni"
    enable_accelerated_networking = true

    ip_configuration {
      name = "primary-target"

      load_balancer_backend_address_pool {
        name = "acctest-tar-lb-ap"
        id   = azurerm_lb_backend_address_pool.test.id
      }

      primary                      = true
      private_ip_address           = "10.0.2.2"
      private_ip_allocation_method = "Static"

      subnet {
        name = "primary-sn"
        id   = azurerm_subnet.test.id
      }
    }

    ip_configuration {
      name                         = "secondary"
      private_ip_allocation_method = "Dynamic"
      primary                      = false

      subnet {
        name = "secondary-sn"
        id   = azurerm_subnet.test.id
      }
    }
  }

  depends_on = [azurerm_resource_mover_move_resource_virtual_network.test]
}
`, r.lbTemplate(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceNetworkInterfaceResource) existing(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = azurerm_resource_mover_move_collection.test.target_region
}

resource "azurerm_virtual_network" "target" {
  name                = "acctestvirtnettar%[2]d"
  address_space       = ["10.1.0.0/16"]
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
}

resource "azurerm_subnet" "target" {
  name                 = "internaltarget"
  resource_group_name  = azurerm_resource_group.target.name
  virtual_network_name = azurerm_virtual_network.target.name
  address_prefix       = "10.1.1.0/24"
}

resource "azurerm_public_ip" "target" {
  name                = "acctest-tar-ip-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
  allocation_method   = "Static"
}

resource "azurerm_lb" "target" {
  name                = "acctesttarlb-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name

  frontend_ip_configuration {
    name                 = "primary"
    public_ip_address_id = azurerm_public_ip.target.id
  }
}

resource "azurerm_lb_backend_address_pool" "target" {
  resource_group_name = azurerm_resource_group.target.name
  loadbalancer_id     = azurerm_lb.target.id
  name                = "acctesttarpool"
}

resource "azurerm_network_interface" "target" {
  name                          = "acctestni-target-%[2]d"
  location                      = azurerm_resource_group.target.location
  resource_group_name           = azurerm_resource_group.target.name
  enable_accelerated_networking = true

  ip_configuration {
    name                          = "target"
    subnet_id                     = azurerm_subnet.target.id
    private_ip_address_allocation = "Static"
    private_ip_address            = "10.1.1.34"
  }
}

resource "azurerm_network_interface_backend_address_pool_association" "target" {
  network_interface_id    = azurerm_network_interface.target.id
  ip_configuration_name   = "target"
  backend_address_pool_id = azurerm_lb_backend_address_pool.target.id
}

resource "azurerm_resource_mover_move_resource_network_interface" "test" {
  name               = "acctest-MR-NI-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_interface.test.id
  existing_target_id = azurerm_network_interface.target.id

  depends_on_override {
    id        = azurerm_resource_group.test.id
    target_id = azurerm_resource_group.target.id
  }

  depends_on_override {
    id        = azurerm_virtual_network.test.id
    target_id = azurerm_virtual_network.target.id
  }

  depends_on_override {
    id        = azurerm_public_ip.test.id
    target_id = azurerm_public_ip.target.id
  }

  depends_on_override {
    id        = azurerm_lb.test.id
    target_id = azurerm_lb.target.id
  }

  resource_setting {
    target_resource_name = "acctestRG-target-%[2]d"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.lbTemplate(data), data.RandomInteger)
}
