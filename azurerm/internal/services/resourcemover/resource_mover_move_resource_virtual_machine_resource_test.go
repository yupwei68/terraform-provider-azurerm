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

type ResourceMoverMoveResourceVirtualMachineResource struct {
}

func TestAccResourceMoverMoveResourceVirtualMachine_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_machine", "test")
	r := ResourceMoverMoveResourceVirtualMachineResource{}
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

func TestAccResourceMoverMoveResourceVirtualMachine_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_machine", "test")
	r := ResourceMoverMoveResourceVirtualMachineResource{}
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

func TestAccResourceMoverMoveResourceVirtualMachine_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_machine", "test")
	r := ResourceMoverMoveResourceVirtualMachineResource{}
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

func TestAccResourceMoverMoveResourceVirtualMachine_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_machine", "test")
	r := ResourceMoverMoveResourceVirtualMachineResource{}
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

func TestAccResourceMoverMoveResourceVirtualMachine_existing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_virtual_machine", "test")
	r := ResourceMoverMoveResourceVirtualMachineResource{}
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

func (ResourceMoverMoveResourceVirtualMachineResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
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

func (r ResourceMoverMoveResourceVirtualMachineResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_virtual_network" "test" {
  name                = "acctvn-%[2]d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name

}

resource "azurerm_subnet" "test" {
  name                 = "acctsub-%[2]d"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_public_ip" "test" {
  name                = "acctestpublicip-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
}

resource "azurerm_network_interface" "test" {
  name                = "acctni-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.test.id
  }
}

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

resource "azurerm_subnet_network_security_group_association" "test" {
  subnet_id                 = azurerm_subnet.test.id
  network_security_group_id = azurerm_network_security_group.test.id
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctvm-%[2]d"
  location              = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name   = azurerm_resource_group.test.name
  network_interface_ids = [azurerm_network_interface.test.id]
  vm_size               = "Standard_DS1_v2"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    create_option     = "FromImage"
    caching           = "ReadWrite"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "host01"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_resource_mover_move_resource_virtual_network" "test" {
  name               = "acctest-MR-VN-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_network.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-vn"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test,azurerm_resource_mover_move_resource_network_security_group.test]
}

resource "azurerm_resource_mover_move_resource_public_ip" "test" {
  name               = "acctest-MR-PIP-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_public_ip.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-pip"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}

resource "azurerm_resource_mover_move_resource_network_security_group" "test" {
  name               = "acctest-MR-NSG-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_security_group.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-nsg"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}

resource "azurerm_resource_mover_move_resource_network_interface" "test" {
  name               = "acctest-MR-NI-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_network_interface.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-ni"
  }

  depends_on = [azurerm_resource_mover_move_resource_virtual_network.test, azurerm_resource_mover_move_resource_public_ip.test, azurerm_resource_mover_move_resource_network_security_group.test]
}
`, ResourceMoverMoveResourceResourceGroupResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualMachineResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_virtual_machine" "test" {
  name               = "acctest-MR-VM-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_machine.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-vm"
target_vm_size       = "Standard_B2s"
  }

  depends_on = [azurerm_resource_mover_move_resource_network_interface.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualMachineResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_virtual_machine" "import" {
  name               = azurerm_resource_mover_move_resource_virtual_machine.test.name
  move_collection_id = azurerm_resource_mover_move_resource_virtual_machine.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_virtual_machine.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_virtual_machine.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourceVirtualMachineResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_virtual_machine" "test" {
  name               = "acctest-MR-VM-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_machine.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-vm"
    target_vm_size       = "Standard_B2s"
  }

  depends_on = [azurerm_resource_mover_move_resource_network_interface.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceVirtualMachineResource) existing(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = azurerm_resource_mover_move_collection.test.target_region
}

resource "azurerm_virtual_network" "target" {
  name                = "acctvn-target-%[2]d"
  address_space       = ["10.1.0.0/16"]
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
}

resource "azurerm_subnet" "target" {
  name                 = "acctsub-target-%[2]d"
  resource_group_name  = azurerm_resource_group.target.name
  virtual_network_name = azurerm_virtual_network.target.name
  address_prefix       = "10.1.2.0/24"
}

resource "azurerm_public_ip" "target" {
  name                = "acctestpublicip-target-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
  allocation_method   = "Static"
}

resource "azurerm_network_interface" "target" {
  name                = "acctni-target-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name

  ip_configuration {
    name                          = "testconfiguration1"
    subnet_id                     = azurerm_subnet.target.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.target.id
  }
}

resource "azurerm_network_security_group" "target" {
  name                = "acctest-nsg-target-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
}

resource "azurerm_network_security_rule" "target" {
  name                        = "acctest-nsg-rule-target-%[2]d"
  network_security_group_name = azurerm_network_security_group.target.name
  resource_group_name         = azurerm_resource_group.target.name
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

resource "azurerm_subnet_network_security_group_association" "target" {
  subnet_id                 = azurerm_subnet.target.id
  network_security_group_id = azurerm_network_security_group.target.id
}

resource "azurerm_virtual_machine" "target" {
  name                  = "acctvm-target-%[2]d"
  location              = azurerm_resource_group.target.location
  resource_group_name   = azurerm_resource_group.target.name
  network_interface_ids = [azurerm_network_interface.target.id]
  vm_size               = "Standard_B2s"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "18.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "myosdisk1"
    create_option     = "FromImage"
    caching           = "ReadWrite"
    managed_disk_type = "Standard_LRS"
  }

  os_profile {
    computer_name  = "host01"
    admin_username = "testadmin"
    admin_password = "Password1234!"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }
}

resource "azurerm_resource_mover_move_resource_virtual_machine" "test" {
  name               = "acctest-MR-VM-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_virtual_machine.test.id
  existing_target_id = azurerm_virtual_machine.target.id

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
    id        = azurerm_network_security_group.test.id
    target_id = azurerm_network_security_group.target.id
  }

  depends_on_override {
    id        = azurerm_network_interface.test.id
    target_id = azurerm_network_interface.target.id
  }

  resource_setting {
    target_resource_name = "acctestRG-target-%[2]d"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}
