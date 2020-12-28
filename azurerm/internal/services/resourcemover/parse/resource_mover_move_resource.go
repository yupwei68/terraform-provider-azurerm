package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ResourceMoverMoveResourceId struct {
	SubscriptionId     string
	ResourceGroup      string
	MoveCollectionName string
	MoveResourceName   string
}

func NewResourceMoverMoveResourceID(subscriptionId, resourceGroup, moveCollectionName, moveResourceName string) ResourceMoverMoveResourceId {
	return ResourceMoverMoveResourceId{
		SubscriptionId:     subscriptionId,
		ResourceGroup:      resourceGroup,
		MoveCollectionName: moveCollectionName,
		MoveResourceName:   moveResourceName,
	}
}

func (id ResourceMoverMoveResourceId) String() string {
	segments := []string{
		fmt.Sprintf("Move Resource Name %q", id.MoveResourceName),
		fmt.Sprintf("Move Collection Name %q", id.MoveCollectionName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Resource Mover Move Resource", segmentsStr)
}

func (id ResourceMoverMoveResourceId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Migrate/moveCollections/%s/moveResources/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.MoveCollectionName, id.MoveResourceName)
}

// ResourceMoverMoveResourceID parses a ResourceMoverMoveResource ID into an ResourceMoverMoveResourceId struct
func ResourceMoverMoveResourceID(input string) (*ResourceMoverMoveResourceId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ResourceMoverMoveResourceId{
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
	if resourceId.MoveResourceName, err = id.PopSegment("moveResources"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
