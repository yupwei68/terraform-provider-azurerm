package compute

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/arm/compute/2020-12-01/armcompute"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-12-01/compute"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/locks"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceVirtualMachineDataDiskAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualMachineDataDiskAttachmentCreateUpdate,
		Read:   resourceVirtualMachineDataDiskAttachmentRead,
		Update: resourceVirtualMachineDataDiskAttachmentCreateUpdate,
		Delete: resourceVirtualMachineDataDiskAttachmentDelete,
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
			"managed_disk_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     azure.ValidateResourceID,
			},

			"virtual_machine_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"lun": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"caching": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.CachingTypesNone),
					string(compute.CachingTypesReadOnly),
					string(compute.CachingTypesReadWrite),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"create_option": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  string(compute.DiskCreateOptionTypesAttach),
				ValidateFunc: validation.StringInSlice([]string{
					string(compute.DiskCreateOptionTypesAttach),
					string(compute.DiskCreateOptionTypesEmpty),
				}, true),
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"write_accelerator_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceVirtualMachineDataDiskAttachmentCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	virtualMachineId := d.Get("virtual_machine_id").(string)
	parsedVirtualMachineId, err := azure.ParseAzureResourceID(virtualMachineId)
	if err != nil {
		return fmt.Errorf("Error parsing Virtual Machine ID %q: %+v", virtualMachineId, err)
	}

	resourceGroup := parsedVirtualMachineId.ResourceGroup
	virtualMachineName := parsedVirtualMachineId.Path["virtualMachines"]

	locks.ByName(virtualMachineName, virtualMachineResourceName)
	defer locks.UnlockByName(virtualMachineName, virtualMachineResourceName)

	virtualMachine, err := client.Get(ctx, resourceGroup, virtualMachineName, "")
	if err != nil {
		if utils.ResponseWasNotFound(virtualMachine.Response) {
			return fmt.Errorf("Virtual Machine %q (Resource Group %q) was not found", virtualMachineName, resourceGroup)
		}

		return fmt.Errorf("Error loading Virtual Machine %q (Resource Group %q): %+v", virtualMachineName, resourceGroup, err)
	}

	managedDiskId := d.Get("managed_disk_id").(string)
	managedDisk, err := retrieveDataDiskAttachmentManagedDisk(d, meta, managedDiskId)
	if err != nil {
		return fmt.Errorf("Error retrieving Managed Disk %q: %+v", managedDiskId, err)
	}

	if managedDisk.SKU == nil {
		return fmt.Errorf("Error: unable to determine Storage Account Type for Managed Disk %q: %+v", managedDiskId, err)
	}

	name := *managedDisk.Name
	resourceId := fmt.Sprintf("%s/dataDisks/%s", virtualMachineId, name)
	lun := int32(d.Get("lun").(int))
	caching := d.Get("caching").(string)
	createOption := compute.DiskCreateOptionTypes(d.Get("create_option").(string))
	writeAcceleratorEnabled := d.Get("write_accelerator_enabled").(bool)

	expandedDisk := compute.DataDisk{
		Name:         utils.String(name),
		Caching:      compute.CachingTypes(caching),
		CreateOption: createOption,
		Lun:          utils.Int32(lun),
		ManagedDisk: &compute.ManagedDiskParameters{
			ID:                 utils.String(managedDiskId),
			StorageAccountType: compute.StorageAccountTypes(*managedDisk.SKU.Name),
		},
		WriteAcceleratorEnabled: utils.Bool(writeAcceleratorEnabled),
	}

	disks := *virtualMachine.StorageProfile.DataDisks

	existingIndex := -1
	for i, disk := range disks {
		if *disk.Name == name {
			existingIndex = i
			break
		}
	}

	if d.IsNewResource() {
		if existingIndex != -1 {
			return tf.ImportAsExistsError("azurerm_virtual_machine_data_disk_attachment", resourceId)
		}

		disks = append(disks, expandedDisk)
	} else {
		if existingIndex == -1 {
			return fmt.Errorf("Unable to find Disk %q attached to Virtual Machine %q (Resource Group %q)", name, virtualMachineName, resourceGroup)
		}

		disks[existingIndex] = expandedDisk
	}

	virtualMachine.StorageProfile.DataDisks = &disks

	// fixes #2485
	virtualMachine.Identity = nil
	// fixes #1600
	virtualMachine.Resources = nil

	// if there's too many disks we get a 409 back with:
	//   `The maximum number of data disks allowed to be attached to a VM of this size is 1.`
	// which we're intentionally not wrapping, since the errors good.
	future, err := client.CreateOrUpdate(ctx, resourceGroup, virtualMachineName, virtualMachine)
	if err != nil {
		return fmt.Errorf("Error updating Virtual Machine %q (Resource Group %q) with Disk %q: %+v", virtualMachineName, resourceGroup, name, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for Virtual Machine %q (Resource Group %q) to finish updating Disk %q: %+v", virtualMachineName, resourceGroup, name, err)
	}

	d.SetId(resourceId)
	return resourceVirtualMachineDataDiskAttachmentRead(d, meta)
}

func resourceVirtualMachineDataDiskAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	virtualMachineName := id.Path["virtualMachines"]
	name := id.Path["dataDisks"]

	virtualMachine, err := client.Get(ctx, resourceGroup, virtualMachineName, "")
	if err != nil {
		if utils.ResponseWasNotFound(virtualMachine.Response) {
			log.Printf("[DEBUG] Virtual Machine %q was not found (Resource Group %q) therefore Data Disk Attachment cannot exist - removing from state", virtualMachineName, resourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error loading Virtual Machine %q (Resource Group %q): %+v", virtualMachineName, resourceGroup, err)
	}

	var disk *compute.DataDisk
	if profile := virtualMachine.StorageProfile; profile != nil {
		if dataDisks := profile.DataDisks; dataDisks != nil {
			for _, dataDisk := range *dataDisks {
				// since this field isn't (and shouldn't be) case-sensitive; we're deliberately not using `strings.EqualFold`
				if *dataDisk.Name == name {
					disk = &dataDisk
					break
				}
			}
		}
	}

	if disk == nil {
		log.Printf("[DEBUG] Data Disk %q was not found on Virtual Machine %q (Resource Group %q) - removing from state", name, virtualMachineName, resourceGroup)
		d.SetId("")
		return nil
	}

	d.Set("virtual_machine_id", virtualMachine.ID)
	d.Set("caching", string(disk.Caching))
	d.Set("create_option", string(disk.CreateOption))
	d.Set("write_accelerator_enabled", disk.WriteAcceleratorEnabled)

	if managedDisk := disk.ManagedDisk; managedDisk != nil {
		d.Set("managed_disk_id", managedDisk.ID)
	}

	if lun := disk.Lun; lun != nil {
		d.Set("lun", int(*lun))
	}

	return nil
}

func resourceVirtualMachineDataDiskAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Compute.VMClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	resourceGroup := id.ResourceGroup
	virtualMachineName := id.Path["virtualMachines"]
	name := id.Path["dataDisks"]

	locks.ByName(virtualMachineName, virtualMachineResourceName)
	defer locks.UnlockByName(virtualMachineName, virtualMachineResourceName)

	virtualMachine, err := client.Get(ctx, resourceGroup, virtualMachineName, "")
	if err != nil {
		if utils.ResponseWasNotFound(virtualMachine.Response) {
			return fmt.Errorf("Virtual Machine %q (Resource Group %q) was not found", virtualMachineName, resourceGroup)
		}

		return fmt.Errorf("Error loading Virtual Machine %q (Resource Group %q): %+v", virtualMachineName, resourceGroup, err)
	}

	dataDisks := make([]compute.DataDisk, 0)
	for _, dataDisk := range *virtualMachine.StorageProfile.DataDisks {
		// since this field isn't (and shouldn't be) case-sensitive; we're deliberately not using `strings.EqualFold`
		if *dataDisk.Name != name {
			dataDisks = append(dataDisks, dataDisk)
		}
	}

	virtualMachine.StorageProfile.DataDisks = &dataDisks

	// fixes #2485
	virtualMachine.Identity = nil
	// fixes #1600
	virtualMachine.Resources = nil

	future, err := client.CreateOrUpdate(ctx, resourceGroup, virtualMachineName, virtualMachine)
	if err != nil {
		return fmt.Errorf("Error removing Disk %q from Virtual Machine %q (Resource Group %q): %+v", name, virtualMachineName, resourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for Disk %q to be removed from Virtual Machine %q (Resource Group %q): %+v", name, virtualMachineName, resourceGroup, err)
	}

	return nil
}

func retrieveDataDiskAttachmentManagedDisk(d *schema.ResourceData, meta interface{}, id string) (*armcompute.Disk, error) {
	client := meta.(*clients.Client).Compute.DisksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	parsedId, err := azure.ParseAzureResourceID(id)
	if err != nil {
		return nil, fmt.Errorf("Error parsing Managed Disk ID %q: %+v", id, err)
	}
	resourceGroup := parsedId.ResourceGroup
	name := parsedId.Path["disks"]

	resp, err := client.Get(ctx, resourceGroup, name, nil)
	if err != nil {
		if utils.Track2ResponseWasNotFound(err) {
			return nil, fmt.Errorf("Error Managed Disk %q (Resource Group %q) was not found!", name, resourceGroup)
		}

		return nil, fmt.Errorf("Error making Read request on Azure Managed Disk %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return resp.Disk, nil
}
