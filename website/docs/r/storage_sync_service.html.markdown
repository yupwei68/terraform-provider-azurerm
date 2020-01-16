---
subcategory: ""
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_storage_sync_service"
sidebar_current: "docs-azurerm-resource-storage-sync-service"
description: |-
  Manage Azure StorageSyncService instance.
---

# azurerm_storage_sync_service

Manage Azure StorageSyncService instance.


## Argument Reference

The following arguments are supported:

* `resource_group` - (Required) The name of the resource group. The name is case insensitive. Changing this forces a new resource to be created.

* `location` - (Required) Required. Gets or sets the location of the resource. This will be one of the supported and registered Azure Geo Regions (e.g. West US, East US, Southeast Asia, etc.). The geo region of a resource cannot be changed once it is created, but if an identical geo region is specified on update, the request will succeed. Changing this forces a new resource to be created.

* `location_name` - (Required) The desired region for the name check. Changing this forces a new resource to be created.

* `name` - (Required) The name to check for availability Changing this forces a new resource to be created.

* `storage_sync_service_name` - (Required) Name of Storage Sync Service resource. Changing this forces a new resource to be created.

* `type` - (Required) The resource type. Must be set to Microsoft.StorageSync/storageSyncServices Changing this forces a new resource to be created.

* `tags` - (Optional) The user-specified tags associated with the storage sync service. Changing this forces a new resource to be created.

## Attributes Reference

The following attributes are exported:

* `storage_sync_service_status` - Storage Sync service status.

* `storage_sync_service_uid` - Storage Sync service Uid

* `id` - Fully qualified resource Id for the resource. Ex - /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}/{resourceName}
