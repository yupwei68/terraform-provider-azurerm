package validate

import (
	"fmt"
	"regexp"
)

func ResourceMoverMoveCollectionName(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if len(v) < 1 {
		errors = append(errors, fmt.Errorf("length should equal to or greater than %d, got %q", 1, v))
		return
	}

	if len(v) > 64 {
		errors = append(errors, fmt.Errorf("length should be equal to or less than %d, got %q", 64, v))
		return
	}

	if !regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]+[A-Za-z0-9])?$`).MatchString(v) {
		errors = append(errors, fmt.Errorf("%q can contain only letters, numbers and '-'. The '-' shouldn't be the first or the last symbol, got %v", k, v))
		return
	}
	return
}
