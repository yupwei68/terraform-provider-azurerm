package datacatalog

import (
	"github.com/Azure/azure-sdk-for-go/services/datacatalog/mgmt/2016-03-30/datacatalog"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	BuildCatalogsClient func(string) *datacatalog.ADCCatalogsClient
}

func BuildClient(o *common.ClientOptions) *Client {
	BuildCatalogsClient := func(catalogNmae string) *datacatalog.ADCCatalogsClient {
		client := datacatalog.NewADCCatalogsClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId, catalogNmae)
		o.ConfigureClient(&client.Client, o.ResourceManagerAuthorizer)
		return &client
	}

	return &Client{
		BuildCatalogsClient: BuildCatalogsClient,
	}
}
