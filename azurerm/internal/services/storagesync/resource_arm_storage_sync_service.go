package storagesync

import (
    "fmt"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/validate"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
    "github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
    "log"
    "time"
)

func resourceArmStorageSyncService() *schema.Resource {
    return &schema.Resource{
        Create: resourceArmStorageSyncServiceCreateUpdate,
        Read: resourceArmStorageSyncServiceRead,
        Update: resourceArmStorageSyncServiceCreateUpdate,
        Delete: resourceArmStorageSyncServiceDelete,

        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },

        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(30 * time.Minute),
            Read:   schema.DefaultTimeout(5 * time.Minute),
            Update: schema.DefaultTimeout(30 * time.Minute),
            Delete: schema.DefaultTimeout(30 * time.Minute),
        },


        Schema: map[string]*schema.Schema{
            "name": {
                Type: schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validate.NoEmptyStrings,
            },
            "resource_group": azure.SchemaResourceGroupName(),

            "location": azure.SchemaLocation(),

            "storage_sync_service_status": {
                Type: schema.TypeInt,
                Computed: true,
            },

            "storage_sync_service_uid": {
                Type: schema.TypeString,
                Computed: true,
            },

            "tags": tags.Schema(),
        },
    }
}

func resourceArmStorageSyncServiceCreateUpdate(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*clients.Client).StorageSync.storageSyncServicesClient
    ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
    defer cancel()

    name := d.Get("name").(string)
    resourceGroupName := d.Get("resource_group").(string)


    if features.ShouldResourcesBeImported(){
        existing, err := client.Get(ctx, resourceGroupName, name)
        if err != nil {
            if !utils.ResponseWasNotFound(existing.Response) {
                return fmt.Errorf("Error checking for present of existing Storage Sync Service (Storage Sync Service Name %q / Resource Group %q): %+v", name, resourceGroupName, err)
            }
        }
        if existing.ID != nil && *existing.ID != "" {
            return tf.ImportAsExistsError("azurerm_storage_sync_service", *existing.ID)
        }
    }

    location := azure.NormalizeLocation(d.Get("location").(string))
    tags := d.Get("tags").(map[string]interface{})

    parameters := storagesync.ServiceUpdateParameters{
        Location: utils.String(location),
        Name: utils.String(name),
        Tags: tags.Expand(tags),
    }


    if _, err := client.Create(ctx, resourceGroupName, name, parameters); err != nil {
        return fmt.Errorf("Error creating Storage Sync Service (Storage Sync Service Name %q / Resource Group %q): %+v", name, resourceGroupName, err)
    }


    resp, err := client.Get(ctx, resourceGroupName, name)
    if err != nil {
        return fmt.Errorf("Error retrieving Storage Sync Service (Storage Sync Service Name %q / Resource Group %q): %+v", name, resourceGroupName, err)
    }
    if resp.ID == nil {
        return fmt.Errorf("Cannot read Storage Sync Service (Storage Sync Service Name %q / Resource Group %q) ID", name, resourceGroupName)
    }
    d.SetId(*resp.ID)

    return resourceArmStorageSyncServiceRead(d, meta)
}

func resourceArmStorageSyncServiceRead(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*clients.Client).StorageSync.storageSyncServicesClient
    ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
    defer cancel()

    id, err := azure.ParseAzureResourceID(d.Id())
    if err != nil {
        return err
    }
    resourceGroupName := id.ResourceGroup
    name := id.Path["storageSyncServices"]

    resp, err := client.Get(ctx, resourceGroupName, name)
    if err != nil {
        if utils.ResponseWasNotFound(resp.Response) {
            log.Printf("[INFO] Storage Sync Service %q does not exist - removing from state", d.Id())
            d.SetId("")
            return nil
        }
        return fmt.Errorf("Error reading Storage Sync Service (Storage Sync Service Name %q / Resource Group %q): %+v", name, resourceGroupName, err)
    }

    d.Set("name", resp.Name)
    d.Set("resource_group", resourceGroupName)
    if location := resp.Location; location != nil {
        d.Set("location", azure.NormalizeLocation(*location))
    }

    if object := resp.Object; object != nil {
        d.Set("storage_sync_service_status", object.StorageSyncServiceStatus)
        d.Set("storage_sync_service_uid", object.StorageSyncServiceUID)
    }

    return tags.FlattenAndSet(d, resp.Tags)
}


func resourceArmStorageSyncServiceDelete(d *schema.ResourceData, meta interface{}) error {
    client := meta.(*clients.Client).StorageSync.storageSyncServicesClient
    ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
    defer cancel()


    id, err := azure.ParseAzureResourceID(d.Id())
    if err != nil {
        return err
    }
    resourceGroupName := id.ResourceGroup
    name := id.Path["storageSyncServices"]

    if _, err := client.Delete(ctx, resourceGroupName, name); err != nil {
        return fmt.Errorf("Error deleting Storage Sync Service (Storage Sync Service Name %q / Resource Group %q): %+v", name, resourceGroupName, err)
    }

    return nil
}
