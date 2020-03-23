---
subcategory: "Data Lake"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_lake_store_firewall_rule"
description: |-
  Manages a Azure Data Lake Store Firewall Rule.
---

# azurerm_data_lake_store_firewall_rule

Manages a Azure Data Lake Store Firewall Rule.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "northeurope"
}

resource "azurerm_data_lake_store" "example" {
  name                = "consumptiondatalake"
  resource_group_name = azurerm_resource_group.example.name
  location            = azurerm_resource_group.example.location
}

resource "azurerm_data_lake_store_firewall_rule" "example" {
  name                = "office-ip-range"
  account_name        = azurerm_data_lake_store.example.name
  resource_group_name = azurerm_resource_group.example.name
  start_ip_address    = "1.2.3.4"
  end_ip_address      = "2.3.4.5"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Data Lake Store. Changing this forces a new resource to be created. Has to be between 3 to 24 characters.

* `resource_group_name` - (Required) The name of the resource group in which to create the Data Lake Store.

* `account_name` - (Required) Specifies the name of the Data Lake Store for which the Firewall Rule should take effect.

* `start_ip_address` - (Required) The Start IP address for the firewall rule.

* `end_ip_address` - (Required) The End IP Address for the firewall rule.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Data Lake Store Firewall Rule.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Data Lake Store Firewall Rule.
* `update` - (Defaults to 30 minutes) Used when updating the Data Lake Store Firewall Rule.
* `read` - (Defaults to 5 minutes) Used when retrieving the Data Lake Store Firewall Rule.
* `delete` - (Defaults to 30 minutes) Used when deleting the Data Lake Store Firewall Rule.

## Import

Data Lake Store Firewall Rules can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_data_lake_store_firewall_rule.rule1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.DataLakeStore/accounts/mydatalakeaccount/firewallRules/rule1
```
