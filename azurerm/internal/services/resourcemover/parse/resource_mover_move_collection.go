package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ResourceMoverMoveCollectionId struct {
	SubscriptionId     string
	ResourceGroup      string
	MoveCollectionName string
}

func NewResourceMoverMoveCollectionID(subscriptionId, resourceGroup, moveCollectionName string) ResourceMoverMoveCollectionId {
	return ResourceMoverMoveCollectionId{
		SubscriptionId:     subscriptionId,
		ResourceGroup:      resourceGroup,
		MoveCollectionName: moveCollectionName,
	}
}

func (id ResourceMoverMoveCollectionId) String() string {
	segments := []string{
		fmt.Sprintf("Move Collection Name %q", id.MoveCollectionName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Resource Mover Move Collection", segmentsStr)
}

func (id ResourceMoverMoveCollectionId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Migrate/moveCollections/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.MoveCollectionName)
}

// ResourceMoverMoveCollectionID parses a ResourceMoverMoveCollection ID into an ResourceMoverMoveCollectionId struct
func ResourceMoverMoveCollectionID(input string) (*ResourceMoverMoveCollectionId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ResourceMoverMoveCollectionId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.MoveCollectionName, err = id.PopSegment("moveCollections"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
