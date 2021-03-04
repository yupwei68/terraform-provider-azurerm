package validate

import (
	"fmt"
	"math"
)

func ClusterCapacity(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(int)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be integer", k))
		return warnings, errors
	}

	if v < 36 || v > 216 {
		errors = append(errors, fmt.Errorf("expected %s to be in the range (%d - %d), got %d", k, 36, 216, v))
		return warnings, errors
	}

	if math.Mod(float64(v), float64(36)) != 0 {
		errors = append(errors, fmt.Errorf("expected %s to be divisible by %d, got: %v", k, 36, i))
		return warnings, errors
	}

	return warnings, errors
}
