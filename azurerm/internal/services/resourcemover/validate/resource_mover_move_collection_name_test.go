package validate

import (
	"testing"
)

func TestResourceMoverMoveCollection(t *testing.T) {
	tests := []struct {
		Name  string
		Input string
		Valid bool
	}{
		{
			Name:  "Empty",
			Input: "",
			Valid: false,
		},
		{
			Name:  "Valid Name 1",
			Input: "a",
			Valid: true,
		},
		{
			Name:  "Invalid character",
			Input: "resource_mover",
			Valid: false,
		},
		{
			Name:  "Valid Name 2",
			Input: "Resource-Mover",
			Valid: true,
		},
		{
			Name:  "End with `-`",
			Input: "Resource-Mover-",
			Valid: false,
		},
		{
			Name:  "Start with `-`",
			Input: "-Resource-Mover",
			Valid: false,
		},
		{
			Name:  "Invalid character",
			Input: "Resource.Mover",
			Valid: false,
		},
		{
			Name:  "Too long",
			Input: "ResourceMoverResourceMoverResourceMoverResourceMoverResourceMover",
			Valid: false,
		},
		{
			Name:  "Max characters",
			Input: "ResourceMoverResourceMoverResourceMoverResourceMoverResourceMove",
			Valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := ResourceMoverMoveCollectionName(tt.Input, "Name")
			valid := err == nil
			if valid != tt.Valid {
				t.Errorf("Expected valid status %t but got %t for input %s", tt.Valid, valid, tt.Input)
			}
		})
	}
}
