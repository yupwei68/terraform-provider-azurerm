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

type ResourceMoverMoveResourcePublicIPResource struct {
}

func TestAccResourceMoverMoveResourcePublicIP_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_public_ip", "test")
	r := ResourceMoverMoveResourcePublicIPResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourcePublicIP_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_public_ip", "test")
	r := ResourceMoverMoveResourcePublicIPResource{}
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

func TestAccResourceMoverMoveResourcePublicIP_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_public_ip", "test")
	r := ResourceMoverMoveResourcePublicIPResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourcePublicIP_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_public_ip", "test")
	r := ResourceMoverMoveResourcePublicIPResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("PreparePending"),
				check.That(data.ResourceName).Key("dependency.0.resolution_status").HasValue("Resolved"),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
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
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
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
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func TestAccResourceMoverMoveResourcePublicIP_existing(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_resource_mover_move_resource_public_ip", "test")
	r := ResourceMoverMoveResourcePublicIPResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.existing(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("error").DoesNotExist(),
				check.That(data.ResourceName).Key("move_status.0.move_state").HasValue("CommitPending"),
				//check.That(data.ResourceName).Key("target_id").HasValue(fmt.Sprintf("/subscriptions/%s/resourceGroups/acctestRG-target-resgp",os.Getenv("ARM_SUBSCRIPTION_ID"))),
			),
		},
		data.ImportStep(),
	})
}

func (ResourceMoverMoveResourcePublicIPResource) Exists(ctx context.Context, clients *clients.Client, state *terraform.InstanceState) (*bool, error) {
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

func (r ResourceMoverMoveResourcePublicIPResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_public_ip" "test" {
  name                = "acctestpublicip-%[2]d"
  location            = azurerm_resource_mover_move_collection.test.source_region
  resource_group_name = azurerm_resource_group.test.name
  allocation_method   = "Static"
  domain_name_label   = "acctestdnl%[2]d"
  sku                 = "Standard"
  zones               = ["1"]
}


`, ResourceMoverMoveResourceResourceGroupResource{}.basic(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourcePublicIPResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_public_ip" "test" {
  name               = "acctest-MR-PIP-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_public_ip.test.id
  resource_setting {
    target_resource_name = "acctestRG-target-pip"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourcePublicIPResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_resource_mover_move_resource_public_ip" "import" {
  name               = azurerm_resource_mover_move_resource_public_ip.test.name
  move_collection_id = azurerm_resource_mover_move_resource_public_ip.test.move_collection_id
  source_id          = azurerm_resource_mover_move_resource_public_ip.test.source_id
  resource_setting {
    target_resource_name = azurerm_resource_mover_move_resource_public_ip.test.resource_setting.0.target_resource_name
  }
}
`, r.basic(data))
}

func (r ResourceMoverMoveResourcePublicIPResource) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_mover_move_resource_public_ip" "test" {
  name               = "acctest-MR-PIP-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_public_ip.test.id
  resource_setting {
    target_resource_name        = "acctestRG-target2-%[2]d"
    domain_name_label           = "acctestdnltar%[2]d"
    public_ip_allocation_method = "Dynamic"
    sku                         = "Basic"
  }

  depends_on = [azurerm_resource_mover_move_resource_resource_group.test]
}
`, r.template(data), data.RandomInteger)
}

func (r ResourceMoverMoveResourcePublicIPResource) existing(data acceptance.TestData) string {
	return fmt.Sprintf(`
%[1]s

resource "azurerm_resource_group" "target" {
  name     = "acctestRG-target-%[2]d"
  location = azurerm_resource_mover_move_collection.test.target_region
}

resource "azurerm_public_ip" "target" {
  name                = "acctestpublicip-tar-%[2]d"
  location            = azurerm_resource_group.target.location
  resource_group_name = azurerm_resource_group.target.name
  allocation_method   = "Dynamic"
  domain_name_label   = "acctestdnltar%[2]d"
  sku                 = "Basic"
}

resource "azurerm_resource_mover_move_resource_public_ip" "test" {
  name               = "acctest-MR-PIP-%[2]d"
  move_collection_id = azurerm_resource_mover_move_collection.test.id
  source_id          = azurerm_public_ip.test.id
  existing_target_id = azurerm_public_ip.target.id
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
