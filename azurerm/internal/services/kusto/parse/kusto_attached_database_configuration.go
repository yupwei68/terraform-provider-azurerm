package parse

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type KustoAttachedDatabaseConfigurationId struct {
	ResourceGroup string
	Cluster       string
	Name          string
}

func KustoAttachedDatabaseConfigurationID(input string) (*KustoAttachedDatabaseConfigurationId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Unable to parse Kusto Attached Database Configuration ID %q: %+v", input, err)
	}

	configuration := KustoAttachedDatabaseConfigurationId{
		ResourceGroup: id.ResourceGroup,
	}

	if configuration.Cluster, err = id.PopSegment("Clusters"); err != nil {
		return nil, err
	}

	if configuration.Name, err = id.PopSegment("AttachedDatabaseConfigurations"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &configuration, nil
}
