package postgres

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2021-06-01/postgresqlflexibleservers"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/postgres/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/postgres/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourcePostgresqlFlexibleServerDatabase() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourcePostgresqlFlexibleServerDatabaseCreate,
		Read:   resourcePostgresqlFlexibleServerDatabaseRead,
		Delete: resourcePostgresqlFlexibleServerDatabaseDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.FlexibleServerDatabaseID(id)
			return err
		}),

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"server_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.FlexibleServerID,
			},

			"charset": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringIsNotEmpty,
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"collation": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}
func resourcePostgresqlFlexibleServerDatabaseCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Postgres.FlexibleServerDatabasesClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	serverId, err := parse.FlexibleServerID(d.Get("server_id").(string))

	id := parse.NewFlexibleServerDatabaseID(subscriptionId, serverId.ResourceGroup, serverId.Name, name)

	existing, err := client.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.DatabaseName)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing %q: %+v", id, err)
		}
	}
	if !utils.ResponseWasNotFound(existing.Response) {
		return tf.ImportAsExistsError("azurerm_postgresql_flexible_server_database", id.ID())
	}

	props := postgresqlflexibleservers.Database{
		DatabaseProperties: &postgresqlflexibleservers.DatabaseProperties{},
	}
	if v, ok := d.GetOk("charset"); ok && v.(string) != "" {
		props.DatabaseProperties.Charset = utils.String(v.(string))
	}

	if v, ok := d.GetOk("collation"); ok && v.(string) != "" {
		props.DatabaseProperties.Collation = utils.String(v.(string))
	}

	future, err := client.Create(ctx, id.ResourceGroup, id.FlexibleServerName, id.DatabaseName, props)
	if err != nil {
		return fmt.Errorf("creating %q: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of the %q: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourcePostgresqlFlexibleServerDatabaseRead(d, meta)
}

func resourcePostgresqlFlexibleServerDatabaseRead(d *pluginsdk.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).Postgres.FlexibleServerDatabasesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerDatabaseID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.FlexibleServerName, id.DatabaseName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Postgresql FlexibleServer Database %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving %q: %+v", id, err)
	}

	d.Set("name", id.DatabaseName)
	d.Set("server_id", parse.NewFlexibleServerID(subscriptionId, id.ResourceGroup, id.FlexibleServerName).ID())
	if props := resp.DatabaseProperties; props != nil {
		d.Set("charset", props.Charset)
		d.Set("collation", props.Collation)
	}

	return nil
}

func resourcePostgresqlFlexibleServerDatabaseDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Postgres.FlexibleServerDatabasesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FlexibleServerDatabaseID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.FlexibleServerName, id.DatabaseName)
	if err != nil {
		return fmt.Errorf("deleting %q: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of the %q: %+v", id, err)
	}

	return nil
}
