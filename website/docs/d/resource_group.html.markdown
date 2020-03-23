---
subcategory: "Base"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_resource_group"
description: |-
  Gets information about an existing Resource Group.
---

# Data Source: azurerm_resource_group

Use this data source to access information about an existing Resource Group.

## Example Usage

```hcl
data "azurerm_resource_group" "example" {
  name = "dsrg_test"
}

resource "azurerm_managed_disk" "example" {
  name                 = "managed_disk_name"
  location             = data.azurerm_resource_group.example.location
  resource_group_name  = data.azurerm_resource_group.example.name
  storage_account_type = "Standard_LRS"
  create_option        = "Empty"
  disk_size_gb         = "1"
}
```

## Argument Reference

* `name` - Specifies the name of the resource group.

~> **Note:** If the specified location doesn't match the actual resource group location, an error message with the actual location value will be shown.

## Attributes Reference

* `id` - The ID of the Resource Group.
* `location` - The location of the resource group.
* `tags` - A mapping of tags assigned to the resource group.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Resource Group.
