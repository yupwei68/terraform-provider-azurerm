package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/resourcemover/mgmt/2019-10-01-preview/resourcemover"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	MoveCollectionClient *resourcemover.MoveCollectionsClient
	MoveResourceClient   *resourcemover.MoveResourcesClient
}

func NewClient(o *common.ClientOptions) *Client {
	moveCollectionClient := resourcemover.NewMoveCollectionsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&moveCollectionClient.Client, o.ResourceManagerAuthorizer)

	moveResourceClient := resourcemover.NewMoveResourcesClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&moveResourceClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		MoveCollectionClient: &moveCollectionClient,
		MoveResourceClient:   &moveResourceClient,
	}
}
