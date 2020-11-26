package network

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-03-01/network"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/network/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmLoadBalancerOutboundRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmLoadBalancerOutboundRuleCreateUpdate,
		Read:   resourceArmLoadBalancerOutboundRuleRead,
		Update: resourceArmLoadBalancerOutboundRuleCreateUpdate,
		Delete: resourceArmLoadBalancerOutboundRuleDelete,

		Importer: loadBalancerSubResourceImporter(func(input string) (*parse.LoadBalancerId, error) {
			id, err := parse.LoadBalancerOutboundRuleID(input)
			if err != nil {
				return nil, err
			}

			lbId := parse.NewLoadBalancerID(id.ResourceGroup, id.LoadBalancerName)
			return &lbId, nil
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"loadbalancer_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.LoadBalancerID,
			},

			"frontend_ip_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"backend_address_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(network.TransportProtocolAll),
					string(network.TransportProtocolTCP),
					string(network.TransportProtocolUDP),
				}, false),
			},

			"enable_tcp_reset": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"allocated_outbound_ports": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1024,
			},

			"idle_timeout_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  4,
			},
		},
	}
}

func resourceArmLoadBalancerOutboundRuleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	loadBalancerId, err := parse.LoadBalancerID(d.Get("loadbalancer_id").(string))
	if err != nil {
		return err
	}
	loadBalancerIDRaw := loadBalancerId.ID(subscriptionId)
	locks.ByID(loadBalancerIDRaw)
	defer locks.UnlockByID(loadBalancerIDRaw)

	loadBalancer, exists, err := retrieveLoadBalancerById(ctx, client, *loadBalancerId)
	if err != nil {
		return fmt.Errorf("Error Getting Load Balancer By ID: %+v", err)
	}
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer %q not found. Removing from state", name)
		return nil
	}

	newOutboundRule, err := expandAzureRmLoadBalancerOutboundRule(d, loadBalancer)
	if err != nil {
		return fmt.Errorf("expanding Load Balancer Rule: %+v", err)
	}

	outboundRules := make([]network.OutboundRule, 0)

	if loadBalancer.LoadBalancerPropertiesFormat.OutboundRules != nil {
		outboundRules = *loadBalancer.LoadBalancerPropertiesFormat.OutboundRules
	}

	existingOutboundRule, existingOutboundRuleIndex, exists := FindLoadBalancerOutboundRuleByName(loadBalancer, name)
	if exists {
		if name == *existingOutboundRule.Name {
			if d.IsNewResource() {
				return tf.ImportAsExistsError("azurerm_lb_outbound_rule", *existingOutboundRule.ID)
			}

			// this outbound rule is being updated/reapplied remove old copy from the slice
			outboundRules = append(outboundRules[:existingOutboundRuleIndex], outboundRules[existingOutboundRuleIndex+1:]...)
		}
	}

	outboundRules = append(outboundRules, *newOutboundRule)

	loadBalancer.LoadBalancerPropertiesFormat.OutboundRules = &outboundRules

	future, err := client.CreateOrUpdate(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, *loadBalancer)
	if err != nil {
		return fmt.Errorf("Error Creating/Updating LoadBalancer: %+v", err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion for Load Balancer updates: %+v", err)
	}

	read, err := client.Get(ctx, loadBalancerId.ResourceGroup, loadBalancerId.Name, "")
	if err != nil {
		return fmt.Errorf("Error Getting LoadBalancer: %+v", err)
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Load Balancer %s (resource group %s) ID", loadBalancerId.Name, loadBalancerId.ResourceGroup)
	}

	var outboundRuleId string
	for _, OutboundRule := range *read.LoadBalancerPropertiesFormat.OutboundRules {
		if *OutboundRule.Name == name {
			outboundRuleId = *OutboundRule.ID
		}
	}

	if outboundRuleId == "" {
		return fmt.Errorf("Cannot find created Load Balancer Outbound Rule ID %q", outboundRuleId)
	}

	d.SetId(outboundRuleId)

	return resourceArmLoadBalancerOutboundRuleRead(d, meta)
}

func resourceArmLoadBalancerOutboundRuleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerOutboundRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.ResourceGroup, id.LoadBalancerName)
	loadBalancer, exists, err := retrieveLoadBalancerById(ctx, client, loadBalancerId)
	if err != nil {
		return fmt.Errorf("Error Getting Load Balancer By ID: %+v", err)
	}
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer %q not found. Removing from state", id.LoadBalancerName)
		return nil
	}

	config, _, exists := FindLoadBalancerOutboundRuleByName(loadBalancer, id.Name)
	if !exists {
		d.SetId("")
		log.Printf("[INFO] Load Balancer Outbound Rule %q not found. Removing from state", id.Name)
		return nil
	}

	d.Set("name", config.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if props := config.OutboundRulePropertiesFormat; props != nil {
		allocatedOutboundPorts := 0
		if props.AllocatedOutboundPorts != nil {
			allocatedOutboundPorts = int(*props.AllocatedOutboundPorts)
		}
		d.Set("allocated_outbound_ports", allocatedOutboundPorts)

		backendAddressPoolId := ""
		if props.BackendAddressPool != nil && props.BackendAddressPool.ID != nil {
			bapid, err := parse.LoadBalancerBackendAddressPoolID(*props.BackendAddressPool.ID)
			if err != nil {
				return err
			}

			backendAddressPoolId = bapid.ID(subscriptionId)
		}
		d.Set("backend_address_pool_id", backendAddressPoolId)
		d.Set("enable_tcp_reset", props.EnableTCPReset)

		frontendIpConfigurations := make([]interface{}, 0)
		for _, feConfig := range *props.FrontendIPConfigurations {
			if feConfig.ID == nil {
				continue
			}
			feid, err := parse.LoadBalancerFrontendIPConfigurationID(*feConfig.ID)
			if err != nil {
				return err
			}

			frontendIpConfigurations = append(frontendIpConfigurations, map[string]interface{}{
				"id":   feid.ID(subscriptionId),
				"name": feid.Name,
			})
		}
		d.Set("frontend_ip_configuration", frontendIpConfigurations)

		idleTimeoutInMinutes := 0
		if props.IdleTimeoutInMinutes != nil {
			idleTimeoutInMinutes = int(*props.IdleTimeoutInMinutes)
		}
		d.Set("idle_timeout_in_minutes", idleTimeoutInMinutes)
		d.Set("protocol", string(props.Protocol))
	}

	return nil
}

func resourceArmLoadBalancerOutboundRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Network.LoadBalancersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.LoadBalancerOutboundRuleID(d.Id())
	if err != nil {
		return err
	}

	loadBalancerId := parse.NewLoadBalancerID(id.ResourceGroup, id.LoadBalancerName)
	loadBalancerID := loadBalancerId.ID(subscriptionId)
	locks.ByID(loadBalancerID)
	defer locks.UnlockByID(loadBalancerID)

	loadBalancer, exists, err := retrieveLoadBalancerById(ctx, client, loadBalancerId)
	if err != nil {
		return fmt.Errorf("retrieving Load Balancer By ID: %+v", err)
	}
	if !exists {
		d.SetId("")
		return nil
	}

	_, index, exists := FindLoadBalancerOutboundRuleByName(loadBalancer, id.Name)
	if !exists {
		return nil
	}

	oldOutboundRules := *loadBalancer.LoadBalancerPropertiesFormat.OutboundRules
	newOutboundRules := append(oldOutboundRules[:index], oldOutboundRules[index+1:]...)
	loadBalancer.LoadBalancerPropertiesFormat.OutboundRules = &newOutboundRules

	future, err := client.CreateOrUpdate(ctx, id.ResourceGroup, id.LoadBalancerName, *loadBalancer)
	if err != nil {
		return fmt.Errorf("Creating/Updating Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for completion of Load Balancer %q (Resource Group %q): %+v", id.LoadBalancerName, id.ResourceGroup, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.LoadBalancerName, "")
	if err != nil {
		return fmt.Errorf("Error Getting LoadBalancer: %+v", err)
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read ID of Load Balancer %q (resource group %s)", id.LoadBalancerName, id.ResourceGroup)
	}

	return nil
}

func expandAzureRmLoadBalancerOutboundRule(d *schema.ResourceData, lb *network.LoadBalancer) (*network.OutboundRule, error) {
	properties := network.OutboundRulePropertiesFormat{
		Protocol: network.LoadBalancerOutboundRuleProtocol(d.Get("protocol").(string)),
	}

	feConfigs := d.Get("frontend_ip_configuration").([]interface{})
	feConfigSubResources := make([]network.SubResource, 0)

	for _, raw := range feConfigs {
		v := raw.(map[string]interface{})
		rule, exists := FindLoadBalancerFrontEndIpConfigurationByName(lb, v["name"].(string))
		if !exists {
			return nil, fmt.Errorf("[ERROR] Cannot find FrontEnd IP Configuration with the name %s", v["name"])
		}

		feConfigSubResource := network.SubResource{
			ID: rule.ID,
		}

		feConfigSubResources = append(feConfigSubResources, feConfigSubResource)
	}

	properties.FrontendIPConfigurations = &feConfigSubResources

	if v := d.Get("backend_address_pool_id").(string); v != "" {
		properties.BackendAddressPool = &network.SubResource{
			ID: &v,
		}
	}

	if v, ok := d.GetOk("idle_timeout_in_minutes"); ok {
		properties.IdleTimeoutInMinutes = utils.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("enable_tcp_reset"); ok {
		properties.EnableTCPReset = utils.Bool(v.(bool))
	}

	if v, ok := d.GetOk("allocated_outbound_ports"); ok {
		properties.AllocatedOutboundPorts = utils.Int32(int32(v.(int)))
	}

	return &network.OutboundRule{
		Name:                         utils.String(d.Get("name").(string)),
		OutboundRulePropertiesFormat: &properties,
	}, nil
}
