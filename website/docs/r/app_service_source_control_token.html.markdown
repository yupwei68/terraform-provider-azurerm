---
subcategory: "App Service (Web Apps)"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_app_service_source_control_token"
description: |-
  Manages an App Service source control token.

---

# azurerm_app_service_source_control_token

Manages an App Service source control token.

~> **NOTE:** Source Control Tokens are configured at the subscription level, not on each App Service - as such this can only be configured Subscription-wide

## Example Usage

```hcl
resource "azurerm_app_service_source_control_token" "example" {
  type  = "GitHub"
  token = "7e57735e77e577e57"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The source control type. Possible values are `BitBucket`, `Dropbox`, `GitHub` and `OneDrive`.

* `token` - (Required) The OAuth access token.

* `token_secret` - (Optional) The OAuth access token secret.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the App Service Source Control Token.
* `update` - (Defaults to 30 minutes) Used when updating the App Service Source Control Token.
* `read` - (Defaults to 5 minutes) Used when retrieving the App Service Source Control Token.
* `delete` - (Defaults to 30 minutes) Used when deleting the App Service Source Control Token.

## Import

App Service Source Control Token's can be imported using the `type`, e.g.

```shell
terraform import azurerm_app_service_source_control_token.example GitHub
```
