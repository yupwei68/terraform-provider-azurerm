---
subcategory: "ResourceMover"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_resource_mover_move_collection"
description: |-
  Manages a Resource Mover Move Collection.
---

# azurerm_resource_mover_move_collection

Manages a Resource Mover Move Collection.

## Example Usage

```hcl

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "East US 2"
}

resource "azurerm_resource_mover_move_collection" "example" {
  name                = "example-mc"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
  source_region       = "eastus"
  target_region       = "westus"
  tags = {
    "foo" = "bar"
  }
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this Resource Mover Move Collection. Changing this forces a new Resource Mover Move Collection to be created.

* `resource_group_name` - (Required) The name of the Resource Group where the Resource Mover Move Collection should exist. Changing this forces a new Resource Mover Move Collection to be created.

* `location` - (Required) The Azure Region where the Resource Mover Move Collection should exist. Changing this forces a new Resource Mover Move Collection to be created.

* `source_region` - (Required) The source region of the Resource Mover Move Collection. Changing this forces a new Resource Mover Move Collection to be created.

* `target_region` - (Required) The target region of the Resource Mover Move Collection. Changing this forces a new Resource Mover Move Collection to be created.

---

* `identity` - (Optional) A `identity` block as defined below.

* `tags` - (Optional) A mapping of tags which should be assigned to the Resource Mover Move Collection.

---

A `identity` block supports the following:

* `principal_id` - (Optional) The principal ID of the identity of the Resource Mover Move Collection.

* `tenant_id` - (Optional) The tenant ID of the identity of the Resource Mover Move Collection.

* `type` - (Optional) The identity type of the Resource Mover Move Collection. Possible values are `SystemAssigned` and `UserAssigned`.


## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the Resource Mover Move Collection.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Resource Mover Move Collection.
* `read` - (Defaults to 5 minutes) Used when retrieving the Resource Mover Move Collection.
* `update` - (Defaults to 30 minutes) Used when updating the Resource Mover Move Collection.
* `delete` - (Defaults to 30 minutes) Used when deleting the Resource Mover Move Collection.

## Import

Resource Mover Move Collections can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_resource_mover_move_collection.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.Migrate/moveCollections/moveCollection1
```
