package resourcemover

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type Registration struct{}

// Name is the name of this Service
func (r Registration) Name() string {
	return "ResourceMover"
}

// WebsiteCategories returns a list of categories which can be used for the sidebar
func (r Registration) WebsiteCategories() []string {
	return []string{
		"ResourceMover",
	}
}

// SupportedDataSources returns the supported Data Sources supported by this Service
func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azurerm_resource_mover_move_collection": dataSourceResourceMoverMoveCollection(),
	}
}

// SupportedResources returns the supported Resources supported by this Service
func (r Registration) SupportedResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"azurerm_resource_mover_move_collection":              resourceResourceMoverMoveCollection(),
		"azurerm_resource_mover_move_resource_resource_group": resourceResourceMoverMoveResourceResourceGroup(),
	}
}
