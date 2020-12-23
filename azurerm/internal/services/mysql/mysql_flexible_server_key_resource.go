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

func resourceMysqlFlexibleServerKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceMysqlFlexibleServerKeyCreateUpdate,
		Read:   resourceMysqlFlexibleServerKeyRead,
		Update: resourceMysqlFlexibleServerKeyCreateUpdate,
		Delete: resourceMysqlFlexibleServerKeyDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.FlexibleServerKeyID(id)
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

			"server_key_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AzureKeyVault",
				}, false),
			},

			"key_vault_key_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
func resourceMysqlFlexibleServerKeyCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).MySQL.FlexibleServerKeysClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	serverId, err := parse.FlexibleServerID(d.Get("flexible_server_id").(string))
	if err != nil {
		return err
	}

	id := parse.NewFlexibleServerKeyID(subscriptionId, serverId.ResourceGroup, serverId.Name, name).ID()

	if d.IsNewResource() {
		existing, err := client.Get(ctx, serverId.ResourceGroup, serverId.Name, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for present of existing MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
			}
		}
		if !utils.ResponseWasNotFound(existing.Response) {
			return tf.ImportAsExistsError("azurerm_mysql_flexible_server_key", id)
		}
	}

	parameters := mysqlflexibleservers.ServerKey{
		ServerKeyProperties: &mysqlflexibleservers.ServerKeyProperties{
			ServerKeyType: utils.String(d.Get("server_key_type").(string)),
			URI:           utils.String(d.Get("key_vault_key_id").(string)),
		},
	}
	future, err := client.CreateOrUpdate(ctx, serverId.ResourceGroup, serverId.Name, name, parameters)
	if err != nil {
		return fmt.Errorf("creating/updating MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the creation/update of the MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	if _, err := client.Get(ctx, serverId.ResourceGroup, serverId.Name, name); err != nil {
		return fmt.Errorf("retrieving MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", name, serverId.ResourceGroup, serverId.Name, err)
	}

	d.SetId(id)

	return resourceMysqlFlexibleServerKeyRead(d, meta)
}

func resourceMysqlFlexibleServerKeyRead(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).MySQL.ServerKeysClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerKeyID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.KeyName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Mysql Flexible Servers %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", id.KeyName, id.ResourceGroup, id.FlexibleServerName, err)
	}

	d.Set("name", id.KeyName)
	d.Set("flexible_server_id", parse.NewFlexibleServerID(subscriptionId, id.ResourceGroup, id.FlexibleServerName).ID())
	if props := resp.ServerKeyProperties; props != nil {
		d.Set("server_key_type", props.ServerKeyType)
		d.Set("key_vault_key_id", props.URI)
	}
	return nil
}

func resourceMysqlFlexibleServerKeyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).MySQL.FlexibleServerKeysClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerKeyID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.FlexibleServerName, id.KeyName)
	if err != nil {
		return fmt.Errorf("deleting MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", id.KeyName, id.ResourceGroup, id.FlexibleServerName, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of the MySql Flexible Server Key %q (Resource Group %q / serverName %q): %+v", id.KeyName, id.ResourceGroup, id.FlexibleServerName, err)
	}
	return nil
}
