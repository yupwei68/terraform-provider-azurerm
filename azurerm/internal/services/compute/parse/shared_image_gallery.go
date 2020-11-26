package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type SharedImageGalleryId struct {
	SubscriptionId string
	ResourceGroup  string
	GalleryName    string
}

func NewSharedImageGalleryID(subscriptionId, resourceGroup, galleryName string) SharedImageGalleryId {
	return SharedImageGalleryId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		GalleryName:    galleryName,
	}
}

func (id SharedImageGalleryId) ID(_ string) string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/galleries/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.GalleryName)
}

func SharedImageGalleryID(input string) (*SharedImageGalleryId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := SharedImageGalleryId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.GalleryName, err = id.PopSegment("galleries"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
