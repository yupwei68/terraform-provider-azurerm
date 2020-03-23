---
subcategory: "Base"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_resource_group"
description: |-
    Manages a resource group on Azure.
---

# azurerm_resource_group

Manages a resource group on Azure.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "testResourceGroup1"
  location = "West US"

  tags = {
    environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the resource group. Must be unique on your
    Azure subscription.

* `location` - (Required) The location where the resource group should be created.
    For a list of all Azure locations, please consult [this link](http://azure.microsoft.com/en-us/regions/) or run `az account list-locations --output table`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - The ID of the Resource Group.


## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 90 minutes) Used when creating the Resource Group.
* `update` - (Defaults to 90 minutes) Used when updating the Resource Group.
* `read` - (Defaults to 5 minutes) Used when retrieving the Resource Group.
* `delete` - (Defaults to 90 minutes) Used when deleting the Resource Group.

## Import

Resource Groups can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_resource_group.mygroup /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/myresourcegroup
```
