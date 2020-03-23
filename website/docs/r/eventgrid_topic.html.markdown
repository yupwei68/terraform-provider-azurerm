---
subcategory: "Messaging"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_eventgrid_topic"
description: |-
  Manages an EventGrid Topic

---

# azurerm_eventgrid_topic

Manages an EventGrid Topic

~> **Note:** at this time EventGrid Topic's are only available in a limited number of regions.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "resourceGroup1"
  location = "West US 2"
}

resource "azurerm_eventgrid_topic" "example" {
  name                = "my-eventgrid-topic"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  tags = {
    environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the EventGrid Topic resource. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) The name of the resource group in which the EventGrid Topic exists. Changing this forces a new resource to be created.

* `location` - (Required) Specifies the supported Azure location where the resource exists. Changing this forces a new resource to be created.

* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

The following attributes are exported:

* `id` - The EventGrid Topic ID.

* `endpoint` - The Endpoint associated with the EventGrid Topic.

* `primary_access_key` - The Primary Shared Access Key associated with the EventGrid Topic.

* `secondary_access_key` - The Secondary Shared Access Key associated with the EventGrid Topic.

## Timeouts



The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 30 minutes) Used when creating the EventGrid Topic.
* `update` - (Defaults to 30 minutes) Used when updating the EventGrid Topic.
* `read` - (Defaults to 5 minutes) Used when retrieving the EventGrid Topic.
* `delete` - (Defaults to 30 minutes) Used when deleting the EventGrid Topic.

## Import

EventGrid Topic's can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_eventgrid_topic.topic1 /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group1/providers/Microsoft.EventGrid/topics/topic1
```
