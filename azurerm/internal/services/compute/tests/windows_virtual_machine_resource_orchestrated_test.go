package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccAzureRMWindowsVirtualMachine_orchestratedZonal(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMWindowsVirtualMachine_orchestratedZonal(data),
				Check: resource.ComposeTestCheckFunc(
					checkWindowsVirtualMachineExists(data.ResourceName),
				),
			},
			data.ImportStep("admin_password"),
		},
	})
}

func TestAccAzureRMWindowsVirtualMachine_orchestratedZonalWithProximityPlacementGroup(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMWindowsVirtualMachine_orchestratedZonalWithProximityPlacementGroup(data),
				Check: resource.ComposeTestCheckFunc(
					checkWindowsVirtualMachineExists(data.ResourceName),
				),
			},
			data.ImportStep("admin_password"),
		},
	})
}

func TestAccAzureRMWindowsVirtualMachine_orchestratedNonZonal(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMWindowsVirtualMachine_orchestratedNonZonal(data),
				Check: resource.ComposeTestCheckFunc(
					checkWindowsVirtualMachineExists(data.ResourceName),
				),
			},
			data.ImportStep("admin_password"),
		},
	})
}

func TestAccAzureRMWindowsVirtualMachine_orchestratedMultipleZonal(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMWindowsVirtualMachine_orchestratedMultipleZonal(data),
				Check: resource.ComposeTestCheckFunc(
					checkWindowsVirtualMachineExists(data.ResourceName),
				),
			},
			data.ImportStep("admin_password"),
		},
	})
}

func TestAccAzureRMWindowsVirtualMachine_orchestratedMultipleNoneZonal(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_windows_virtual_machine", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: checkWindowsVirtualMachineIsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMWindowsVirtualMachine_orchestratedMultipleNonZonal(data),
				Check: resource.ComposeTestCheckFunc(
					checkWindowsVirtualMachineExists(data.ResourceName),
				),
			},
			data.ImportStep("admin_password"),
		},
	})
}

func testAccAzureRMWindowsVirtualMachine_orchestratedZonal(data acceptance.TestData) string {
	template := testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data)
	return fmt.Sprintf(`
%s

resource "azurerm_network_interface" "test" {
  name                = "acctestnic-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_orchestrated_virtual_machine_scale_set" "test" {
  name                = "acctestVMO-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  platform_fault_domain_count = 1

  zones = ["1"]

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_windows_virtual_machine" "test" {
  name                = local.vm_name
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.test.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
  zone                         = azurerm_orchestrated_virtual_machine_scale_set.test.zones.0
}
`, template, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMWindowsVirtualMachine_orchestratedZonalWithProximityPlacementGroup(data acceptance.TestData) string {
	template := testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data)
	return fmt.Sprintf(`
%s

resource "azurerm_proximity_placement_group" "test" {
  name                = "acctestPPG-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
}

resource "azurerm_network_interface" "test" {
  name                = "acctestnic-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_orchestrated_virtual_machine_scale_set" "test" {
  name                = "acctestVMO-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  platform_fault_domain_count = 1

  proximity_placement_group_id = azurerm_proximity_placement_group.test.id

  zones = ["1"]

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_windows_virtual_machine" "test" {
  name                = local.vm_name
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"

  proximity_placement_group_id = azurerm_proximity_placement_group.test.id

  network_interface_ids = [
    azurerm_network_interface.test.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
  zone                         = azurerm_orchestrated_virtual_machine_scale_set.test.zones.0
}
`, template, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMWindowsVirtualMachine_orchestratedNonZonal(data acceptance.TestData) string {
	template := testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data)
	return fmt.Sprintf(`
%s

resource "azurerm_network_interface" "test" {
  name                = "acctestnic-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_orchestrated_virtual_machine_scale_set" "test" {
  name                = "acctestVMO-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  platform_fault_domain_count = 2

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_windows_virtual_machine" "test" {
  name                = local.vm_name
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.test.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
}
`, template, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMWindowsVirtualMachine_orchestratedMultipleZonal(data acceptance.TestData) string {
	template := testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data)
	return fmt.Sprintf(`
%s

resource "azurerm_orchestrated_virtual_machine_scale_set" "test" {
  name                = "acctestVMO-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  platform_fault_domain_count = 1

  zones = ["1"]

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_network_interface" "first" {
  name                = "acctestnic1-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface" "second" {
  name                = "acctestnic2-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_windows_virtual_machine" "test" {
  name                = "accVM1%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.first.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
  zone                         = azurerm_orchestrated_virtual_machine_scale_set.test.zones.0
}

resource "azurerm_windows_virtual_machine" "another" {
  name                = "accVM2%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.second.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2019-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
  zone                         = azurerm_orchestrated_virtual_machine_scale_set.test.zones.0
}
`, template, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomIntOfLength(9), data.RandomIntOfLength(9))
}

func testAccAzureRMWindowsVirtualMachine_orchestratedMultipleNonZonal(data acceptance.TestData) string {
	template := testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data)
	return fmt.Sprintf(`
%s

resource "azurerm_orchestrated_virtual_machine_scale_set" "test" {
  name                = "acctestVMO-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  platform_fault_domain_count = 2

  tags = {
    ENV = "Test"
  }
}

resource "azurerm_network_interface" "first" {
  name                = "acctestnic1-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_network_interface" "second" {
  name                = "acctestnic2-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
  }
}

resource "azurerm_windows_virtual_machine" "test" {
  name                = "accVM1%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.first.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2016-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
}

resource "azurerm_windows_virtual_machine" "another" {
  name                = "accVM2%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  size                = "Standard_F2"
  admin_username      = "adminuser"
  admin_password      = "P@ssw0rd1234!"
  network_interface_ids = [
    azurerm_network_interface.second.id,
  ]

  source_image_reference {
    publisher = "MicrosoftWindowsServer"
    offer     = "WindowsServer"
    sku       = "2019-Datacenter"
    version   = "latest"
  }

  os_disk {
    storage_account_type = "Standard_LRS"
    caching              = "ReadWrite"
  }

  virtual_machine_scale_set_id = azurerm_orchestrated_virtual_machine_scale_set.test.id
}
`, template, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomIntOfLength(9), data.RandomIntOfLength(9))
}

func testWindowsVirtualMachine_templateBaseForOchestratedVMSS(data acceptance.TestData) string {
	return fmt.Sprintf(`
locals {
  vm_name = "acctestvm%s"
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestnw-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}
`, data.RandomString, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
