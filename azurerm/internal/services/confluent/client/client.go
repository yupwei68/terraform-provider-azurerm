package client

import (
	"github.com/Azure/azure-sdk-for-go/services/preview/confluent/mgmt/2020-03-01-preview/confluent"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/common"
)

type Client struct {
	OrganizationClient *confluent.OrganizationClient
}

func NewClient(o *common.ClientOptions) *Client {
	organizationClient := confluent.NewOrganizationClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
	o.ConfigureClient(&organizationClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		OrganizationClient: &organizationClient,
	}
}
