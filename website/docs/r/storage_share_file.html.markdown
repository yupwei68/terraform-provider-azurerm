---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_storage_share_file"
sidebar_current: "docs-azurerm-resource-storage-share-file"
description: |-
  Manages a File within an Azure Storage File Share.
---

# azurerm_storage_share_file

Manage a File within an Azure Storage File Share.

## Example Usage

```hcl
resource "azurerm_resource_group" "test" {
  name     = "azuretest"
  location = "West Europe"
}

resource "azurerm_storage_account" "test" {
  name                     = "azureteststorage"
  resource_group_name      = "${azurerm_resource_group.test.name}"
  location                 = "${azurerm_resource_group.test.location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_share" "test" {
  name                 = "sharename"
  storage_account_name = "${azurerm_storage_account.test.name}"
  quota                = 50
}

resource "azurerm_storage_share_directory" "test" {
  name                 = "example"
  share_name           = "${azurerm_storage_share.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"
}

resource "azurerm_storage_share_file" "test" {
  name                 = "example.txt"
  share_name           = "${azurerm_storage_share.test.name}"
  share_directory_name = "${azurerm_storage_share_directory.test.name}"
  storage_account_name = "${azurerm_storage_account.test.name}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the File that should be created within this File Share. Changing this forces a new resource to be created.

* `share_name` - (Required) The name of the File Share where this File should be created. Changing this forces a new resource to be created.

* `storage_account_name` - (Required) The name of the Storage Account within which the File Share is located. Changing this forces a new resource to be created.

* `share_directory_name` - (Required) The name of the Directory in which to create the file. Leaving this empty puts the file into the top level of the File Share. 
Changing this forces a new resource.

* `content_length` - (Optional) Specifies the maximum size for the file, up to 1 TiB.

* `content_type` - (Optional) Specifies the MIME content type of the file.

* `content_encoding` - (Optional) Specifies which content encodings have been applied to the file.

* `content_language` - (Optional) Specifies the natural languages used by the File.

* `content_md5` - (Optional) Specifies the file's MD5 hash.

* `content_disposition` - (Optional) Specifies the File’s Content-Disposition header.

* `metadata` - (Optional) A mapping of metadata to assign to this File.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the Directory within the File Share.

## Import

Directories within an Azure Storage File Share can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_storage_share_directory.test https://tomdevsa20.file.core.windows.net/share1/directory1
```
