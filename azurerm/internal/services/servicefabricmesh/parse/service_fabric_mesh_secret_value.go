package parse

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ServiceFabricMeshSecretValueId struct {
	ResourceGroup string
	SecretName    string
	Name          string
}

func ServiceFabricMeshSecretValueID(input string) (*ServiceFabricMeshSecretValueId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Unable to parse Service Fabric Mesh Secret ID %q: %+v", input, err)
	}

	value := ServiceFabricMeshSecretValueId{
		ResourceGroup: id.ResourceGroup,
	}

	if value.SecretName, err = id.PopSegment("secrets"); err != nil {
		return nil, err
	}

	if value.Name, err = id.PopSegment("values"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &value, nil
}
