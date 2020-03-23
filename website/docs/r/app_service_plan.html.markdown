---
subcategory: "App Service (Web Apps)"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_app_service_plan"
description: |-
  Manages an App Service Plan component.
---

# azurerm_app_service_plan

Manages an App Service Plan component.

## Example Usage (Dedicated)

```hcl
resource "azurerm_resource_group" "example" {
  name     = "api-rg-pro"
  location = "West Europe"
}

resource "azurerm_app_service_plan" "example" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  sku {
    tier = "Standard"
    size = "S1"
  }
}
```

## Example Usage (Shared / Consumption Plan)

```hcl
resource "azurerm_resource_group" "example" {
  name     = "api-rg-pro"
  location = "West Europe"
}

resource "azurerm_app_service_plan" "example" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "FunctionApp"

  sku {
    tier = "Dynamic"
    size = "Y1"
  }
}
```

## Example Usage (Linux)

```hcl
resource "azurerm_resource_group" "example" {
  name     = "api-rg-pro"
  location = "West Europe"
}

resource "azurerm_app_service_plan" "example" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "Linux"
  reserved            = true

  sku {
    tier = "Standard"
    size = "S1"
  }
}
```

## Example Usage (Windows Container)

```hcl
resource "azurerm_resource_group" "example" {
  name     = "api-rg-pro"
  location = "West Europe"
}

resource "azurerm_app_service_plan" "example" {
  name                = "api-appserviceplan-pro"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  kind                = "xenon"
  is_xenon            = true

  sku {
    tier = "PremiumContainer"
    size = "PC2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the App Service Plan component. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which to create the App Service Plan component.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `kind` - (Optional) The kind of the App Service Plan to create. Possible values are `Windows` (also available as `App`), `Linux`, `elastic` (for Premium Consumption) and `FunctionApp` (for a Consumption Plan). Defaults to `Windows`. Changing this forces a new resource to be created.

~> **NOTE:** When creating a `Linux` App Service Plan, the `reserved` field must be set to `true`, and when creating a `Windows`/`app` App Service Plan the `reserved` field must be set to `false`.

* `maximum_elastic_worker_count` - The maximum number of total workers allowed for this ElasticScaleEnabled App Service Plan.

* `sku` - (Required) A `sku` block as documented below.

* `app_service_environment_id` - (Optional) The ID of the App Service Environment where the App Service Plan should be located. Changing forces a new resource to be created.

~> **NOTE:** Attaching to an App Service Environment requires the App Service Plan use a `Premium` SKU (when using an ASEv1) and the `Isolated` SKU (for an ASEv2).

* `reserved` - (Optional) Is this App Service Plan `Reserved`. Defaults to `false`.

* `per_site_scaling` - (Optional) Can Apps assigned to this App Service Plan be scaled independently? If set to `false` apps assigned to this plan will scale to all instances of the plan.  Defaults to `false`.

* `tags` - (Optional) A mapping of tags to assign to the resource.

`sku` supports the following:

* `tier` - (Required) Specifies the plan's pricing tier.

* `size` - (Required) Specifies the plan's instance size.

* `capacity` - (Optional) Specifies the number of workers associated with this App Service Plan.


## Attributes Reference

The following attributes are exported:

* `id` - The ID of the App Service Plan component.
* `maximum_number_of_workers` - The maximum number of workers supported with the App Service Plan's sku.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 60 minutes) Used when creating the App Service Plan.
* `update` - (Defaults to 60 minutes) Used when updating the App Service Plan.
* `read` - (Defaults to 5 minutes) Used when retrieving the App Service Plan.
* `delete` - (Defaults to 60 minutes) Used when deleting the App Service Plan.

## Import

App Service Plan instances can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_app_service_plan.instance1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mygroup1/providers/Microsoft.Web/serverfarms/instance1
```
