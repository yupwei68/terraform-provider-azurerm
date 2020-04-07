---
subcategory: "Data Share"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_share_account"
description: |-
  Manages a Data Share Account.
---

# azurerm_data_share_account

Manages a Data Share Account.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_data_share_account" "example" {
  name                = "example-dsa"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  tags = {
    foo = "bar"
  }
}
```

## Arguments Reference

The following arguments are supported:

* `location` - (Required) The Azure Region where the Data Share Account should exist. Changing this forces a new Data Share Account to be created.

* `name` - (Required) The name which should be used for this Data Share Account. Changing this forces a new Data Share Account to be created.

* `resource_group_name` - (Required) The name of the Resource Group where the Data Share Account should exist. Changing this forces a new Data Share Account to be created.

---

* `tags` - (Optional) A mapping of tags which should be assigned to the Data Share Account.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the Data Share Account.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Data Share Account.
* `read` - (Defaults to 5 minutes) Used when retrieving the Data Share Account.
* `update` - (Defaults to 30 minutes) Used when updating the Data Share Account.
* `delete` - (Defaults to 30 minutes) Used when deleting the Data Share Account.

## Import

Data Share Accounts can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_data_share_account.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.DataShare/accounts/account1
```