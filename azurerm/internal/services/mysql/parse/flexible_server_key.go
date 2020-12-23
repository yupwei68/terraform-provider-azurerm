package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type FlexibleServerKeyId struct {
	SubscriptionId     string
	ResourceGroup      string
	FlexibleServerName string
	KeyName            string
}

func NewFlexibleServerKeyID(subscriptionId, resourceGroup, flexibleServerName, keyName string) FlexibleServerKeyId {
	return FlexibleServerKeyId{
		SubscriptionId:     subscriptionId,
		ResourceGroup:      resourceGroup,
		FlexibleServerName: flexibleServerName,
		KeyName:            keyName,
	}
}

func (id FlexibleServerKeyId) String() string {
	segments := []string{
		fmt.Sprintf("Key Name %q", id.KeyName),
		fmt.Sprintf("Flexible Server Name %q", id.FlexibleServerName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Flexible Server Key", segmentsStr)
}

func (id FlexibleServerKeyId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DBforMySQL/flexibleServers/%s/keys/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.FlexibleServerName, id.KeyName)
}

// FlexibleServerKeyID parses a FlexibleServerKey ID into an FlexibleServerKeyId struct
func FlexibleServerKeyID(input string) (*FlexibleServerKeyId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := FlexibleServerKeyId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.FlexibleServerName, err = id.PopSegment("flexibleServers"); err != nil {
		return nil, err
	}
	if resourceId.KeyName, err = id.PopSegment("keys"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
