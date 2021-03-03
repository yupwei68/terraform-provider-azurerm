package sentinel

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/preview/securityinsight/mgmt/2019-01-01-preview/securityinsight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/sentinel/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
)

func importSentinelDataConnector(expectKind securityinsight.DataConnectorKind) func(d *schema.ResourceData, meta interface{}) (data []*schema.ResourceData, err error) {
	return func(d *schema.ResourceData, meta interface{}) (data []*schema.ResourceData, err error) {
		id, err := parse.DataConnectorID(d.Id())
		if err != nil {
			return nil, err
		}

		client := meta.(*clients.Client).Sentinel.DataConnectorsClient
		ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
		defer cancel()

		resp, err := client.Get(ctx, id.ResourceGroup, OperationalInsightsResourceProvider, id.WorkspaceName, id.Name)
		if err != nil {
			return nil, fmt.Errorf("retrieving Sentinel Alert Rule %q: %+v", id, err)
		}

		if err := assertDataConnectorKind(resp.Value, expectKind); err != nil {
			return nil, err
		}
		return []*schema.ResourceData{d}, nil
	}
}

func assertDataConnectorKind(dc securityinsight.BasicDataConnector, expectKind securityinsight.DataConnectorKind) error {
	var kind securityinsight.DataConnectorKind
	switch dc.(type) {
	case securityinsight.AADDataConnector:
		kind = securityinsight.DataConnectorKindAzureActiveDirectory
	case securityinsight.ASCDataConnector:
		kind = securityinsight.DataConnectorKindAzureSecurityCenter
	case securityinsight.MCASDataConnector:
		kind = securityinsight.DataConnectorKindMicrosoftCloudAppSecurity
	case securityinsight.TIDataConnector:
		kind = securityinsight.DataConnectorKindThreatIntelligence
	case securityinsight.OfficeDataConnector:
		kind = securityinsight.DataConnectorKindOffice365
	case securityinsight.OfficeATPDataConnector:
		kind = securityinsight.DataConnectorKindOfficeATP
	case securityinsight.AwsCloudTrailDataConnector:
		kind = securityinsight.DataConnectorKindAmazonWebServicesCloudTrail
	case securityinsight.MDATPDataConnector:
		kind = securityinsight.DataConnectorKindMicrosoftDefenderAdvancedThreatProtection
	}
	if expectKind != kind {
		return fmt.Errorf("Sentinel Data Connector has mismatched kind, expected: %q, got %q", expectKind, kind)
	}
	return nil
}
