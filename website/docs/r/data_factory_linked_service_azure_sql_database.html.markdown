---
subcategory: "Data Factory"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_data_factory_linked_service_azure_sql_database"
description: |-
  Manages a Linked Service (connection) between Azure SQL Database and Azure Data Factory.
---

# azurerm_data_factory_linked_service_azure_sql_database

Manages a Linked Service (connection) between Azure SQL Database and Azure Data Factory.

~> **Note:** All arguments including the connection_string will be stored in the raw state as plain-text. [Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "northeurope"
}

resource "azurerm_data_factory" "example" {
  name                = "example"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
}

resource "azurerm_data_factory_linked_service_azure_sql_database" "example" {
  name                = "example"
  resource_group_name = azurerm_resource_group.example.name
  data_factory_name   = azurerm_data_factory.example.name
  connection_string   = "data source=serverhostname;initial catalog=master;user id=testUser;Password=test;integrated security=False;encrypt=True;connection timeout=30"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Data Factory Linked Service Azure SQL Database. Changing this forces a new resource to be created. Must be globally unique. See the [Microsoft documentation](https://docs.microsoft.com/en-us/azure/data-factory/naming-rules) for all restrictions.

* `resource_group_name` - (Required) The name of the resource group in which to create the Data Factory Linked Service Azure SQL Database. Changing this forces a new resource to be created.

* `data_factory_name` - (Required) The Data Factory name in which to associate the Linked Service with. Changing this forces a new resource to be created.

* `connection_string` - (Required) The connection string in which to authenticate with Azure SQL Database.

* `description` - (Optional) The description for the Data Factory Linked Service Azure SQL Database.

* `integration_runtime_name` - (Optional) The integration runtime reference to associate with the Data Factory Linked Service Azure SQL Database.

* `annotations` - (Optional) List of tags that can be used for describing the Data Factory Linked Service Azure SQL Database.

* `parameters` - (Optional) A map of parameters to associate with the Data Factory Linked Service Azure SQL Database.

* `additional_properties` - (Optional) A map of additional properties to associate with the Data Factory Linked Service Azure SQL Database.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Data Factory Azure SQL Database Linked Service.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Data Factory Azure SQL Database Linked Service.
* `update` - (Defaults to 30 minutes) Used when updating the Data Factory Azure SQL Database Linked Service.
* `read` - (Defaults to 5 minutes) Used when retrieving the Data Factory Azure SQL Database Linked Service.
* `delete` - (Defaults to 30 minutes) Used when deleting the Data Factory Azure SQL Database Linked Service.

## Import

Data Factory Azure SQL Database Linked Service's can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_data_factory_linked_service_azure_sql_database.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example/providers/Microsoft.DataFactory/factories/example/linkedservices/example
```
