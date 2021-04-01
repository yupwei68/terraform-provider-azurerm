package azure

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func SchemaZones() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func SchemaSingleZone() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		MaxItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func SchemaMultipleZones() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		MinItems: 1,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func SchemaZonesComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringIsNotEmpty,
		},
	}
}

func ExpandZones(v []interface{}) *[]string {
	zones := make([]string, 0)
	for _, zone := range v {
		zones = append(zones, zone.(string))
	}
	if len(zones) > 0 {
		return &zones
	} else {
		return nil
	}
}

func FlattenZones(v *[]string) []interface{} {
	zones := make([]interface{}, 0)
	if v == nil {
		return zones
	}

	for _, s := range *v {
		zones = append(zones, s)
	}
	return zones
}

func Track2ExpandZones(v []interface{}) *[]*string {
	zones := make([]*string, 0)
	for _, zone := range v {
		zones = append(zones, utils.String(zone.(string)))
	}
	if len(zones) > 0 {
		return &zones
	} else {
		return nil
	}
}

func Track2FlattenZones(v *[]*string) []interface{} {
	zones := make([]interface{}, 0)
	if v == nil {
		return zones
	}

	for _, s := range *v {
		if s != nil {
			zones = append(zones, *s)
		}
	}
	return zones
}
