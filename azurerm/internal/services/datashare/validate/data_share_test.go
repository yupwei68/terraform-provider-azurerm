package validate

import "testing"

func TestDataShareAccountName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "test 1",
			input: "DC\\",
			valid: false,
		},
		{
			name:  "test 2",
			input: "[abc]",
			valid: false,
		},
		{
			name:  "test 3",
			input: "acc-test",
			valid: true,
		},
		{
			name:  "test 4",
			input: "test&",
			valid: false,
		},
		{
			name:  "test 5",
			input: "ab",
			valid: false,
		},
		{
			name:  "test 6",
			input: "aa-BB_88",
			valid: true,
		},
		{
			name:  "test 7",
			input: "aac-",
			valid: true,
		},
	}
	var validationFunction = DataShareAccountName()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validationFunction(tt.input, "")
			valid := err == nil
			if valid != tt.valid {
				t.Errorf("Expected valid status %t but got %t for input %s", tt.valid, valid, tt.input)
			}
		})
	}
}
