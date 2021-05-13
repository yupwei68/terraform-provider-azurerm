module github.com/terraform-providers/terraform-provider-azurerm

require (
	github.com/Azure/azure-sdk-for-go v52.5.0+incompatible
	github.com/Azure/azure-sdk-for-go/sdk/arm/avs v0.0.0-00010101000000-000000000000
	github.com/Azure/azure-sdk-for-go/sdk/arm/compute v0.0.0-00010101000000-000000000000
	github.com/Azure/azure-sdk-for-go/sdk/arm/storage/2019-06-01/armstorage v0.2.0
	github.com/Azure/azure-sdk-for-go/sdk/armcore v0.7.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.16.0
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/date v0.3.0
	github.com/Azure/go-autorest/autorest/validation v0.3.1
	github.com/btubbs/datetime v0.1.0
	github.com/davecgh/go-spew v1.1.1
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.1
	github.com/hashicorp/go-azure-helpers v0.14.0
	github.com/hashicorp/go-getter v1.5.2
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.16.1-0.20210222152151-32f0219df5b5
	github.com/rickb777/date v1.12.5-0.20200422084442-6300e543c4d9
	github.com/sergi/go-diff v1.1.0
	github.com/terraform-providers/terraform-provider-azuread v0.9.0
	github.com/tombuildsstuff/giovanni v0.15.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/hashicorp/go-azure-helpers => github.com/ArcturusZhang/go-azure-helpers v0.16.3

replace github.com/Azure/azure-sdk-for-go/sdk/arm/compute => github.com/ArcturusZhang/azure-sdk-for-go/sdk/arm/compute v0.3.0

replace github.com/Azure/azure-sdk-for-go/sdk/arm/storage => github.com/ArcturusZhang/azure-sdk-for-go/sdk/arm/storage v0.1.1

replace github.com/Azure/azure-sdk-for-go/sdk/arm/avs => github.com/ArcturusZhang/azure-sdk-for-go/sdk/arm/avs v0.1.1

go 1.16
