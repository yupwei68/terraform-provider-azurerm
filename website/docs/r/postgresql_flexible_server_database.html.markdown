---
subcategory: "Database"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_postgresql_flexible_server_database"
description: |-
  Manages a PostgreSQL Flexible Server Database.
---

# azurerm_postgresql_flexible_server_database

Manages a PostgreSQL Flexible Server Database.

## Example Usage

```hcl
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_postgresql_flexible_server" "example" {
  name                   = "example-psqlflexibleserver"
  resource_group_name    = azurerm_resource_group.example.name
  location               = azurerm_resource_group.example.location
  version                = "12"
  administrator_login    = "psqladminun"
  administrator_password = "H@Sh1CoR3!"

  storage_mb = 32768

  sku_name = "GP_Standard_D4s_v3"
}

resource "azurerm_postgresql_flexible_server_database" "example" {
  name             = "example-fw"
  server_id        = azurerm_postgresql_flexible_server.example.id
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name which should be used for this PostgreSQL Flexible Server Database. Changing this forces a new PostgreSQL Flexible Server Database to be created.

* `server_id` - (Required) The ID of the PostgreSQL Flexible Server from which to create this PostgreSQL Flexible Server Database. Changing this forces a new PostgreSQL Flexible Server Database to be created.

---

* `charset` - (Optional) The charset of the PostgreSQL Flexible Server Database. Changing this forces a new PostgreSQL Flexible Server Database to be created.

* `collation` - (Optional) The collation of the PostgreSQL Flexible Server Database. Changing this forces a new PostgreSQL Flexible Server Database to be created.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the PostgreSQL Flexible Server Database.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the PostgreSQL Flexible Server Database.
* `read` - (Defaults to 5 minutes) Used when retrieving the PostgreSQL Flexible Server Database.
* `delete` - (Defaults to 30 minutes) Used when deleting the PostgreSQL Flexible Server Database.

## Import

PostgreSQL Flexible Server Databases can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_postgresql_flexible_server_database.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.DBforPostgreSQL/flexibleServers/flexibleServer1/databases/database1
```