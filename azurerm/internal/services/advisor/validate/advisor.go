package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/advisor/parse"
)

func AdvisorRecommendationID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := parse.AdvisorRecommendationID(v); err != nil {
		errors = append(errors, fmt.Errorf("can not parse %q as an Advisor Recommendation resource id: %v", k, err))
	}

	return warnings, errors
}

func AdvisorSuppressionName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[a-zA-Z0-9-_.~ ]{1,259}$`), `This is not a valid Suppression name.`)
}