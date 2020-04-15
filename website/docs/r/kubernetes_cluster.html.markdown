---
subcategory: "Container"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_kubernetes_cluster"
description: |-
  Manages a managed Kubernetes Cluster (also known as AKS / Azure Kubernetes Service)
---

# azurerm_kubernetes_cluster

Manages a Managed Kubernetes Cluster (also known as AKS / Azure Kubernetes Service)

~> **Note:** All arguments including the client secret will be stored in the raw state as plain-text. [Read more about sensitive data in state](/docs/state/sensitive-data.html).

## Example Usage

This example provisions a basic Managed Kubernetes Cluster. Other examples of the `azurerm_kubernetes_cluster` resource can be found in [the `./examples/kubernetes` directory within the Github Repository](https://github.com/terraform-providers/terraform-provider-azurerm/tree/master/examples/kubernetes)

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_kubernetes_cluster" "example" {
  name                = "example-aks1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  dns_prefix          = "exampleaks1"

  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_D2_v2"
  }

  identity {
    type = "SystemAssigned"
  }

  tags = {
    Environment = "Production"
  }
}

output "client_certificate" {
  value = azurerm_kubernetes_cluster.example.kube_config.0.client_certificate
}

output "kube_config" {
  value = azurerm_kubernetes_cluster.example.kube_config_raw
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Managed Kubernetes Cluster to create. Changing this forces a new resource to be created.

* `location` - (Required) The location where the Managed Kubernetes Cluster should be created. Changing this forces a new resource to be created.

* `resource_group_name` - (Required) Specifies the Resource Group where the Managed Kubernetes Cluster should exist. Changing this forces a new resource to be created.

* `default_node_pool` - (Optional) A `default_node_pool` block as defined below.

-> **NOTE:** The `default_node_pool` block will become required in 2.0

* `dns_prefix` - (Required) DNS prefix specified when creating the managed cluster. Changing this forces a new resource to be created.

-> **NOTE:** The `dns_prefix` must contain between 3 and 45 characters, and can contain only letters, numbers, and hyphens. It must start with a letter and must end with a letter or a number.

In addition, one of either `identity` or `service_principal` must be specified.

---

* `addon_profile` - (Optional) A `addon_profile` block as defined below.

* `api_server_authorized_ip_ranges` - (Optional) The IP ranges to whitelist for incoming traffic to the masters.

* `enable_pod_security_policy` - (Optional) Whether Pod Security Policies are enabled. Note that this also requires role based access control to be enabled.

-> **NOTE:** Support for `enable_pod_security_policy` is currently in Preview on an opt-in basis. To use it, enable feature `PodSecurityPolicyPreview` for `namespace Microsoft.ContainerService`. For an example of how to enable a Preview feature, please visit [Register scale set feature provider](https://docs.microsoft.com/en-us/azure/aks/cluster-autoscaler#register-scale-set-feature-provider).

* `identity` - (Optional) A `identity` block as defined below. Changing this forces a new resource to be created.

-> **NOTE:** One of either `identity` or `service_principal` must be specified.

* `kubernetes_version` - (Optional) Version of Kubernetes specified when creating the AKS managed cluster. If not specified, the latest recommended version will be used at provisioning time (but won't auto-upgrade).

-> **NOTE:** Upgrading your cluster may take up to 10 minutes per node.

* `linux_profile` - (Optional) A `linux_profile` block as defined below.

* `network_profile` - (Optional) A `network_profile` block as defined below.

-> **NOTE:** If `network_profile` is not defined, `kubenet` profile will be used by default.

* `node_resource_group` - (Optional) The name of the Resource Group where the Kubernetes Nodes should exist. Changing this forces a new resource to be created.

-> **NOTE:** Azure requires that a new, non-existent Resource Group is used, as otherwise the provisioning of the Kubernetes Service will fail.

* `private_link_enabled` Should this Kubernetes Cluster have Private Link Enabled? This provides a Private IP Address for the Kubernetes API on the Virtual Network where the Kubernetes Cluster is located. Defaults to `false`. Changing this forces a new resource to be created.

-> **NOTE:**  At this time Private Link is in Public Preview. For an example of how to enable a Preview feature, please visit [Private Azure Kubernetes Service cluster](https://docs.microsoft.com/en-gb/azure/aks/private-clusters)

* `role_based_access_control` - (Optional) A `role_based_access_control` block. Changing this forces a new resource to be created.

* `service_principal` - (Optional) A `service_principal` block as documented below.

-> **NOTE:** One of either `identity` or `service_principal` must be specified.

* `tags` - (Optional) A mapping of tags to assign to the resource.

* `windows_profile` - (Optional) A `windows_profile` block as defined below.


---

A `aci_connector_linux` block supports the following:

* `enabled` - (Required) Is the virtual node addon enabled?

* `subnet_name` - (Optional) The subnet name for the virtual nodes to run. This is required when `aci_connector_linux` `enabled` argument is set to `true`.

-> **NOTE:** AKS will add a delegation to the subnet named here. To prevent further runs from failing you should make sure that the subnet you create for virtual nodes has a delegation, like so.

```
resource "azurerm_subnet" "virtual" {

  #...

  delegation {
    name = "aciDelegation"
    service_delegation {
      name    = "Microsoft.ContainerInstance/containerGroups"
      actions = ["Microsoft.Network/virtualNetworks/subnets/action"]
    }
  }
}
```

---

A `addon_profile` block supports the following:

* `aci_connector_linux` - (Optional) A `aci_connector_linux` block. For more details, please visit [Create and configure an AKS cluster to use virtual nodes](https://docs.microsoft.com/en-us/azure/aks/virtual-nodes-portal).

-> **NOTE:** At this time ACI Connector's are not supported in Azure China.

* `azure_policy` - (Optional) A `azure_policy` block as defined below. For more details please visit [Understand Azure Policy for Azure Kubernetes Service](https://docs.microsoft.com/en-ie/azure/governance/policy/concepts/rego-for-aks)

-> **NOTE**: Azure Policy for Azure Kubernetes Service is currently in preview and not available to subscriptions that have not [opted-in](https://docs.microsoft.com/en-us/azure/governance/policy/concepts/rego-for-aks?toc=/azure/aks/toc.json) to join `Azure Policy` preview.

* `http_application_routing` - (Optional) A `http_application_routing` block as defined below.

-> **NOTE:** At this time HTTP Application Routing is not supported in Azure China or Azure US Government.

* `kube_dashboard` - (Optional) A `kube_dashboard` block as defined below.

* `oms_agent` - (Optional) A `oms_agent` block as defined below. For more details, please visit [How to onboard Azure Monitor for containers](https://docs.microsoft.com/en-us/azure/monitoring/monitoring-container-insights-onboard).

---

A `azure_active_directory` block supports the following:

* `client_app_id` - (Required) The Client ID of an Azure Active Directory Application.

* `server_app_id` - (Required) The Server ID of an Azure Active Directory Application.

* `server_app_secret` - (Required) The Server Secret of an Azure Active Directory Application.

* `tenant_id` - (Optional) The Tenant ID used for Azure Active Directory Application. If this isn't specified the Tenant ID of the current Subscription is used.


---

A `azure_policy` block supports the following:

* `enabled` - (Required) Is the Azure Policy for Kubernetes Add On enabled?

---

A `default_node_pool` block supports the following:

* `name` - (Required) The name which should be used for the default Kubernetes Node Pool. Changing this forces a new resource to be created.

* `vm_size` - (Required) The size of the Virtual Machine, such as `Standard_DS2_v2`.

* `availability_zones` - (Optional) A list of Availability Zones across which the Node Pool should be spread.

-> **NOTE:** This requires that the `type` is set to `VirtualMachineScaleSets` and that `load_balancer_sku` is set to `Standard`.

* `enable_auto_scaling` - (Optional) Should [the Kubernetes Auto Scaler](https://docs.microsoft.com/en-us/azure/aks/cluster-autoscaler) be enabled for this Node Pool? Defaults to `false`.

-> **NOTE:** This requires that the `type` is set to `VirtualMachineScaleSets`.

-> **NOTE:** If you're using AutoScaling, you may wish to use [Terraform's `ignore_changes` functionality](https://www.terraform.io/docs/configuration/resources.html#ignore_changes) to ignore changes to the `node_count` field.

* `enable_node_public_ip` - (Optional) Should nodes in this Node Pool have a Public IP Address? Defaults to `false`.

* `max_pods` - (Optional) The maximum number of pods that can run on each agent. Changing this forces a new resource to be created.

* `node_labels` - (Optional) A map of Kubernetes labels which should be applied to nodes in the Default Node Pool.

* `node_taints` - (Optional) A list of Kubernetes taints which should be applied to nodes in the agent pool (e.g `key=value:NoSchedule`).

* `os_disk_size_gb` - (Optional) The size of the OS Disk which should be used for each agent in the Node Pool. Changing this forces a new resource to be created.

* `type` - (Optional) The type of Node Pool which should be created. Possible values are `AvailabilitySet` and `VirtualMachineScaleSets`. Defaults to `VirtualMachineScaleSets`.

* `tags` - (Optional) A mapping of tags to assign to the Node Pool.

~> At this time there's a bug in the AKS API where Tags for a Node Pool are not stored in the correct case - you [may wish to use Terraform's `ignore_changes` functionality to ignore changes to the casing](https://www.terraform.io/docs/configuration/resources.html#ignore_changes) until this is fixed in the AKS API. 

* `vnet_subnet_id` - (Optional) The ID of a Subnet where the Kubernetes Node Pool should exist. Changing this forces a new resource to be created.

~> **NOTE:** A Route Table must be configured on this Subnet.

If `enable_auto_scaling` is set to `true`, then the following fields can also be configured:

* `max_count` - (Required) The maximum number of nodes which should exist in this Node Pool. If specified this must be between `1` and `100`.

* `min_count` - (Required) The minimum number of nodes which should exist in this Node Pool. If specified this must be between `1` and `100`.

* `node_count` - (Optional) The initial number of nodes which should exist in this Node Pool. If specified this must be between `1` and `100` and between `min_count` and `max_count`.

-> **NOTE:** If specified you may wish to use [Terraform's `ignore_changes` functionality](https://www.terraform.io/docs/configuration/resources.html#ignore_changes) to ignore changes to this field.

If `enable_auto_scaling` is set to `false`, then the following fields can also be configured:

* `node_count` - (Required) The number of nodes which should exist in this Node Pool. If specified this must be between `1` and `100`.

-> **NOTE:** If `enable_auto_scaling` is set to `false` both `min_count` and `max_count` fields need to be set to `null` or omitted from the configuration.

---

A `http_application_routing` block supports the following:

* `enabled` (Required) Is HTTP Application Routing Enabled? Changing this forces a new resource to be created.

---

A `identity` block supports the following:

* `type` - The type of identity used for the managed cluster. At this time the only supported value is `SystemAssigned`.

---

A `kube_dashboard` block supports the following:

* `enabled` - (Required) Is the Kubernetes Dashboard enabled?

---

A `linux_profile` block supports the following:

* `admin_username` - (Required) The Admin Username for the Cluster. Changing this forces a new resource to be created.

* `ssh_key` - (Required) An `ssh_key` block. Only one is currently allowed. Changing this forces a new resource to be created.

---

A `network_profile` block supports the following:

* `network_plugin` - (Required) Network plugin to use for networking. Currently supported values are `azure` and `kubenet`. Changing this forces a new resource to be created.

-> **NOTE:** When `network_plugin` is set to `azure` - the `vnet_subnet_id` field in the `default_node_pool` block must be set and `pod_cidr` must not be set.

* `network_policy` - (Optional) Sets up network policy to be used with Azure CNI. [Network policy allows us to control the traffic flow between pods](https://docs.microsoft.com/en-us/azure/aks/use-network-policies). Currently supported values are `calico` and `azure`. Changing this forces a new resource to be created.

~> **NOTE:** When `network_plugin` is set to `kubenet` the `network_policy` field can only be set to `calico`, otherwise it has to be set to `azure`.

* `dns_service_ip` - (Optional) IP address within the Kubernetes service address range that will be used by cluster service discovery (kube-dns). This is required when `network_plugin` is set to `azure`. Changing this forces a new resource to be created.

* `docker_bridge_cidr` - (Optional) IP address (in CIDR notation) used as the Docker bridge IP address on nodes. This is required when `network_plugin` is set to `azure`. Changing this forces a new resource to be created.

* `outbound_type` - (Optional) The outbound (egress) routing method which should be used for this Kubernetes Cluster. Possible values are `loadBalancer` and `userDefinedRouting`. Defaults to `loadBalancer`.

* `pod_cidr` - (Optional) The CIDR to use for pod IP addresses. This field can only be set when `network_plugin` is set to `kubenet`. Changing this forces a new resource to be created.

* `service_cidr` - (Optional) The Network Range used by the Kubernetes service. This is required when `network_plugin` is set to `azure`. Changing this forces a new resource to be created.

~> **NOTE:** This range should not be used by any network element on or connected to this VNet. Service address CIDR must be smaller than /12.

Examples of how to use [AKS with Advanced Networking](https://docs.microsoft.com/en-us/azure/aks/networking-overview#advanced-networking) can be [found in the `./examples/kubernetes/` directory in the Github repository](https://github.com/terraform-providers/terraform-provider-azurerm/tree/master/examples/kubernetes).

* `load_balancer_sku` - (Optional) Specifies the SKU of the Load Balancer used for this Kubernetes Cluster. Possible values are `Basic` and `Standard`. Defaults to `Standard`.

* `load_balancer_profile` - (Optional) A `load_balancer_profile` block. This can only be specified when `load_balancer_sku` is set to `Standard`.

---

A `load_balancer_profile` block supports the following:

~> **NOTE:** These options are mutually exclusive. Note that when specifying `outbound_ip_address_ids` ([azurerm_public_ip](/docs/providers/azurerm/r/public_ip.html)) the SKU must be `Standard`.

* `managed_outbound_ip_count` - (Optional) Count of desired managed outbound IPs for the cluster load balancer. Must be in the range of [1, 100].

* `outbound_ip_prefix_ids` - (Optional) The ID of the outbound Public IP Address Prefixes which should be used for the cluster load balancer.

* `outbound_ip_address_ids` - (Optional) The ID of the Public IP Addresses which should be used for outbound communication for the cluster load balancer.

---

A `oms_agent` block supports the following:

* `enabled` - (Required) Is the OMS Agent Enabled?

* `log_analytics_workspace_id` - (Optional) The ID of the Log Analytics Workspace which the OMS Agent should send data to. Must be present if `enabled` is `true`.

---

A `role_based_access_control` block supports the following:

* `azure_active_directory` - (Optional) An `azure_active_directory` block.

* `enabled` - (Required) Is Role Based Access Control Enabled? Changing this forces a new resource to be created.

---

A `service_principal` block supports the following:

* `client_id` - (Required) The Client ID for the Service Principal.

* `client_secret` - (Required) The Client Secret for the Service Principal.

---

A `ssh_key` block supports the following:

* `key_data` - (Required) The Public SSH Key used to access the cluster. Changing this forces a new resource to be created.

---

A `windows_profile` block supports the following:

* `admin_username` - (Required) The Admin Username for Windows VMs.

* `admin_password` - (Required) The Admin Password for Windows VMs.


## Attributes Reference

The following attributes are exported:

* `id` - The Kubernetes Managed Cluster ID.

* `fqdn` - The FQDN of the Azure Kubernetes Managed Cluster.

* `private_fqdn` - The FQDN for the Kubernetes Cluster when private link has been enabled, which is only resolvable inside the Virtual Network used by the Kubernetes Cluster.

* `kube_admin_config` - A `kube_admin_config` block as defined below. This is only available when Role Based Access Control with Azure Active Directory is enabled.

* `kube_admin_config_raw` - Raw Kubernetes config for the admin account to be used by [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) and other compatible tools. This is only available when Role Based Access Control with Azure Active Directory is enabled.

* `kube_config` - A `kube_config` block as defined below.

* `kube_config_raw` - Raw Kubernetes config to be used by [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/) and other compatible tools

* `http_application_routing` - A `http_application_routing` block as defined below.

* `node_resource_group` - The auto-generated Resource Group which contains the resources for this Managed Kubernetes Cluster.

---

A `http_application_routing` block exports the following:

* `http_application_routing_zone_name` - The Zone Name of the HTTP Application Routing.

---

A `load_balancer_profile` block exports the following:

* `effective_outbound_ips` - The outcome (resource IDs) of the specified arguments.

---

The `identity` block exports the following:

* `principal_id` - The principal id of the system assigned identity which is used by master components.

* `tenant_id` - The tenant id of the system assigned identity which is used by master components.

---

The `kube_admin_config` and `kube_config` blocks export the following:

* `client_key` - Base64 encoded private key used by clients to authenticate to the Kubernetes cluster.

* `client_certificate` - Base64 encoded public certificate used by clients to authenticate to the Kubernetes cluster.

* `cluster_ca_certificate` - Base64 encoded public CA certificate used as the root of trust for the Kubernetes cluster.

* `host` - The Kubernetes cluster server host.

* `username` - A username used to authenticate to the Kubernetes cluster.

* `password` - A password or token used to authenticate to the Kubernetes cluster.

-> **NOTE:** It's possible to use these credentials with [the Kubernetes Provider](/docs/providers/kubernetes/index.html) like so:

```
provider "kubernetes" {
  load_config_file       = "false"
  host                   = "${azurerm_kubernetes_cluster.main.kube_config.0.host}"
  username               = "${azurerm_kubernetes_cluster.main.kube_config.0.username}"
  password               = "${azurerm_kubernetes_cluster.main.kube_config.0.password}"
  client_certificate     = "${base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_certificate)}"
  client_key             = "${base64decode(azurerm_kubernetes_cluster.main.kube_config.0.client_key)}"
  cluster_ca_certificate = "${base64decode(azurerm_kubernetes_cluster.main.kube_config.0.cluster_ca_certificate)}"
}
```

## Timeouts



The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 90 minutes) Used when creating the Kubernetes Cluster.
* `update` - (Defaults to 90 minutes) Used when updating the Kubernetes Cluster.
* `read` - (Defaults to 5 minutes) Used when retrieving the Kubernetes Cluster.
* `delete` - (Defaults to 90 minutes) Used when deleting the Kubernetes Cluster.

## Import

Managed Kubernetes Clusters can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_kubernetes_cluster.cluster1 /subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group1/providers/Microsoft.ContainerService/managedClusters/cluster1
```
