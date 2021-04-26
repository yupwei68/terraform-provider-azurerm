package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/arm/avs/2020-03-20/armavs"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	PrivateCloudClient *armavs.PrivateCloudsClient
}

func NewClient(o *common.ClientOptions) *Client {

	return &Client{
		PrivateCloudClient: armavs.NewPrivateCloudsClient(o.ResourceManagerConnection, o.SubscriptionId),
	}
}
