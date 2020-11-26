package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type DataSetId struct {
	SubscriptionId string
	ResourceGroup  string
	AccountName    string
	ShareName      string
	Name           string
}

func NewDataSetID(subscriptionId, resourceGroup, accountName, shareName, name string) DataSetId {
	return DataSetId{
		SubscriptionId: subscriptionId,
		ResourceGroup:  resourceGroup,
		AccountName:    accountName,
		ShareName:      shareName,
		Name:           name,
	}
}

func (id DataSetId) ID(_ string) string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DataShare/accounts/%s/shares/%s/dataSets/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.AccountName, id.ShareName, id.Name)
}

func DataSetID(input string) (*DataSetId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := DataSetId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.AccountName, err = id.PopSegment("accounts"); err != nil {
		return nil, err
	}
	if resourceId.ShareName, err = id.PopSegment("shares"); err != nil {
		return nil, err
	}
	if resourceId.Name, err = id.PopSegment("dataSets"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
