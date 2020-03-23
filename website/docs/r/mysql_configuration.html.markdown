---
subcategory: "Database"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_mysql_configuration"
description: |-
  Sets a MySQL Configuration value on a MySQL Server.
---

# azurerm_mysql_configuration

Sets a MySQL Configuration value on a MySQL Server.

## Disclaimers

~> **Note:** Since this resource is provisioned by default, the Azure Provider will not check for the presence of an existing resource prior to attempting to create it.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "api-rg-pro"
  location = "West Europe"
}

resource "azurerm_mysql_server" "example" {
  name                = "mysql-server-1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku_name = "GP_Gen5_2"

  storage_profile {
    storage_mb            = 5120
    backup_retention_days = 7
    geo_redundant_backup  = "Disabled"
  }

  administrator_login          = "psqladminun"
  administrator_login_password = "H@Sh1CoR3!"
  version                      = "5.7"
  ssl_enforcement              = "Enabled"
}

resource "azurerm_mysql_configuration" "example" {
  name                = "interactive_timeout"
  resource_group_name = azurerm_resource_group.example.name
  server_name         = azurerm_mysql_server.example.name
  value               = "600"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the MySQL Configuration, which needs [to be a valid MySQL configuration name](https://dev.mysql.com/doc/refman/5.7/en/server-configuration.html). Changing this forces a new resource to be created.

* `server_name` - (Required) Specifies the name of the MySQL Server. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which the MySQL Server exists. Changing this forces a new resource to be created.

* `value` - (Required) Specifies the value of the MySQL Configuration. See the MySQL documentation for valid values.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the MySQL Configuration.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the MySQL Configuration.
* `update` - (Defaults to 30 minutes) Used when updating the MySQL Configuration.
* `read` - (Defaults to 5 minutes) Used when retrieving the MySQL Configuration.
* `delete` - (Defaults to 30 minutes) Used when deleting the MySQL Configuration.

## Import

MySQL Configurations can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_mysql_configuration.interactive_timeout /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.DBforMySQL/servers/server1/configurations/interactive_timeout
```
