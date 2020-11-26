package tests

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMTemplateDeployment_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_basicMultiple(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
				),
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_basicMultiple(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
				),
			},
			data.RequiresImportErrorStep(testAccAzureRMTemplateDeployment_requiresImport),
		},
	})
}

func TestAccAzureRMTemplateDeployment_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_basicSingle(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
					testCheckAzureRMTemplateDeploymentDisappears(data.ResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_nestedTemplate(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_nestedTemplate(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
				),
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_withParams(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_withParams(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
					resource.TestCheckResourceAttr("azurerm_template_deployment.test", "outputs.testOutput", "Output Value"),
				),
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_withParamsBody(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccAzureRMTemplateDeployment_withParamsBody(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
					resource.TestCheckResourceAttr("azurerm_template_deployment.test", "outputs.testOutput", "Output Value"),
				),
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_withOutputs(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMTemplateDeployment_withOutputs(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMTemplateDeploymentExists(data.ResourceName),
					resource.TestCheckOutput("tfIntOutput", "-123"),
					resource.TestCheckOutput("tfStringOutput", "Standard_GRS"),

					// these values *should* be 'true' and 'false' but,
					// due to a bug in the way terraform represents bools at various times these are for now 0 and 1
					// see https://github.com/hashicorp/terraform/issues/13512#issuecomment-295389523
					// at a later date these may return the expected 'true' / 'false' and should be changed back
					resource.TestCheckOutput("tfFalseOutput", "false"),
					resource.TestCheckOutput("tfTrueOutput", "true"),
					resource.TestCheckResourceAttr(data.ResourceName, "outputs.stringOutput", "Standard_GRS"),
				),
			},
		},
	})
}

func TestAccAzureRMTemplateDeployment_withError(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_template_deployment", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMTemplateDeploymentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAzureRMTemplateDeployment_withError(data),
				ExpectError: regexp.MustCompile("Error waiting for deployment"),
			},
		},
	})
}

func testCheckAzureRMTemplateDeploymentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Resource.DeploymentsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for template deployment: %s", name)
		}

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on deploymentsClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: TemplateDeployment %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMTemplateDeploymentDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Resource.DeploymentsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		deploymentName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for template deployment: %s", deploymentName)
		}

		if _, err := client.Delete(ctx, resourceGroup, deploymentName); err != nil {
			return fmt.Errorf("Failed deleting Deployment %q (Resource Group %q): %+v", deploymentName, resourceGroup, err)
		}

		// we can't use the Waiter here since the API returns a 200 once it's deleted which is considered a polling status code..
		log.Printf("[DEBUG] Waiting for Template Deployment (%q in Resource Group %q) to be deleted", deploymentName, resourceGroup)
		stateConf := &resource.StateChangeConf{
			Pending: []string{"200"},
			Target:  []string{"404"},
			Timeout: 40 * time.Minute,
			Refresh: func() (interface{}, string, error) {
				res, err := client.Get(ctx, resourceGroup, deploymentName)

				log.Printf("Retrieving Template Deployment %q (Resource Group %q) returned Status %d", resourceGroup, deploymentName, res.StatusCode)

				if err != nil {
					if utils.ResponseWasNotFound(res.Response) {
						return res, strconv.Itoa(res.StatusCode), nil
					}
					return nil, "", fmt.Errorf("Error polling for the status of the Template Deployment %q (RG: %q): %+v", deploymentName, resourceGroup, err)
				}

				return res, strconv.Itoa(res.StatusCode), nil
			},
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("Error waiting for Template Deployment (%q in Resource Group %q) to be deleted: %+v", deploymentName, resourceGroup, err)
		}

		return nil
	}
}

func testCheckAzureRMTemplateDeploymentDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Resource.DeploymentsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_template_deployment" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Template Deployment still exists:\n%#v", resp.Properties)
		}
	}

	return nil
}

func testAccAzureRMTemplateDeployment_basicSingle(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "variables": {
    "location": "[resourceGroup().location]",
    "publicIPAddressType": "Dynamic",
    "apiVersion": "2015-06-15",
    "dnsLabelPrefix": "[concat('terraform-tdacctest', uniquestring(resourceGroup().id))]"
  },
  "resources": [
     {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "[variables('apiVersion')]",
      "name": "acctestpip-%d",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[variables('dnsLabelPrefix')]"
        }
      }
    }
  ]
}
DEPLOY


  deployment_mode = "Complete"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMTemplateDeployment_basicMultiple(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "storageAccountType": {
      "type": "string",
      "defaultValue": "Standard_LRS",
      "allowedValues": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_ZRS"
      ],
      "metadata": {
        "description": "Storage Account type"
      }
    }
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "storageAccountName": "[concat(uniquestring(resourceGroup().id), 'storage')]",
    "publicIPAddressName": "[concat('myPublicIp', uniquestring(resourceGroup().id))]",
    "publicIPAddressType": "Dynamic",
    "apiVersion": "2015-06-15",
    "dnsLabelPrefix": "[concat('terraform-tdacctest', uniquestring(resourceGroup().id))]"
  },
  "resources": [
    {
      "type": "Microsoft.Storage/storageAccounts",
      "name": "[variables('storageAccountName')]",
      "apiVersion": "[variables('apiVersion')]",
      "location": "[variables('location')]",
      "properties": {
        "accountType": "[parameters('storageAccountType')]"
      }
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "[variables('apiVersion')]",
      "name": "[variables('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[variables('dnsLabelPrefix')]"
        }
      }
    }
  ]
}
DEPLOY


  deployment_mode = "Complete"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func testAccAzureRMTemplateDeployment_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMTemplateDeployment_basicMultiple(data)
	return fmt.Sprintf(`
%s

resource "azurerm_template_deployment" "import" {
  name                = azurerm_template_deployment.test.name
  resource_group_name = azurerm_template_deployment.test.resource_group_name

  template_body   = azurerm_template_deployment.test.template_body
  deployment_mode = azurerm_template_deployment.test.deployment_mode
}
`, template)
}

func testAccAzureRMTemplateDeployment_nestedTemplate(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "resourceGroupName": "[resourceGroup().name]"
  },
  "resources": [
    {
      "apiVersion": "2017-05-10",
      "name": "nested-template-deployment",
      "type": "Microsoft.Resources/deployments",
      "resourceGroup": "[variables('resourceGroupName')]",
      "properties": {
        "mode": "Incremental",
        "template": {
          "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
          "contentVersion": "1.0.0.0",
          "variables": {
            "location": "[variables('location')]",
            "resourceGroupName": "[variables('resourceGroupName')]"
          },
          "resources": [
            {
              "type": "Microsoft.Network/publicIPAddresses",
              "apiVersion": "2015-06-15",
              "name": "acctest-pip",
              "location": "[variables('location')]",
              "properties": {
                "publicIPAllocationMethod": "Dynamic"
              }
            }
          ]
        }
      }
    }
  ]
}
DEPLOY


  deployment_mode = "Complete"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func testaccAzureRMTemplateDeployment_withParamsBody(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_client_config" "current" {}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

output "test" {
  value = azurerm_template_deployment.test.outputs["testOutput"]
}

resource "azurerm_storage_container" "using-outputs" {
  name                  = "vhds"
  storage_account_name  = azurerm_template_deployment.test.outputs["accountName"]
  container_access_type = "private"
}


resource "azurerm_key_vault" "test" {
  location            = "%s"
  name                = "vault%d"
  resource_group_name = "${azurerm_resource_group.test.name}"

  sku_name = "standard"

  tenant_id                       = data.azurerm_client_config.current.tenant_id
  enabled_for_template_deployment = true

  access_policy {
    key_permissions = []
    object_id       = data.azurerm_client_config.current.object_id

    secret_permissions = [
      "delete",
      "get",
      "list",
      "set",
      "purge",
    ]

    tenant_id = "${data.azurerm_client_config.current.tenant_id}"
  }
}

resource "azurerm_key_vault_secret" "test-secret" {
  name         = "acctestsecret-%d"
  value        = "terraform-test-%d"
  key_vault_id = azurerm_key_vault.test.id
}

locals {
  templated-file = <<TPL
{
"dnsLabelPrefix": {
    "reference": {
      "keyvault": {
        "id": "${azurerm_key_vault.test.id}"
      },
      "secretName": "${azurerm_key_vault_secret.test-secret.name}"
    }
  },
"storageAccountType": {
   "value": "Standard_GRS"
  }
}
TPL
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "storageAccountType": {
      "type": "string",
      "defaultValue": "Standard_LRS",
      "allowedValues": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_ZRS"
      ],
      "metadata": {
        "description": "Storage Account type"
      }
    },
    "dnsLabelPrefix": {
      "type": "string",
      "metadata": {
        "description": "DNS Label for the Public IP. Must be lowercase. It should match with the following regular expression: ^[a-z][a-z0-9-]{1,61}[a-z0-9]$ or it will raise an error."
      }
    }
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "storageAccountName": "[concat(uniquestring(resourceGroup().id), 'storage')]",
    "publicIPAddressName": "[concat('myPublicIp', uniquestring(resourceGroup().id))]",
    "publicIPAddressType": "Dynamic",
    "apiVersion": "2015-06-15"
  },
  "resources": [
    {
      "type": "Microsoft.Storage/storageAccounts",
      "name": "[variables('storageAccountName')]",
      "apiVersion": "[variables('apiVersion')]",
      "location": "[variables('location')]",
      "properties": {
        "accountType": "[parameters('storageAccountType')]"
      }
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "[variables('apiVersion')]",
      "name": "[variables('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[parameters('dnsLabelPrefix')]"
        }
      }
    }
  ],
  "outputs": {
    "testOutput": {
      "type": "string",
      "value": "Output Value"
    },
    "accountName": {
      "type": "string",
      "value": "[variables('storageAccountName')]"
    }
  }
}
DEPLOY

  parameters_body = "${local.templated-file}"
  deployment_mode = "Incremental"
  depends_on      = ["azurerm_key_vault_secret.test-secret"]
}
`, data.RandomInteger, data.Locations.Primary, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMTemplateDeployment_withParams(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

output "test" {
  value = azurerm_template_deployment.test.outputs["testOutput"]
}

resource "azurerm_storage_container" "using-outputs" {
  name                  = "vhds"
  storage_account_name  = azurerm_template_deployment.test.outputs["accountName"]
  container_access_type = "private"
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "storageAccountType": {
      "type": "string",
      "defaultValue": "Standard_LRS",
      "allowedValues": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_ZRS"
      ],
      "metadata": {
        "description": "Storage Account type"
      }
    },
    "dnsLabelPrefix": {
      "type": "string",
      "metadata": {
        "description": "DNS Label for the Public IP. Must be lowercase. It should match with the following regular expression: ^[a-z][a-z0-9-]{1,61}[a-z0-9]$ or it will raise an error."
      }
    }
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "storageAccountName": "[concat(uniquestring(resourceGroup().id), 'storage')]",
    "publicIPAddressName": "[concat('myPublicIp', uniquestring(resourceGroup().id))]",
    "publicIPAddressType": "Dynamic",
    "apiVersion": "2015-06-15"
  },
  "resources": [
    {
      "type": "Microsoft.Storage/storageAccounts",
      "name": "[variables('storageAccountName')]",
      "apiVersion": "[variables('apiVersion')]",
      "location": "[variables('location')]",
      "properties": {
        "accountType": "[parameters('storageAccountType')]"
      }
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "[variables('apiVersion')]",
      "name": "[variables('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[parameters('dnsLabelPrefix')]"
        }
      }
    }
  ],
  "outputs": {
    "testOutput": {
      "type": "string",
      "value": "Output Value"
    },
    "accountName": {
      "type": "string",
      "value": "[variables('storageAccountName')]"
    }
  }
}
DEPLOY


  parameters = {
    dnsLabelPrefix     = "terraform-test-%d"
    storageAccountType = "Standard_GRS"
  }

  deployment_mode = "Complete"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func testAccAzureRMTemplateDeployment_withOutputs(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

output "tfStringOutput" {
  value = azurerm_template_deployment.test.outputs["stringOutput"]
}

output "tfIntOutput" {
  value = azurerm_template_deployment.test.outputs["intOutput"]
}

output "tfFalseOutput" {
  value = azurerm_template_deployment.test.outputs["falseOutput"]
}

output "tfTrueOutput" {
  value = azurerm_template_deployment.test.outputs["trueOutput"]
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "storageAccountType": {
      "type": "string",
      "defaultValue": "Standard_LRS",
      "allowedValues": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_ZRS"
      ],
      "metadata": {
        "description": "Storage Account type"
      }
    },
    "dnsLabelPrefix": {
      "type": "string",
      "metadata": {
        "description": "DNS Label for the Public IP. Must be lowercase. It should match with the following regular expression: ^[a-z][a-z0-9-]{1,61}[a-z0-9]$ or it will raise an error."
      }
    },
    "intParameter": {
      "type": "int",
      "defaultValue": -123
    },
    "falseParameter": {
      "type": "bool",
      "defaultValue": false
    },
    "trueParameter": {
      "type": "bool",
      "defaultValue": true
    }
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "storageAccountName": "[concat(uniquestring(resourceGroup().id), 'storage')]",
    "publicIPAddressName": "[concat('myPublicIp', uniquestring(resourceGroup().id))]",
    "publicIPAddressType": "Dynamic",
    "apiVersion": "2015-06-15"
  },
  "resources": [
    {
      "type": "Microsoft.Storage/storageAccounts",
      "name": "[variables('storageAccountName')]",
      "apiVersion": "[variables('apiVersion')]",
      "location": "[variables('location')]",
      "properties": {
        "accountType": "[parameters('storageAccountType')]"
      }
    },
    {
      "type": "Microsoft.Network/publicIPAddresses",
      "apiVersion": "[variables('apiVersion')]",
      "name": "[variables('publicIPAddressName')]",
      "location": "[variables('location')]",
      "properties": {
        "publicIPAllocationMethod": "[variables('publicIPAddressType')]",
        "dnsSettings": {
          "domainNameLabel": "[parameters('dnsLabelPrefix')]"
        }
      }
    }
  ],
  "outputs": {
    "stringOutput": {
      "type": "string",
      "value": "[parameters('storageAccountType')]"
    },
    "intOutput": {
      "type": "int",
      "value": "[parameters('intParameter')]"
    },
    "falseOutput": {
      "type": "bool",
      "value": "[parameters('falseParameter')]"
    },
    "trueOutput": {
      "type": "bool",
      "value": "[parameters('trueParameter')]"
    }
  }
}
DEPLOY


  parameters = {
    dnsLabelPrefix     = "terraform-test-%d"
    storageAccountType = "Standard_GRS"
  }

  deployment_mode = "Incremental"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

// StorageAccount name is too long, forces error
func testAccAzureRMTemplateDeployment_withError(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

output "test" {
  value = azurerm_template_deployment.test.outputs["testOutput"]
}

resource "azurerm_template_deployment" "test" {
  name                = "acctesttemplate-%d"
  resource_group_name = azurerm_resource_group.test.name

  template_body = <<DEPLOY
{
  "$schema": "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "storageAccountType": {
      "type": "string",
      "defaultValue": "Standard_LRS",
      "allowedValues": [
        "Standard_LRS",
        "Standard_GRS",
        "Standard_ZRS"
      ],
      "metadata": {
        "description": "Storage Account type"
      }
    }
  },
  "variables": {
    "location": "[resourceGroup().location]",
    "storageAccountName": "badStorageAccountNameTooLong",
    "apiVersion": "2015-06-15"
  },
  "resources": [
    {
      "type": "Microsoft.Storage/storageAccounts",
      "name": "[variables('storageAccountName')]",
      "apiVersion": "[variables('apiVersion')]",
      "location": "[variables('location')]",
      "properties": {
        "accountType": "[parameters('storageAccountType')]"
      }
    }
  ],
  "outputs": {
    "testOutput": {
      "type": "string",
      "value": "Output Value"
    }
  }
}
DEPLOY


  parameters = {
    storageAccountType = "Standard_GRS"
  }

  deployment_mode = "Complete"
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}
