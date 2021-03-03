package azure

import "testing"

func TestQuotedStringSlice(t *testing.T) {
	testData := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{""},
			expected: `""`,
		},
		{
			input:    []string{"Din Djarin"},
			expected: `"Din Djarin"`,
		},
		{
			input:    []string{"Baby Yoda", "Grogu"},
			expected: `"Baby Yoda" or "Grogu"`,
		},
		{
			input:    []string{"This", "is", "the", "way"},
			expected: `"This", "is", "the" or "way"`,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q..", v.input)

		actual := QuotedStringSlice(v.input)
		if v.expected != actual {
			t.Fatalf("Expected %s but got %s", v.expected, actual)
		}
	}
}
