package parse

import (
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/resourceid"
)

var _ resourceid.Formatter = FrontDoorId{}

func TestFrontDoorIDFormatter(t *testing.T) {
	subscriptionId := "12345678-1234-5678-1234-123456789012"
	actual := NewFrontDoorID("group1", "frontDoor1").ID(subscriptionId)
	expected := "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/frontDoors/frontDoor1"
	if actual != expected {
		t.Fatalf("Expected %q but got %q", expected, actual)
	}
}

func TestFrontDoorIDParser(t *testing.T) {
	testData := []struct {
		input    string
		expected *FrontDoorId
	}{
		{
			// lower case
			input: "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/frontdoors/frontDoor1",
			expected: &FrontDoorId{
				ResourceGroup: "group1",
				Name:          "frontDoor1",
			},
		},
		{
			// camel case
			input: "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/frontDoors/frontDoor1",
			expected: &FrontDoorId{
				ResourceGroup: "group1",
				Name:          "frontDoor1",
			},
		},
		{
			// title case
			input: "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/Frontdoors/frontDoor1",
			expected: &FrontDoorId{
				ResourceGroup: "group1",
				Name:          "frontDoor1",
			},
		},
		{
			// pascal case
			input: "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/FrontDoors/frontDoor1",
			expected: &FrontDoorId{
				ResourceGroup: "group1",
				Name:          "frontDoor1",
			},
		},
	}
	for _, test := range testData {
		t.Logf("Testing %q..", test.input)
		actual, err := FrontDoorID(test.input)
		if err != nil && test.expected == nil {
			continue
		} else {
			if err == nil && test.expected == nil {
				t.Fatalf("Expected an error but didn't get one")
			} else if err != nil && test.expected != nil {
				t.Fatalf("Expected no error but got: %+v", err)
			}
		}

		if actual.ResourceGroup != test.expected.ResourceGroup {
			t.Fatalf("Expected ResourceGroup to be %q but was %q", test.expected.ResourceGroup, actual.ResourceGroup)
		}

		if actual.Name != test.expected.Name {
			t.Fatalf("Expected name to be %q but was %q", test.expected.Name, actual.Name)
		}
	}
}

func TestFrontDoorIDForImportParser(t *testing.T) {
	testData := []struct {
		input    string
		expected *FrontDoorId
	}{
		{
			// lower case
			input:    "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/frontdoors/frontDoor1",
			expected: nil,
		},
		{
			// camel case
			input: "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/frontDoors/frontDoor1",
			expected: &FrontDoorId{
				ResourceGroup: "group1",
				Name:          "frontDoor1",
			},
		},
		{
			// title case
			input:    "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/Frontdoors/frontDoor1",
			expected: nil,
		},
		{
			// pascal case
			input:    "/subscriptions/12345678-1234-5678-1234-123456789012/resourceGroups/group1/providers/Microsoft.Network/FrontDoors/frontDoor1",
			expected: nil,
		},
	}
	for _, test := range testData {
		t.Logf("Testing %q..", test.input)
		actual, err := FrontDoorIDForImport(test.input)
		if err != nil && test.expected == nil {
			continue
		} else {
			if err == nil && test.expected == nil {
				t.Fatalf("Expected an error but didn't get one")
			} else if err != nil && test.expected != nil {
				t.Fatalf("Expected no error but got: %+v", err)
			}
		}

		if actual.ResourceGroup != test.expected.ResourceGroup {
			t.Fatalf("Expected ResourceGroup to be %q but was %q", test.expected.ResourceGroup, actual.ResourceGroup)
		}

		if actual.Name != test.expected.Name {
			t.Fatalf("Expected name to be %q but was %q", test.expected.Name, actual.Name)
		}
	}
}
