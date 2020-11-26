package validate

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	dataLakeParse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/datalake/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/datashare/parse"
	StorageParse "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/storage/parse"
)

func DataShareAccountName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[^<>%&:\\?/#*$^();,.\|+={}\[\]!~@]{3,90}$`), `Data share account name should have length of 3 - 90, and cannot contain <>%&:\?/#*$^();,.|+={}[]!~@.`,
	)
}

func DatashareName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^\w{2,90}$`), `DataShare name can only contain alphanumeric characters and _, and must be between 2 and 90 characters long.`,
	)
}

func DatashareAccountID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := parse.AccountID(v); err != nil {
		errors = append(errors, fmt.Errorf("can not parse %q as a Datashare account id: %v", k, err))
	}

	return warnings, errors
}

func DataShareSyncName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[^&%#/]{1,90}$`), `Data share snapshot schedule name should have length of 1 - 90, and cannot contain &%#/`,
	)
}

func DataShareID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := parse.ShareID(v); err != nil {
		errors = append(errors, fmt.Errorf("can not parse %q as a data share id: %v", k, err))
	}

	return warnings, errors
}

func DatashareDataSetName() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile(`^[\w-]{2,90}$`), `Dataset name can only contain number, letters, - and _, and must be between 2 and 90 characters long.`,
	)
}

func DatalakeStoreID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := dataLakeParse.AccountID(v); err != nil {
		errors = append(errors, fmt.Errorf("can not parse %q as a Data Lake Store id: %v", k, err))
	}

	return warnings, errors
}

func StorageAccountID(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
		return warnings, errors
	}

	if _, err := StorageParse.AccountID(v); err != nil {
		errors = append(errors, fmt.Errorf("can not parse %q as a Storage Account id: %v", k, err))
	}

	return warnings, errors
}
