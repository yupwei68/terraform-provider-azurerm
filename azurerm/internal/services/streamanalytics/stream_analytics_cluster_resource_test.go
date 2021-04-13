package streamanalytics_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/streamanalytics/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"testing"
)

type StreamanalyticsClusterResource struct{}

func TestAccStreamanalyticsCluster_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_cluster", "test")
	r := StreamanalyticsClusterResource{}
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

func TestAccStreamanalyticsCluster_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_cluster", "test")
	r := StreamanalyticsClusterResource{}
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

func TestAccStreamanalyticsCluster_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_cluster", "test")
	r := StreamanalyticsClusterResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStreamanalyticsCluster_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_cluster", "test")
	r := StreamanalyticsClusterResource{}
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
			),
		},
		data.ImportStep(),
		{
			Config: r.basic(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccStreamanalyticsCluster_updateSku(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_stream_analytics_cluster", "test")
	r := StreamanalyticsClusterResource{}
	data.ResourceTest(t, r, []resource.TestStep{
		{
			Config: r.complete(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.updateSku(data),
			Check: resource.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (r StreamanalyticsClusterResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	id, err := parse.ClusterID(state.ID)
	if err != nil {
		return nil, err
	}
	resp, err := client.StreamAnalytics.ClustersClient.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Streamanalytics Cluster %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return utils.Bool(true), nil
}

func (r StreamanalyticsClusterResource) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctest-streamanalytics-%d"
  location = "%s"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r StreamanalyticsClusterResource) basic(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_cluster" "test" {
  name                = "acctest-sc-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku {
    name     = "Default"
    capacity = 36
  }
}
`, template, data.RandomInteger)
}

func (r StreamanalyticsClusterResource) requiresImport(data acceptance.TestData) string {
	config := r.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_cluster" "import" {
  name                = azurerm_stream_analytics_cluster.test.name
  resource_group_name = azurerm_stream_analytics_cluster.test.resource_group_name
  location            = azurerm_stream_analytics_cluster.test.location
}
`, config)
}

func (r StreamanalyticsClusterResource) complete(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_cluster" "test" {
  name                = "acctest-sc-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku {
    name     = "Default"
    capacity = 36
  }

  tags = {
    ENV = "Test"
  }
}
`, template, data.RandomInteger)
}

func (r StreamanalyticsClusterResource) updateSku(data acceptance.TestData) string {
	template := r.template(data)
	return fmt.Sprintf(`
%s

resource "azurerm_stream_analytics_cluster" "test" {
  name                = "acctest-sc-%d"
  resource_group_name = azurerm_resource_group.test.name
  location            = azurerm_resource_group.test.location
  sku {
    name     = "Default"
    capacity = 48
  }

  tags = {
    ENV = "Test"
  }
}
`, template, data.RandomInteger)
}
