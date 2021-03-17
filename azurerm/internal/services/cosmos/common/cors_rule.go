package common

import (
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-01-15/documentdb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func SchemaCorsRule(patchEnabled bool) *schema.Schema {
	// CorsRule "PATCH" method is only supported by blob
	allowedMethods := []string{
		"DELETE",
		"GET",
		"HEAD",
		"MERGE",
		"POST",
		"OPTIONS",
		"PUT",
	}

	if patchEnabled {
		allowedMethods = append(allowedMethods, "PATCH")
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 5,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allowed_origins": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				"exposed_headers": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				"allowed_headers": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				"allowed_methods": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},

				"max_age_in_seconds": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(1, 2000000000),
				},
			},
		},
	}
}

func ExpandCosmosCorsRule(input []interface{}) *[]documentdb.CorsPolicy {
	corsRules := make([]documentdb.CorsPolicy, 0)

	if len(input) == 0 {
		return &corsRules
	}

	for _, attr := range input {
		corsRuleAttr := attr.(map[string]interface{})
		corsRule := documentdb.CorsPolicy{}
		corsRule.AllowedOrigins = utils.String(corsRuleAttr["allowed_origins"].(string))
		corsRule.AllowedHeaders = utils.String(corsRuleAttr["allowed_headers"].(string))
		corsRule.AllowedMethods = utils.String(corsRuleAttr["allowed_headers"].(string))
		corsRule.ExposedHeaders = utils.String(corsRuleAttr["exposed_headers"].(string))
		corsRule.MaxAgeInSeconds = utils.Int64(int64(corsRuleAttr["max_age_in_seconds"].(int)))
		corsRules = append(corsRules, corsRule)
	}

	return &corsRules
}

func FlattenCosmosCorsRule(input *[]documentdb.CorsPolicy) []interface{} {
	corsRules := make([]interface{}, 0)

	if input == nil || len(*input) == 0 {
		return corsRules
	}

	for _, corsRule := range *input {
		var allowedOrigins, allowedMethods, allowedHeaders, exposedHeaders string
		var maxAgeInSeconds int
		if corsRule.AllowedOrigins != nil {
			allowedOrigins = *corsRule.AllowedOrigins
		}

		if corsRule.AllowedMethods != nil {
			allowedMethods = *corsRule.AllowedMethods
		}

		if corsRule.AllowedHeaders != nil {
			allowedHeaders = *corsRule.AllowedHeaders
		}

		if corsRule.ExposedHeaders != nil {
			exposedHeaders = *corsRule.ExposedHeaders
		}

		if corsRule.MaxAgeInSeconds != nil {
			maxAgeInSeconds = int(*corsRule.MaxAgeInSeconds)
		}

		corsRules = append(corsRules, map[string]interface{}{
			"allowed_headers":    allowedHeaders,
			"allowed_origins":    allowedOrigins,
			"allowed_methods":    allowedMethods,
			"exposed_headers":    exposedHeaders,
			"max_age_in_seconds": maxAgeInSeconds,
		})
	}

	return corsRules
}
