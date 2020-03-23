---
subcategory: "Recovery Services"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_recovery_services_vault"
description: |-
  Manages a Recovery Services Vault.
---

# azurerm_recovery_services_vault

Manages an Recovery Services Vault.

## Example Usage

```hcl
resource "azurerm_resource_group" "rg" {
  name     = "tfex-recovery_vault"
  location = "West US"
}

resource "azurerm_recovery_services_vault" "vault" {
  name                = "example_recovery_vault"
  location            = "${azurerm_resource_group.rg.location}"
  resource_group_name = "${azurerm_resource_group.rg.name}"
  sku                 = "Standard"

  soft_delete_enabled = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the Recovery Services Vault. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the Recovery Services Vault. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `tags` - (Optional) A mapping of tags to assign to the resource.

* `sku` - (Required) Sets the vault's SKU. Possible values include: `Standard`, `RS0`.

* `soft_delete_enabled` - (Optional) Is soft delete enable for this Vault? Defaults to `true`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the Recovery Services Vault.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the Recovery Services Vault.
* `update` - (Defaults to 30 minutes) Used when updating the Recovery Services Vault.
* `read` - (Defaults to 5 minutes) Used when retrieving the Recovery Services Vault.
* `delete` - (Defaults to 30 minutes) Used when deleting the Recovery Services Vault.

## Import

Recovery Services Vaults can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_recovery_services_vault.vault1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.RecoveryServices/vaults/vault1
```
