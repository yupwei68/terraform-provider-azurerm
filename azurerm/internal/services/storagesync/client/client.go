package client

import (
	"github.com/Azure/azure-sdk-for-go/services/storagesync/mgmt/2020-03-01/storagesync"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	StorageSyncServiceClient    *storagesync.ServicesClient
}

func NewClient(options *common.ClientOptions) *Client {
	storageSyncServiceClient := storagesync.NewServicesClientWithBaseURI(options.ResourceManagerEndpoint, options.SubscriptionId)
	options.ConfigureClient(&storageSyncServiceClient.Client, options.ResourceManagerAuthorizer)

	return &Client{
		StorageSyncServiceClient:     &storageSyncServiceClient,
	}
}

