package validate

import (
	"fmt"
)

func ConfluentOrganizationTermUnit(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}
	if len(v) > 25 {
		errors = append(errors, fmt.Errorf("length should be less than %d", 25))
		return
	}
	return
}
