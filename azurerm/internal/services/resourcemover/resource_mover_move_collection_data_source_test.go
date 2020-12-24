package resourcemover_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
)

type ResourceMoverMoveCollectionDataSource struct {
}

func TestAccDataSourceMoverMoveCollection_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_resource_mover_move_collection", "test")
	r := ResourceMoverMoveCollectionDataSource{}

	data.DataSourceTest(t, []resource.TestStep{
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("location").Exists(),
				check.That(data.ResourceName).Key("source_region").Exists(),
				check.That(data.ResourceName).Key("target_region").Exists(),
			),
		},
	})
}

func (ResourceMoverMoveCollectionDataSource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

data "azurerm_resource_mover_move_collection" "test" {
  name                = azurerm_resource_mover_move_collection.test.name
  resource_group_name = azurerm_resource_mover_move_collection.test.resource_group_name
}
`, ResourceMoverMoveCollectionResource{}.basic(data))
}
