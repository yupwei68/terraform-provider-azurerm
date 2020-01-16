---
subcategory: "Database"
layout: "azurerm"
page_title: "Azure Resource Manager: Data Source: azurerm_mssql_server"
description: |-
  Gets information about an existing Microsoft SQL Server.
---

# Data Source: azurerm_mssql_server

Use this data source to access information about an existing Microsoft SQL Server.

## Example Usage

```hcl
data "azurerm_mssql_server" "example" {
  name                = "example-mssql-server"
  resource_group_name = "example-resource-group"
}

output "id" {
  value = data.azurerm_mssql_server.example.id
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name of this Microsoft SQL Server.

* `resource_group_name` - (Required) The name of the Resource Group where the Microsoft SQL Server exists.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the Microsoft SQL Server.

* `location` - The Azure Region where the Microsoft SQL Server exists.

* `fqdn` - The fully qualified domain name of the SQL Server.

* `administrator_login` - The administrator login name for the new server.

* `identity` - A `identity` block as defined below.

* `recoverable_databases` - A list of recoverable database ids of the server.

* `restorable_dropped_databases` - A list of restorable database ids of the server.

* `version` - The version for the new server.

* `tags` - A mapping of tags assigned to the Microsoft SQL Server.

---

A `identity` block exports the following:

* `principal_id` - The ID of the Principal (Client) in Azure Active Directory.

* `tenant_id` - The ID of the Azure Active Directory Tenant.

* `type` - The identity type of the SQL Server.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Microsoft SQL Server.
