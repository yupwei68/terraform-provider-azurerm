package validate

import (
	"testing"
)

func TestClusterCapacity(t *testing.T) {
	tests := []struct {
		Name  string
		Input int
		Valid bool
	}{
		{
			Name:  "1",
			Input: 1,
			Valid: false,
		},
		{
			Name:  "24",
			Input: 24,
			Valid: false,
		},
		{
			Name:  "36",
			Input: 36,
			Valid: true,
		},
		{
			Name:  "108",
			Input: 108,
			Valid: true,
		},
		{
			Name:  "100",
			Input: 100,
			Valid: false,
		},
		{
			Name:  "216",
			Input: 216,
			Valid: true,
		},
		{
			Name:  "252",
			Input: 252,
			Valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := ClusterCapacity(tt.Input, "Name")
			valid := err == nil
			if valid != tt.Valid {
				t.Errorf("Expected valid status %t but got %t for input %v", tt.Valid, valid, tt.Input)
			}
		})
	}
}
