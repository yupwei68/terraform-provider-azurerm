package parse

import (
	"testing"
)

func TestSynapseSparkPoolID(t *testing.T) {
	testData := []struct {
		Name     string
		Input    string
		Expected *SynapseSparkPoolId
	}{
		{
			Name:     "Empty",
			Input:    "",
			Expected: nil,
		},
		{
			Name:     "No Resource Groups Segment",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000",
			Expected: nil,
		},
		{
			Name:     "No Resource Groups Value",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/",
			Expected: nil,
		},
		{
			Name:     "Resource Group ID",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foo/",
			Expected: nil,
		},
		{
			Name:     "Missing BigDataPool Value",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.Synapse/workspaces/workspace1/bigDataPools",
			Expected: nil,
		},
		{
			Name:  "synapse BigDataPool ID",
			Input: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.Synapse/workspaces/workspace1/bigDataPools/sparkPool1",
			Expected: &SynapseSparkPoolId{
				Workspace: &SynapseWorkspaceId{
					ResourceGroup: "resourceGroup1",
					Name:          "workspace1",
				},
				Name: "sparkPool1",
			},
		},
		{
			Name:     "Wrong Casing",
			Input:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourceGroup1/providers/Microsoft.Synapse/workspaces/workspace1/BigDataPools/sparkPool1",
			Expected: nil,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q..", v.Name)

		actual, err := SynapseSparkPoolID(v.Input)
		if err != nil {
			if v.Expected == nil {
				continue
			}
			t.Fatalf("Expected a value but got an error: %s", err)
		}

		if actual.Workspace.ResourceGroup != v.Expected.Workspace.ResourceGroup {
			t.Fatalf("Expected %q but got %q for ResourceGroup", v.Expected.Workspace.ResourceGroup, actual.Workspace.ResourceGroup)
		}

		if actual.Workspace.Name != v.Expected.Workspace.Name {
			t.Fatalf("Expected %q but got %q for WorkspaceName", v.Expected.Workspace.Name, actual.Workspace.Name)
		}

		if actual.Name != v.Expected.Name {
			t.Fatalf("Expected %q but got %q for Name", v.Expected.Name, actual.Name)
		}
	}
}
