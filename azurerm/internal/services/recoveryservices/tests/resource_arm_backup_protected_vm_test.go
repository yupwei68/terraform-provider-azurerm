package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMBackupProtectedVm_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_backup_protected_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMBackupProtectedVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBackupProtectedVm_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMBackupProtectedVmExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "resource_group_name"),
				),
			},
			data.ImportStep(),
			{ //vault cannot be deleted unless we unregister all backups
				Config: testAccAzureRMBackupProtectedVm_base(data),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccAzureRMBackupProtectedVm_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_backup_protected_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMBackupProtectedVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBackupProtectedVm_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMBackupProtectedVmExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "resource_group_name"),
				),
			},
			data.RequiresImportErrorStep(testAccAzureRMBackupProtectedVm_requiresImport),
			{ //vault cannot be deleted unless we unregister all backups
				Config: testAccAzureRMBackupProtectedVm_base(data),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccAzureRMBackupProtectedVm_separateResourceGroups(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_backup_protected_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMBackupProtectedVmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMBackupProtectedVm_separateResourceGroups(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMBackupProtectedVmExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "resource_group_name"),
				),
			},
			data.ImportStep(),
			{ //vault cannot be deleted unless we unregister all backups
				Config: testAccAzureRMBackupProtectedVm_additionalVault(data),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccAzureRMBackupProtectedVm_updateBackupPolicyId(t *testing.T) {
	virtualMachine := "azurerm_virtual_machine.test"
	fBackupPolicyResourceName := "azurerm_backup_policy_vm.test"
	sBackupPolicyResourceName := "azurerm_backup_policy_vm.test_change_backup"
	data := acceptance.BuildTestData(t, "azurerm_backup_protected_vm", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMBackupProtectedVmDestroy,
		Steps: []resource.TestStep{
			{ // Create resources and link first backup policy id
				ResourceName: fBackupPolicyResourceName,
				Config:       testAccAzureRMBackupProtectedVm_linkFirstBackupPolicy(data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(data.ResourceName, "backup_policy_id", fBackupPolicyResourceName, "id"),
				),
			},
			{ // Modify backup policy id to the second one
				// Set Destroy false to prevent error from cleaning up dangling resource
				ResourceName: sBackupPolicyResourceName,
				Config:       testAccAzureRMBackupProtectedVm_linkSecondBackupPolicy(data),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(data.ResourceName, "backup_policy_id", sBackupPolicyResourceName, "id"),
				),
			},
			{ // Remove backup policy link
				// Backup policy link will need to be removed first so the VM's backup policy subsequently reverts to Default
				// Azure API is quite sensitive, adding the step to control resource cleanup order
				ResourceName: fBackupPolicyResourceName,
				Config:       testAccAzureRMBackupProtectedVm_withVM(data),
				Check:        resource.ComposeTestCheckFunc(),
			},
			{ // Then VM can be removed
				ResourceName: virtualMachine,
				Config:       testAccAzureRMBackupProtectedVm_withSecondPolicy(data),
				Check:        resource.ComposeTestCheckFunc(),
			},
			{ // Remove backup policies and vault
				ResourceName: data.ResourceName,
				Config:       testAccAzureRMBackupProtectedVm_basePolicyTest(data),
				Check:        resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testCheckAzureRMBackupProtectedVmDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ProtectedItemsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_backup_protected_vm" {
			continue
		}

		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		vaultName := rs.Primary.Attributes["recovery_vault_name"]
		vmId := rs.Primary.Attributes["source_vm_id"]

		parsedVmId, err := azure.ParseAzureResourceID(vmId)
		if err != nil {
			return fmt.Errorf("[ERROR] Unable to parse source_vm_id '%s': %+v", vmId, err)
		}
		vmName, hasName := parsedVmId.Path["virtualMachines"]
		if !hasName {
			return fmt.Errorf("[ERROR] parsed source_vm_id '%s' doesn't contain 'virtualMachines'", vmId)
		}

		protectedItemName := fmt.Sprintf("VM;iaasvmcontainerv2;%s;%s", parsedVmId.ResourceGroup, vmName)
		containerName := fmt.Sprintf("iaasvmcontainer;iaasvmcontainerv2;%s;%s", parsedVmId.ResourceGroup, vmName)

		resp, err := client.Get(ctx, vaultName, resourceGroup, "Azure", containerName, protectedItemName, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Recovery Services Protected VM still exists:\n%#v", resp)
	}

	return nil
}

func testCheckAzureRMBackupProtectedVmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).RecoveryServices.ProtectedItemsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %q", resourceName)
		}

		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Recovery Services Protected VM: %q", resourceName)
		}

		vaultName := rs.Primary.Attributes["recovery_vault_name"]
		vmId := rs.Primary.Attributes["source_vm_id"]

		//get VM name from id
		parsedVmId, err := azure.ParseAzureResourceID(vmId)
		if err != nil {
			return fmt.Errorf("[ERROR] Unable to parse source_vm_id '%s': %+v", vmId, err)
		}
		vmName, hasName := parsedVmId.Path["virtualMachines"]
		if !hasName {
			return fmt.Errorf("[ERROR] parsed source_vm_id '%s' doesn't contain 'virtualMachines'", vmId)
		}

		protectedItemName := fmt.Sprintf("VM;iaasvmcontainerv2;%s;%s", parsedVmId.ResourceGroup, vmName)
		containerName := fmt.Sprintf("iaasvmcontainer;iaasvmcontainerv2;%s;%s", parsedVmId.ResourceGroup, vmName)

		resp, err := client.Get(ctx, vaultName, resourceGroup, "Azure", containerName, protectedItemName, "")
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Recovery Services Protected VM %q (resource group: %q) was not found: %+v", protectedItemName, resourceGroup, err)
			}

			return fmt.Errorf("Bad: Get on recoveryServicesVaultsClient: %+v", err)
		}

		return nil
	}
}

func testAccAzureRMBackupProtectedVm_base(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-backup-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "vnet"
  location            = "${azurerm_resource_group.test.location}"
  address_space       = ["10.0.0.0/16"]
  resource_group_name = "${azurerm_resource_group.test.name}"
}

resource "azurerm_subnet" "test" {
  name                 = "acctest_subnet"
  virtual_network_name = "${azurerm_virtual_network.test.name}"
  resource_group_name  = "${azurerm_resource_group.test.name}"
  address_prefix       = "10.0.10.0/24"
}

resource "azurerm_network_interface" "test" {
  name                = "acctest_nic"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  ip_configuration {
    name                          = "acctestipconfig"
    subnet_id                     = "${azurerm_subnet.test.id}"
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = "${azurerm_public_ip.test.id}"
  }
}

resource "azurerm_public_ip" "test" {
  name                = "acctest-ip"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  allocation_method   = "Dynamic"
  domain_name_label   = "acctestip%d"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctest%s"
  location                 = "${azurerm_resource_group.test.location}"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_managed_disk" "test" {
  name                 = "acctest-datadisk"
  location             = "${azurerm_resource_group.test.location}"
  resource_group_name  = "${azurerm_resource_group.test.name}"
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1023"
}

resource "azurerm_virtual_machine" "test" {
  name                  = "acctestvm"
  location              = "${azurerm_resource_group.test.location}"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  vm_size               = "Standard_A0"
  network_interface_ids = ["${azurerm_network_interface.test.id}"]

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "acctest-osdisk"
    managed_disk_type = "Standard_LRS"
    caching           = "ReadWrite"
    create_option     = "FromImage"
  }

  storage_data_disk {
    name              = "acctest-datadisk"
    managed_disk_id   = "${azurerm_managed_disk.test.id}"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = "${azurerm_managed_disk.test.disk_size_gb}"
    create_option     = "Attach"
    lun               = 0
  }

  os_profile {
    computer_name  = "acctest"
    admin_username = "vmadmin"
    admin_password = "Password123!@#"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  boot_diagnostics {
    enabled     = true
    storage_uri = "${azurerm_storage_account.test.primary_blob_endpoint}"
  }
}

resource "azurerm_recovery_services_vault" "test" {
  name                = "acctest-%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard"

  soft_delete_enabled = false
}

resource "azurerm_backup_policy_vm" "test" {
  name                = "acctest-%d"
  resource_group_name = "${azurerm_resource_group.test.name}"
  recovery_vault_name = "${azurerm_recovery_services_vault.test.name}"

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 10
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomString, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMBackupProtectedVm_basic(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_base(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_protected_vm" "test" {
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  source_vm_id        = azurerm_virtual_machine.test.id
  backup_policy_id    = azurerm_backup_policy_vm.test.id
}
`, template)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_basePolicyTest(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-backup-%d-1"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "vnet"
  location            = azurerm_resource_group.test.location
  address_space       = ["10.0.0.0/16"]
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "acctest_subnet"
  virtual_network_name = azurerm_virtual_network.test.name
  resource_group_name  = azurerm_resource_group.test.name
  address_prefix       = "10.0.10.0/24"
}

resource "azurerm_network_interface" "test" {
  name                = "acctest_nic"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  ip_configuration {
    name                          = "acctestipconfig"
    subnet_id                     = azurerm_subnet.test.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.test.id
  }
}

resource "azurerm_public_ip" "test" {
  name                = "acctest-ip"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Dynamic"
  domain_name_label   = "acctestip%d"
}

resource "azurerm_storage_account" "test" {
  name                     = "acctest%s"
  location                 = azurerm_resource_group.test.location
  resource_group_name      = azurerm_resource_group.test.name
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_managed_disk" "test" {
  name                 = "acctest-datadisk"
  location             = azurerm_resource_group.test.location
  resource_group_name  = azurerm_resource_group.test.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1023"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomString)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_withVault(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_basePolicyTest(data)
	return fmt.Sprintf(`
%s

resource "azurerm_recovery_services_vault" "test" {
  name                = "acctest-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Standard"

  soft_delete_enabled = false
}
`, template, data.RandomInteger)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_withFirstPolicy(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_withVault(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_policy_vm" "test" {
  name                = "acctest-%d"
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 10
  }
}
`, template, data.RandomInteger)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_withSecondPolicy(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_withFirstPolicy(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_policy_vm" "test_change_backup" {
  name                = "acctest2-%d"
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 15
  }
}
`, template, data.RandomInteger)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_withVM(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_withSecondPolicy(data)
	return fmt.Sprintf(`
%s

resource "azurerm_virtual_machine" "test" {
  name                          = "acctestvm-%d"
  location                      = azurerm_resource_group.test.location
  resource_group_name           = azurerm_resource_group.test.name
  vm_size                       = "Standard_A0"
  network_interface_ids         = [azurerm_network_interface.test.id]
  delete_os_disk_on_termination = true

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "16.04-LTS"
    version   = "latest"
  }

  storage_os_disk {
    name              = "acctest-osdisk"
    managed_disk_type = "Standard_LRS"
    caching           = "ReadWrite"
    create_option     = "FromImage"
  }

  storage_data_disk {
    name              = "acctest-datadisk"
    managed_disk_id   = azurerm_managed_disk.test.id
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = azurerm_managed_disk.test.disk_size_gb
    create_option     = "Attach"
    lun               = 0
  }

  os_profile {
    computer_name  = "acctest"
    admin_username = "vmadmin"
    admin_password = "Password123!@#"
  }

  os_profile_linux_config {
    disable_password_authentication = false
  }

  boot_diagnostics {
    enabled     = true
    storage_uri = azurerm_storage_account.test.primary_blob_endpoint
  }
}
`, template, data.RandomInteger)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_linkFirstBackupPolicy(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_withVM(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_protected_vm" "test" {
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  source_vm_id        = azurerm_virtual_machine.test.id
  backup_policy_id    = azurerm_backup_policy_vm.test.id
}
`, template)
}

// For update backup policy id test
func testAccAzureRMBackupProtectedVm_linkSecondBackupPolicy(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_withVM(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_protected_vm" "test" {
  resource_group_name = azurerm_resource_group.test.name
  recovery_vault_name = azurerm_recovery_services_vault.test.name
  source_vm_id        = azurerm_virtual_machine.test.id
  backup_policy_id    = azurerm_backup_policy_vm.test_change_backup.id
}
`, template)
}

func testAccAzureRMBackupProtectedVm_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_protected_vm" "import" {
  resource_group_name = azurerm_backup_protected_vm.test.resource_group_name
  recovery_vault_name = azurerm_backup_protected_vm.test.recovery_vault_name
  source_vm_id        = azurerm_backup_protected_vm.test.source_vm_id
  backup_policy_id    = azurerm_backup_protected_vm.test.backup_policy_id
}
`, template)
}

func testAccAzureRMBackupProtectedVm_additionalVault(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_base(data)
	return fmt.Sprintf(`
%s

resource "azurerm_resource_group" "test2" {
  name     = "acctestRG-backup-%d-2"
  location = "%s"
}

resource "azurerm_recovery_services_vault" "test2" {
  name                = "acctest2-%d"
  location            = azurerm_resource_group.test2.location
  resource_group_name = azurerm_resource_group.test2.name
  sku                 = "Standard"

  soft_delete_enabled = false
}

resource "azurerm_backup_policy_vm" "test2" {
  name                = "acctest2-%d"
  resource_group_name = azurerm_resource_group.test2.name
  recovery_vault_name = azurerm_recovery_services_vault.test2.name

  backup {
    frequency = "Daily"
    time      = "23:00"
  }

  retention_daily {
    count = 10
  }
}
`, template, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMBackupProtectedVm_separateResourceGroups(data acceptance.TestData) string {
	template := testAccAzureRMBackupProtectedVm_additionalVault(data)
	return fmt.Sprintf(`
%s

resource "azurerm_backup_protected_vm" "test" {
  resource_group_name = azurerm_resource_group.test2.name
  recovery_vault_name = azurerm_recovery_services_vault.test2.name
  backup_policy_id    = azurerm_backup_policy_vm.test2.id
  source_vm_id        = azurerm_virtual_machine.test.id
}
`, template)
}
