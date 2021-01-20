package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ConfluentOrganizationId struct {
	SubscriptionId   string
	ResourceGroup    string
	OrganizationName string
}

func NewConfluentOrganizationID(subscriptionId, resourceGroup, organizationName string) ConfluentOrganizationId {
	return ConfluentOrganizationId{
		SubscriptionId:   subscriptionId,
		ResourceGroup:    resourceGroup,
		OrganizationName: organizationName,
	}
}

func (id ConfluentOrganizationId) String() string {
	segments := []string{
		fmt.Sprintf("Organization Name %q", id.OrganizationName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Confluent Organization", segmentsStr)
}

func (id ConfluentOrganizationId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Confluent/organizations/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.OrganizationName)
}

// ConfluentOrganizationID parses a ConfluentOrganization ID into an ConfluentOrganizationId struct
func ConfluentOrganizationID(input string) (*ConfluentOrganizationId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ConfluentOrganizationId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.OrganizationName, err = id.PopSegment("organizations"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}
