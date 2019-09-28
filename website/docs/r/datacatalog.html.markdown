---
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_datacatalog"
sidebar_current: "docs-azurerm-resource-datacatalog"
description: |-
  Manage a Data Catalog.
---

# azurerm_datacatalog

Manage a Data Catalog.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example"
  location = "northeurope"
}

resource "azurerm_datacatalog" "test" {
  name                = "example"
  resource_group_name = "${azurerm_resource_group.example.name}"
  location            = "${azurerm_resource_group.example.location}"
  sku                 = "Free"

  admin {
    upn = "${azuread_user.example.user_principal_name}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Data Catalog. Changing this forces a new resource to be created. Must be globally unique. 

* `resource_group_name` - (Required) The name of the resource group in which to create the Data Catalog. Changing this forces a new resource

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `sku` - (Required) The Sku which should be used for the Data Catalog. Possible values are `Free`, or `Standard`. Changing this forces a new resource to be created.

* `admin` - (Optional) An admin block as defined below.

* `user` - (Optional) A user block as defined below.

* `enable_automatic_unit_adjustment` - (Optional) Specifies whether automatic unit adjustment is enabled.

* `units` - (Optional) The number of Data Catalog units.

---

A `admin` block supports the following:

* `upn` - (Optional) The UPN of the admin.

* `object_id` - (Optional) The Object Id for the admin.

---

A `user` block supports the following:

* `upn` - (Optional) The UPN of the user.

* `object_id` - (Optional) The Object Id for the user.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Data Catalog.

## Import

Data Catalog can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_datacatalog.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example/providers/Microsoft.DataCatalog/catalogs/example
```
