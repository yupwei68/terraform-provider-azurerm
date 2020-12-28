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

type ResourceMoverMoveResourceResourceGroupResource struct {
}

func TestAccResourceMoverMoveResourceResourceGroup_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_resource_group", "test")
	r := ResourceMoverMoveResourceResourceGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").Exists(),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourceResourceGroup_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_resource_group", "test")
	r := ResourceMoverMoveResourceResourceGroupResource{}
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

func TestAccResourceMoverMoveResourceResourceGroup_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_resource_group", "test")
	r := ResourceMoverMoveResourceResourceGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").Exists(),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourceResourceGroup_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_resource_group", "test")
	r := ResourceMoverMoveResourceResourceGroupResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").Exists(),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").Exists(),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").Exists(),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func (ResourceMoverMoveResourceResourceGroupResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
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

func (r ResourceMoverMoveResourceResourceGroupResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_subscription" "primary" {}

data "azurerm_role_definition" "contributor" {
  name = "Contributor"
}

resource "azurerm_role_assignment" "contributor" {
  scope              = data.azurerm_subscription.primary.id
  role_definition_id = "${data.azurerm_subscription.primary.id}${data.azurerm_role_definition.contributor.id}"
  principal_id       = azurerm_resource_mover_move_collection.test.identity.0.principal_id
}

data "azurerm_role_definition" "user_access" {
  name = "User Access Administrator"
}

resource "azurerm_role_assignment" "user_access" {
  scope              = data.azurerm_subscription.primary.id
  role_definition_id = "${data.azurerm_subscription.primary.id}${data.azurerm_role_definition.user_access.id}"
  principal_id       = azurerm_resource_mover_move_collection.test.identity.0.principal_id
}

`, ResourceMoverMoveCollectionResource{}.complete(data))
}

func (r ResourceMoverMoveResourceResourceGroupResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_resource_group" "test" {
  name               = "acctest-MR-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_resource_group.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-resgp"
  }

  depends_on = [azurerm_role_assignment.contributor, azurerm_role_assignment.user_access]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourceResourceGroupResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_resource_group" "import" {
  name               = azurerm_resource_mover_move_resource_resource_group.test.name
  move_collection_id = azurerm_resource_mover_move_resource_resource_group.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_resource_group.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_resource_group.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourceResourceGroupResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = "%[3]s"
}

resource "azurerm_resource_mover_move_resource_resource_group" "test" {
  name               = "acctest-MR-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_resource_group.test.id
  existing_target_id = azurerm_resource_group.target.id
  resource_setting {
    target_resource_name = "acctestRG-target-%[2]d"
  }

  depends_on = [azurerm_role_assignment.contributor, azurerm_role_assignment.user_access]
}
`, r.template(data), data.RandomInteger, data.Locations.Ternary)
}
