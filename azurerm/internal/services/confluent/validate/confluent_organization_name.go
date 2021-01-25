package validate

import (
	"fmt"
	"regexp"
)

func ConfluentOrganizationName(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}
	if len(v) > 31 {
		errors = append(errors, fmt.Errorf("length should be less than %d", 31))
		return
	}
	if !regexp.MustCompile("^[^<>#%'*^`{|}~\\\\\"]$").MatchString(v) {
		errors = append(errors, fmt.Errorf("%q cannot contain whitespace and special characters:<>#%'*^`{|}~\\\", got %v", k, v))
		return
	}
	return
}
