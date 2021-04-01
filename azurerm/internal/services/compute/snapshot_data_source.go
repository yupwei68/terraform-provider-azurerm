package compute

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSnapshotRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			// Computed
			"os_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size_gb": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"time_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_option": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"disk_encryption_key": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"secret_url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"source_vault_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"key_encryption_key": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key_url": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"source_vault_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.SnapshotsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	resp, err := client.Get(ctx, resourceGroup, name, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			return fmt.Errorf("Error: Snapshot %q (Resource Group %q) was not found", name, resourceGroup)
		}
		return fmt.Errorf("Error loading Snapshot %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	snapshot := resp.Snapshot
	d.SetId(*snapshot.ID)

	if props := snapshot.Properties; props != nil {
		if props.OSType != nil {
			d.Set("os_type", string(*props.OSType))
		}
		d.Set("time_created", props.TimeCreated.String())

		if props.DiskSizeGB != nil {
			d.Set("disk_size_gb", int(*props.DiskSizeGB))
		}

		if err := d.Set("encryption_settings", flattenManagedDiskEncryptionSettings(props.EncryptionSettingsCollection)); err != nil {
			return fmt.Errorf("Error setting `encryption_settings`: %+v", err)
		}
	}

	if data := snapshot.Properties.CreationData; data != nil {
		if data.CreateOption != nil {
			d.Set("creation_option", string(*data.CreateOption))
		}
		d.Set("source_uri", data.SourceURI)
		d.Set("source_resource_id", data.SourceResourceID)
		d.Set("storage_account_id", data.StorageAccountID)
	}

	return nil
}
