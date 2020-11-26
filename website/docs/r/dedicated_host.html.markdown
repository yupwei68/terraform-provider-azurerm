---
subcategory: "Compute"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_dedicated_host"
description: |-
  Manage a Dedicated Host within a Dedicated Host Group.
---

# azurerm_dedicated_host

Manage a Dedicated Host within a Dedicated Host Group.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resourcs"
  location = "West Europe"
}

resource "azurerm_dedicated_host_group" "example" {
  name                        = "example-host-group"
  resource_group_name         = azurerm_resource_group.example.name
  location                    = azurerm_resource_group.example.location
  platform_fault_domain_count = 2
}

resource "azurerm_dedicated_host" "example" {
  name                    = "example-host"
  location                = azurerm_resource_group.example.location
  dedicated_host_group_id = azurerm_dedicated_host_group.example.id
  sku_name                = "DSv3-Type1"
  platform_fault_domain   = 1
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of this Dedicated Host. Changing this forces a new resource to be created.

* `dedicated_host_group_id` - (Required) Specifies the ID of the Dedicated Host Group where the Dedicated Host should exist. Changing this forces a new resource to be created.

* `location` - (Required) Specify the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `sku_name` - (Required) Specify the sku name of the Dedicated Host. Possible values are `DSv3-Type1`, `DSv3-Type2`, `DSv4-Type1`, `ESv3-Type1`, `ESv3-Type2`,`FSv2-Type2`. Changing this forces a new resource to be created.

* `platform_fault_domain` - (Required) Specify the fault domain of the Dedicated Host Group in which to create the Dedicated Host. Changing this forces a new resource to be created.

---

* `auto_replace_on_failure` - (Optional) Should the Dedicated Host automatically be replaced in case of a Hardware Failure? Defaults to `true`.

* `license_type` - (Optional) Specifies the software license type that will be applied to the VMs deployed on the Dedicated Host. Possible values are `None`, `Windows_Server_Hybrid` and `Windows_Server_Perpetual`. Defaults to `None`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Dedicated Host.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Dedicated Host.
* `update` - (Defaults to 30 minutes) Used when updating the Dedicated Host.
* `read` - (Defaults to 5 minutes) Used when retrieving the Dedicated Host.
* `delete` - (Defaults to 30 minutes) Used when deleting the Dedicated Host.

## Import

Dedicated Hosts can be imported using the `resource id`, e.g.

```shell
$ terraform import azurerm_dedicated_host.example /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Compute/hostGroups/group1/hosts/host1
```
