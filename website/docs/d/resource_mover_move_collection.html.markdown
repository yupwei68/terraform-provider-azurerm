---
subcategory: "ResourceMover"
layout: "azurerm"
page_title: "Azure Resource Manager: Data Source: azurerm_resource_mover_move_collection"
description: |-
  Gets information about an existing Resource Mover Move Collection.
---

# Data Source: azurerm_resource_mover_move_collection

Use this data source to access information about an existing Resource Mover Move Collection.

## Example Usage

```hcl
data "azurerm_resource_mover_move_collection" "example" {
  name                = "existing-resource-mover-move-collection"
  resource_group_name = "existing-resgroup"
}

output "id" {
  value = data.azurerm_resource_mover_move_collection.example.id
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name of this Resource Mover Move Collection.

* `resource_group_name` - (Required) The name of the Resource Group where the Resource Mover Move Collection exists.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the Resource Mover Move Collection.

* `identity` - A `identity` block as defined below.

* `location` - The Azure Region where the Resource Mover Move Collection exists.

* `source_region` - The source region of the Resource Mover Move Collection.

* `target_region` - The target region of the Resource Mover Move Collection.

* `tags` - A mapping of tags assigned to the Resource Mover Move Collection.

---

A `identity` block exports the following:

* `principal_id` - The principal ID of the identity of the Resource Mover Move Collection.

* `tenant_id` - The tenant ID of the identity of the Resource Mover Move Collection.

* `type` - The identity type of the Resource Mover Move Collection.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Resource Mover Move Collection.
