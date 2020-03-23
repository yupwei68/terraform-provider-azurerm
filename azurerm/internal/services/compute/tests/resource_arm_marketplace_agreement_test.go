package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMMarketplaceAgreement(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"basic": {
			"basic":             testAccAzureRMMarketplaceAgreement_basic,
			"requiresImport":    testAccAzureRMMarketplaceAgreement_requiresImport,
			"agreementCanceled": testAccAzureRMMarketplaceAgreement_agreementCanceled,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAzureRMMarketplaceAgreement_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_marketplace_agreement", "test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMMarketplaceAgreementDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMMarketplaceAgreement_basicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMMarketplaceAgreementExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "license_text_link"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "privacy_policy_link"),
				),
			},
			data.ImportStep(),
		},
	})
}

func testAccAzureRMMarketplaceAgreement_requiresImport(t *testing.T) {
	if !features.ShouldResourcesBeImported() {
		t.Skip("Skipping since resources aren't required to be imported")
		return
	}

	data := acceptance.BuildTestData(t, "azurerm_marketplace_agreement", "test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMMarketplaceAgreementDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMMarketplaceAgreement_basicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMMarketplaceAgreementExists(data.ResourceName),
				),
			},
			{
				Config:      testAccAzureRMMarketplaceAgreement_requiresImportConfig(),
				ExpectError: acceptance.RequiresImportError("azurerm_marketplace_agreement"),
			},
		},
	})
}

func testAccAzureRMMarketplaceAgreement_agreementCanceled(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_marketplace_agreement", "test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMMarketplaceAgreementDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMMarketplaceAgreement_basicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMMarketplaceAgreementExists(data.ResourceName),
					testCheckAzureRMMarketplaceAgreementCanceled(data.ResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckAzureRMMarketplaceAgreementExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Compute.MarketplaceAgreementsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		offer := rs.Primary.Attributes["offer"]
		plan := rs.Primary.Attributes["plan"]
		publisher := rs.Primary.Attributes["publisher"]

		resp, err := client.Get(ctx, publisher, offer, plan)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Marketplace Agreement for Publisher %q / Offer %q / Plan %q does not exist", publisher, offer, plan)
			}
			return fmt.Errorf("Bad: Get on MarketplaceAgreementsClient: %+v", err)
		}

		if resp.AgreementProperties == nil || resp.AgreementProperties.Accepted == nil || !*resp.AgreementProperties.Accepted {
			return fmt.Errorf("Bad: Marketplace Agreement for Publisher %q / Offer %q / Plan %q is not accepted", publisher, offer, plan)
		}

		return nil
	}
}

func testCheckAzureRMMarketplaceAgreementCanceled(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := acceptance.AzureProvider.Meta().(*clients.Client).Compute.MarketplaceAgreementsClient
		ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		offer := rs.Primary.Attributes["offer"]
		plan := rs.Primary.Attributes["plan"]
		publisher := rs.Primary.Attributes["publisher"]

		resp, err := client.Cancel(ctx, publisher, offer, plan)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Marketplace Agreement for Publisher %q / Offer %q / Plan %q does not exist", publisher, offer, plan)
			}
			return fmt.Errorf("Bad: Get on MarketplaceAgreementsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMMarketplaceAgreementDestroy(s *terraform.State) error {
	client := acceptance.AzureProvider.Meta().(*clients.Client).Compute.MarketplaceAgreementsClient
	ctx := acceptance.AzureProvider.Meta().(*clients.Client).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_marketplace_agreement" {
			continue
		}

		offer := rs.Primary.Attributes["offer"]
		plan := rs.Primary.Attributes["plan"]
		publisher := rs.Primary.Attributes["publisher"]

		resp, err := client.Get(ctx, publisher, offer, plan)
		if err != nil {
			if !utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Marketplace Agreement still exists:\n%#v", resp)
			}
		}
	}

	return nil
}

func testAccAzureRMMarketplaceAgreement_basicConfig() string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_marketplace_agreement" "test" {
  publisher = "barracudanetworks"
  offer     = "waf"
  plan      = "hourly"
}
`)
}

func testAccAzureRMMarketplaceAgreement_requiresImportConfig() string {
	template := testAccAzureRMMarketplaceAgreement_basicConfig()
	return fmt.Sprintf(`
%s

resource "azurerm_marketplace_agreement" "import" {
  publisher = azurerm_marketplace_agreement.test.publisher
  offer     = azurerm_marketplace_agreement.test.offer
  plan      = azurerm_marketplace_agreement.test.plan
}
`, template)
}
