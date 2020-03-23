package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
)

func TestAccAzureRMManagementGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_basic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
				),
			},
			data.ImportStep(),
		},
	})
}

func TestAccAzureRMManagementGroup_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_basic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
				),
			},
			{
				Config:      testAzureRMManagementGroup_requiresImport(),
				ExpectError: acceptance.RequiresImportError("azurerm_management_group"),
			},
		},
	})
}

func TestAccAzureRMManagementGroup_nested(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_nested(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists("azurerm_management_group.parent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.child"),
				),
			},
			{
				ResourceName:      "azurerm_management_group.child",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMManagementGroup_multiLevel(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_multiLevel(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists("azurerm_management_group.grandparent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.parent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.child"),
				),
			},
			{
				ResourceName:      "azurerm_management_group.child",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAzureRMManagementGroup_multiLevelUpdated(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_nested(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists("azurerm_management_group.parent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.child"),
				),
			},
			{
				Config: testAzureRMManagementGroup_multiLevel(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists("azurerm_management_group.grandparent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.parent"),
					testCheckAzureRMManagementGroupExists("azurerm_management_group.child"),
				),
			},
		},
	})
}

func TestAccAzureRMManagementGroup_withName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_withName(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
				),
			},
		},
	})
}

func TestAccAzureRMManagementGroup_updateName(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_basic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
				),
			},
			{
				Config: testAzureRMManagementGroup_withName(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "display_name", fmt.Sprintf("acctestmg-%d", data.RandomInteger)),
				),
			},
		},
	})
}

func TestAccAzureRMManagementGroup_withSubscriptions(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_management_group", "test")
	subscriptionID := os.Getenv("ARM_SUBSCRIPTION_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMManagementGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAzureRMManagementGroup_basic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "subscription_ids.#", "0"),
				),
			},
			{
				Config: testAzureRMManagementGroup_withSubscriptions(subscriptionID),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "subscription_ids.#", "1"),
				),
			},
			{
				Config: testAzureRMManagementGroup_basic(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMManagementGroupExists(data.ResourceName),
					resource.TestCheckResourceAttr(data.ResourceName, "subscription_ids.#", "0"),
				),
			},
		},
	})
}

func testCheckAzureRMManagementGroupExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).ManagementGroups.GroupsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		groupName := rs.Primary.Attributes["group_id"]

		recurse := false
		resp, err := client.Get(ctx, groupName, "", &recurse, "", "no-cache")
		if err != nil {
			return fmt.Errorf("Bad: Get on managementGroupsClient: %s", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Management Group does not exist: %s", groupName)
		}

		return nil
	}
}

func testCheckAzureRMManagementGroupDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).ManagementGroups.GroupsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_management_group" {
			continue
		}

		name := rs.Primary.Attributes["group_id"]
		recurse := false
		resp, err := client.Get(ctx, name, "", &recurse, "", "no-cache")

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Management Group still exists: %s", *resp.Name)
		}
	}

	return nil
}

func testAzureRMManagementGroup_basic() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
}
`)
}

func testAzureRMManagementGroup_requiresImport() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
}

resource "azurerm_management_group" "import" {
  group_id = azurerm_management_group.test.group_id
}
`)
}

func testAzureRMManagementGroup_nested() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "parent" {
}

resource "azurerm_management_group" "child" {
  parent_management_group_id = azurerm_management_group.parent.id
}
`)
}

func testAzureRMManagementGroup_multiLevel() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "grandparent" {
}

resource "azurerm_management_group" "parent" {
  parent_management_group_id = azurerm_management_group.grandparent.id
}

resource "azurerm_management_group" "child" {
  parent_management_group_id = azurerm_management_group.parent.id
}
`)
}

func testAzureRMManagementGroup_withName(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  display_name = "acctestmg-%d"
}
`, data.RandomInteger)
}

// TODO: switch this out for dynamically creating a subscription once that's supported in the future
func testAzureRMManagementGroup_withSubscriptions(subscriptionID string) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_management_group" "test" {
  subscription_ids = [
    "%s",
  ]
}
`, subscriptionID)
}
