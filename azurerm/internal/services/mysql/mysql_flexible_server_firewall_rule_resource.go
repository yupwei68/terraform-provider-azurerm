package mysql

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/mysql/mgmt/2020-07-01-preview/mysqlflexibleservers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mysql/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/mysql/validate"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
	"log"
	"time"
)

func resourceMysqlFlexibleServerFirewallRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceMysqlFlexibleServerFirewallRuleCreateUpdate,
		Read:   resourceMysqlFlexibleServerFirewallRuleRead,
		Update: resourceMysqlFlexibleServerFirewallRuleCreateUpdate,
		Delete: resourceMysqlFlexibleServerFirewallRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.FlexibleServerFirewallRuleID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"flexible_server_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FlexibleServerID,
			},

			"end_ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
			},

			"start_ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	}
}
func resourceMysqlFlexibleServerFirewallRuleCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).MySQL.FlexibleServerFirewallRulesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	serverId, err := parse.FlexibleServerID(d.Get("flexible_server_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewFlexibleServerFirewallRuleID(subscriptionId, serverId.ResourceGroup, serverId.Name, name).ID()

	if d.IsNewResource() {
		existing, err := client.Get(ctx, serverId.ResourceGroup, serverId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_mysql_flexible_server_firewall_rule", id)
		}
	}

	parameters := mysqlflexibleservers.FirewallRule{
		FirewallRuleProperties: &mysqlflexibleservers.FirewallRuleProperties{
			EndIPAddress:   utils.String(d.Get("end_ip_address").(string)),
			StartIPAddress: utils.String(d.Get("start_ip_address").(string)),
		},
	}

	future, err := client.CreateOrUpdate(ctx, serverId.ResourceGroup, serverId.Name, name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on creation/update of the Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverId.Name %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	d.SetId(id)

	return resourceMysqlFlexibleServerFirewallRuleRead(d, meta)
}

func resourceMysqlFlexibleServerFirewallRuleRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).MySQL.FlexibleServerFirewallRulesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerFirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.FirewallRuleName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Mysql Flexible Server %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverName %q): %+v", id.FirewallRuleName, id.ResourceGroup, id.FlexibleServerName, err)
	}
	d.Set("name", id.FirewallRuleName)
	d.Set("flexible_server_id", parse.NewFlexibleServerID(subscriptionId, id.ResourceGroup, id.FlexibleServerName).ID())
	if props := resp.FirewallRuleProperties; props != nil {
		d.Set("end_ip_address", props.EndIPAddress)
		d.Set("start_ip_address", props.StartIPAddress)
	}
	return nil
}

func resourceMysqlFlexibleServerFirewallRuleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MySQL.FlexibleServerFirewallRulesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerFirewallRuleID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.FlexibleServerName, id.FirewallRuleName)
	if err != nil {
		return fmt.Errorf("deleting Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverName %q): %+v", id.FirewallRuleName, id.ResourceGroup, id.FlexibleServerName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting on deletion of the Mysql Flexible Server Firewall Rule %q (Resource Group %q / serverName %q): %+v", id.FirewallRuleName, id.ResourceGroup, id.FlexibleServerName, err)
	}
	return nil
}
