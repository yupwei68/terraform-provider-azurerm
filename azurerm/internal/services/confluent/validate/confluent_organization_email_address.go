package validate

import (
	"fmt"
	"regexp"
)

func ConfluentOrganizationEmailAddress(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}
	if !regexp.MustCompile(`^\S+@\S+\.\S+$`).MatchString(v) {
		errors = append(errors, fmt.Errorf("expected value of %s not match regular expression, got %v", k, v))
		return
	}
	return
}
