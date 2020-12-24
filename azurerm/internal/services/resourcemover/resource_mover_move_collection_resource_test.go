package resourcemover_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcemover/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type ResourceMoverMoveCollectionResource struct {
}

func TestAccResourceMoverMoveCollection_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_collection", "test")
	r := ResourceMoverMoveCollectionResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveCollection_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_collection", "test")
	r := ResourceMoverMoveCollectionResource{}
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

func TestAccResourceMoverMoveCollection_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_collection", "test")
	r := ResourceMoverMoveCollectionResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identity.0.principal_id").Exists(),
				check.That(data.ResourceName).Key("identity.0.tenant_id").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveCollection_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_collection", "test")
	r := ResourceMoverMoveCollectionResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("identity.0.principal_id").Exists(),
				check.That(data.ResourceName).Key("identity.0.tenant_id").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.update(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateRestore(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (ResourceMoverMoveCollectionResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.ResourceMoverMoveCollectionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.ResourceMover.MoveCollectionClient.Get(ctx, id.ResourceGroup, id.MoveCollectionName)
	if err != nil {
		return nil, fmt.Errorf("retrieving Resource Mover Move Resource %q (resource group: %q): %+v", id.MoveCollectionName, id.ResourceGroup, err)
	}

	return utils.Bool(resp.Properties != nil), nil
}

func (r ResourceMoverMoveCollectionResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-resource-mover-%[1]d"
  location = "%[2]s"
}

resource "azurerm_resource_mover_move_collection" "test" {
  name                = "acctest-MC-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  source_region       = "%[3]s"
  target_region       = "%[4]s"
}
`, data.RandomInteger, data.Locations.Primary, data.Locations.Secondary, data.Locations.Ternary)
}

func (r ResourceMoverMoveCollectionResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_collection" "import" {
  name                = azurerm_resource_mover_move_collection.test.name
  resource_group_name = azurerm_resource_mover_move_collection.test.resource_group_name
  location            = azurerm_resource_mover_move_collection.test.location
  source_region       = azurerm_resource_mover_move_collection.test.source_region
  target_region       = azurerm_resource_mover_move_collection.test.target_region
}
`, r.basic(data))
}

func (r ResourceMoverMoveCollectionResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-resource-mover-%[1]d"
  location = "%[2]s"
}

resource "azurerm_resource_mover_move_collection" "test" {
  name                = "acctest-MC-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  source_region       = "%[3]s"
  target_region       = "%[4]s"
  identity {
    type = "SystemAssigned"
  }

  tags = {
    Env = "Test"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.Locations.Secondary, data.Locations.Ternary)
}

func (r ResourceMoverMoveCollectionResource) update(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-resource-mover-%[1]d"
  location = "%[2]s"
}

resource "azurerm_user_assigned_identity" "test" {
  name                = "acctestusi%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

}

resource "azurerm_resource_mover_move_collection" "test" {
  name                = "acctest-MC-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  source_region       = "%[3]s"
  target_region       = "%[4]s"
  identity {
    type         = "UserAssigned"
    principal_id = azurerm_user_assigned_identity.test.principal_id
    tenant_id    = "%[5]s"
  }

  tags = {
    Env = "Stage"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.Locations.Secondary, data.Locations.Ternary, os.Getenv("ARM_TENANT_ID"))
}

func (r ResourceMoverMoveCollectionResource) updateRestore(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-resource-mover-%[1]d"
  location = "%[2]s"
}

resource "azurerm_user_assigned_identity" "test" {
  name                = "acctestusi%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location

}

resource "azurerm_resource_mover_move_collection" "test" {
  name                = "acctest-MC-%[1]d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  source_region       = "%[3]s"
  target_region       = "%[4]s"
}
`, data.RandomInteger, data.Locations.Primary, data.Locations.Secondary, data.Locations.Ternary, os.Getenv("ARM_TENANT_ID"))
}
