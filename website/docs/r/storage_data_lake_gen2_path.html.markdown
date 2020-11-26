---
subcategory: "Storage"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_storage_data_lake_gen2_path"
description: |-
  Manages a Data Lake Gen2 Path in a File System within an Azure Storage Account.
---

# azurerm_storage_data_lake_gen2_path

Manages a Data Lake Gen2 Path in a File System within an Azure Storage Account.

~> **NOTE:** This Resource requires using Azure Active Directory to connect to Azure Storage, which in turn requires the `Storage` specific roles - which are not granted by default.

## Example Usage

```terraform
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_storage_account" "example" {
  name                     = "examplestorageacc"
  resource_group_name      = azurerm_resource_group.example.name
  location                 = azurerm_resource_group.example.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"
  is_hns_enabled           = "true"
}

resource "azurerm_storage_data_lake_gen2_filesystem" "example" {
  name               = "example"
  storage_account_id = azurerm_storage_account.example.id
}
resource "azurerm_storage_data_lake_gen2_path" "example" {
  path               = "example"
  filesystem_name    = azurerm_storage_data_lake_gen2_filesystem.example.name
  storage_account_id = azurerm_storage_account.example.id
  resource           = "directory"
}

```

## Argument Reference

The following arguments are supported:

* `path` - (Required) The path which should be created within the Data Lake Gen2 File System in the Storage Account. Changing this forces a new resource to be created.

* `filesystem_name` - (Required) The name of the Data Lake Gen2 File System which should be created within the Storage Account. Must be unique within the storage account the queue is located. Changing this forces a new resource to be created.

* `storage_account_id` - (Required) Specifies the ID of the Storage Account in which the Data Lake Gen2 File System should exist. Changing this forces a new resource to be created.

* `resource` - (Required) Specifies the type for path to create. Currently only `directory` is supported.

* `owner` - (Optional) Specifies the Object ID of the Azure Active Directory User to make the owning user.

* `group` - (Optional) Specifies the Object ID of the Azure Active Directory Group to make the owning group.

* `ace` - (Required) One or more `ace` blocks as defined below to specify the entries for the ACL for the path.


---

An `ace` block supports the following:

* `scope` - (Optional) Specifies whether the ACE represents an `access` entry or a `default` entry. Default value is `access`.

* `type` - (Required) Specifies the type of entry. Can be `user`, `group`, `mask` or `other`.

* `id` - (Optional) Specifies the Object ID of the Azure Active Directory User or Group that the entry relates to. Only valid for `user` or `group` entries.

* `permissions` - (Required) Specifies the permissions for the entry in `rwx` form. For example, `rwx` gives full permissions but `r--` only gives read permissions.

More details on ACLs can be found here: https://docs.microsoft.com/en-us/azure/storage/blobs/data-lake-storage-access-control#access-control-lists-on-files-and-directories

~> **NOTE:** The Storage Account requires `account_kind` to be either `StorageV2` or `BlobStorage`. In addition, `is_hns_enabled` has to be set to `true`.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the Data Lake Gen2 File System.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the path.
* `update` - (Defaults to 30 minutes) Used when updating the path.
* `read` - (Defaults to 5 minutes) Used when retrieving the path.
* `delete` - (Defaults to 30 minutes) Used when deleting the path.

## Import

Data Lake Gen2 Paths can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_storage_data_lake_gen2_path.example https://account1.dfs.core.windows.net/fileSystem1/path
```
