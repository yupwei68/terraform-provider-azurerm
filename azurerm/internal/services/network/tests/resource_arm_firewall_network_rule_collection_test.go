package tests

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
)

func TestAccAzureRMFirewallNetworkRuleCollection_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMFirewallNetworkRuleCollection_requiresImport(data),
				ExpectError: acceptance.RequiresImportError("azurerm_firewall_network_rule_collection"),
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_updatedName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.3765122797.name", "rule1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_updatedName(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.1700340761.name", "rule2"),
				),
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_multipleRuleCollections(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")
	secondRule := "azurerm_firewall_network_rule_collection.test_add"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_multiple(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					testCheckAzureRMFirewallNetworkRuleCollectionExists(secondRule),
					resource.TestCheckResourceAttr(secondRule, "name", "acctestnrc_add"),
					resource.TestCheckResourceAttr(secondRule, "priority", "200"),
					resource.TestCheckResourceAttr(secondRule, "action", "Deny"),
					resource.TestCheckResourceAttr(secondRule, "rule.#", "1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					testCheckAzureRMFirewallNetworkRuleCollectionDoesNotExist("azurerm_firewall.test", "acctestnrc_add"),
				),
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")
	secondResourceName := "azurerm_firewall_network_rule_collection.test_add"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_multiple(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					testCheckAzureRMFirewallNetworkRuleCollectionExists(secondResourceName),
					resource.TestCheckResourceAttr(secondResourceName, "name", "acctestnrc_add"),
					resource.TestCheckResourceAttr(secondResourceName, "priority", "200"),
					resource.TestCheckResourceAttr(secondResourceName, "action", "Deny"),
					resource.TestCheckResourceAttr(secondResourceName, "rule.#", "1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_multipleUpdate(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "300"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Deny"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					testCheckAzureRMFirewallNetworkRuleCollectionExists(secondResourceName),
					resource.TestCheckResourceAttr(secondResourceName, "name", "acctestnrc_add"),
					resource.TestCheckResourceAttr(secondResourceName, "priority", "400"),
					resource.TestCheckResourceAttr(secondResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(secondResourceName, "rule.#", "1"),
				),
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_disappears(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
					testCheckAzureRMFirewallNetworkRuleCollectionDisappears(data.ResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_multipleRules(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_multipleRules(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "2"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
		},
	})
}

func TestAccAzureRMFirewallNetworkRuleCollection_updateFirewallTags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_firewall_network_rule_collection", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
			{
				Config: testAccAzureRMFirewallNetworkRuleCollection_updateFirewallTags(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMFirewallNetworkRuleCollectionExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "name", "acctestnrc"),
					resource.TestCheckResourceAttr(data.ResourceName, "priority", "100"),
					resource.TestCheckResourceAttr(data.ResourceName, "action", "Allow"),
					resource.TestCheckResourceAttr(data.ResourceName, "rule.#", "1"),
				),
			},
		},
	})
}

func testCheckAzureRMFirewallNetworkRuleCollectionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Network.AzureFirewallsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		firewallName := rs.Primary.Attributes["azure_firewall_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		read, err := client.Get(ctx, resourceGroup, firewallName)
		if err != nil {
			return err
		}

		found := false
		for _, collection := range *read.AzureFirewallPropertiesFormat.NetworkRuleCollections {
			if *collection.Name == name {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Expected Network Rule Collection %q (Firewall %q / Resource Group %q) to exist but it didn't", name, firewallName, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMFirewallNetworkRuleCollectionDoesNotExist(resourceName string, collectionName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Network.AzureFirewallsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		firewallName := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		read, err := client.Get(ctx, resourceGroup, firewallName)
		if err != nil {
			return err
		}

		for _, collection := range *read.AzureFirewallPropertiesFormat.NetworkRuleCollections {
			if *collection.Name == collectionName {
				return fmt.Errorf("Network Rule Collection %q exists in Firewall %q: %+v", collectionName, firewallName, collection)
			}
		}

		return nil
	}
}

func testCheckAzureRMFirewallNetworkRuleCollectionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Network.AzureFirewallsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		name := rs.Primary.Attributes["name"]
		firewallName := rs.Primary.Attributes["azure_firewall_name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		read, err := client.Get(ctx, resourceGroup, firewallName)
		if err != nil {
			return err
		}

		rules := make([]network.AzureFirewallNetworkRuleCollection, 0)
		for _, collection := range *read.AzureFirewallPropertiesFormat.NetworkRuleCollections {
			if *collection.Name != name {
				rules = append(rules, collection)
			}
		}

		read.AzureFirewallPropertiesFormat.NetworkRuleCollections = &rules

		future, err := client.CreateOrUpdate(ctx, resourceGroup, firewallName, read)
		if err != nil {
			return fmt.Errorf("Error removing Network Rule Collection from Firewall: %+v", err)
		}

		if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
			return fmt.Errorf("Error waiting for the removal of Network Rule Collection from Firewall: %+v", err)
		}

		_, err = client.Get(ctx, resourceGroup, firewallName)
		return err
	}
}

func testAccAzureRMFirewallNetworkRuleCollection_basic(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 100
  action              = "Allow"

  rule {
    name = "rule1"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_requiresImport(data acceptance.TestData) string {
	template := testAccAzureRMFirewallNetworkRuleCollection_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "import" {
  name                = azurerm_firewall_network_rule_collection.test.name
  azure_firewall_name = azurerm_firewall_network_rule_collection.test.azure_firewall_name
  resource_group_name = azurerm_firewall_network_rule_collection.test.resource_group_name
  priority            = 100
  action              = "Allow"

  rule {
    name = "rule1"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_updatedName(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 100
  action              = "Allow"

  rule {
    name = "rule2"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_multiple(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 100
  action              = "Allow"

  rule {
    name = "acctestrule"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}

resource "azurerm_firewall_network_rule_collection" "test_add" {
  name                = "acctestnrc_add"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 200
  action              = "Deny"

  rule {
    name = "acctestruleadd"

    source_addresses = [
      "10.0.0.0/8",
    ]

    destination_ports = [
      "8080",
    ]

    destination_addresses = [
      "8.8.4.4",
    ]

    protocols = [
      "TCP",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_multipleUpdate(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 300
  action              = "Deny"

  rule {
    name = "acctestrule"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}

resource "azurerm_firewall_network_rule_collection" "test_add" {
  name                = "acctestnrc_add"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 400
  action              = "Allow"

  rule {
    name = "acctestruleadd"

    source_addresses = [
      "10.0.0.0/8",
    ]

    destination_ports = [
      "8080",
    ]

    destination_addresses = [
      "8.8.4.4",
    ]

    protocols = [
      "TCP",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_multipleRules(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 100
  action              = "Allow"

  rule {
    name = "acctestrule"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }

  rule {
    name = "acctestrule_add"

    source_addresses = [
      "192.168.0.1",
    ]

    destination_ports = [
      "8888",
    ]

    destination_addresses = [
      "1.1.1.1",
    ]

    protocols = [
      "TCP",
    ]
  }
}
`, template)
}

func testAccAzureRMFirewallNetworkRuleCollection_updateFirewallTags(data acceptance.TestData) string {
	template := testAccAzureRMFirewall_withTags(data)
	return fmt.Sprintf(`
%s

resource "azurerm_firewall_network_rule_collection" "test" {
  name                = "acctestnrc"
  azure_firewall_name = azurerm_firewall.test.name
  resource_group_name = azurerm_resource_group.test.name
  priority            = 100
  action              = "Allow"

  rule {
    name = "rule1"

    source_addresses = [
      "10.0.0.0/16",
    ]

    destination_ports = [
      "53",
    ]

    destination_addresses = [
      "8.8.8.8",
    ]

    protocols = [
      "Any",
    ]
  }
}
`, template)
}
