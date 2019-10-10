## Examples for the Virtual Machine Scale Set resources

In 1.x versions of the Provider, Terraform has a single resource for Virtual Machine Scale Sets: `azurerm_virtual_machine_scale_set`.

Coming in the next major version of the Azure Provider (2.0 - [more details here](https://github.com/terraform-providers/terraform-provider-azurerm/issues/2807)) are three new resources:

* `azurerm_linux_virtual_machine_scale_set`
* `azurerm_virtual_machine_scale_set_extension`
* `azurerm_windows_virtual_machine_scale_set`

which longer term will replace the existing `azurerm_virtual_machine_scale_set` resource - but **do not support unmanaged disks** - if you're looking to use Unmanaged Disks you'll need to continue using the `azurerm_virtual_machine_scale_set` resource..

This directory contains 4 sub-directories:

* `./virtual_machine_scale_set` - which are examples of how to use the `azurerm_virtual_machine_scale_set` resource.
* `./linux` - which are examples of how to use the `azurerm_linux_virtual_machine_scale_set` resource.
* `./extensions` - which are examples of how to use the `azurerm_virtual_machine_scale_set_extension` resource.
* `./windows` - which are examples of how to use the `azurerm_windows_virtual_machine_scale_set` resource.
