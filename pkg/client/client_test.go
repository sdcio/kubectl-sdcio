package client

import (
	"testing"

	tu "github.com/sdcio/kubectl-sdcio/pkg/testutils"
	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestBlameTree() *sdcpb.BlameTreeElement {
	//current running folder is expected to be pkg/client
	tree, err := tu.LoadBlameTreeFromFile("../testutils/blame_test.yaml")
	if err != nil {

		return nil
	}

	return tree.ToProtobuf()
}

func TestFilterBlameTree_FilterByLeafName(t *testing.T) {
	tests := []struct {
		name           string
		filter         BlameFilter
		expectedLeaves []string
	}{
		{
			name: "Filter by exact leaf name",
			filter: BlameFilter{
				LeafName: "ambulance",
			},
			expectedLeaves: []string{"ambulance"},
		},
		{
			name: "Filter by wildcard leaf name",
			filter: BlameFilter{
				LeafName: "*brigade*",
			},
			expectedLeaves: []string{"fire-brigade"},
		},
		{
			name: "Filter by multiple matches",
			filter: BlameFilter{
				LeafName: "*type*",
			},
			expectedLeaves: []string{"appl-type"},
		},
		{
			name: "No matches",
			filter: BlameFilter{
				LeafName: "nonexistent",
			},
			expectedLeaves: []string{},
		},
	}

	client := &ConfigClient{}
	tree := createTestBlameTree()
	require.NotNil(t, tree, "Expected sample tree loaded - start test from client pkg folder")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.filterBlameTree(tree, []string{}, tt.filter)

			if len(tt.expectedLeaves) == 0 {
				assert.Nil(t, result, "Expected no results for filter: %s", tt.filter.LeafName)
				return
			}

			require.NotNil(t, result, "Expected results for filter: %s", tt.filter.LeafName)

			// Collect all leaves to compare to expected ones
			leaves := tu.CollectLeaves(result)
			leafNames := make([]string, len(leaves))
			for i, leaf := range leaves {
				leafNames[i] = leaf.Name
			}

			assert.ElementsMatch(t, tt.expectedLeaves, leafNames)
		})
	}
}

func TestFilterBlameTree_FilterDeviation(t *testing.T) {
	tests := []struct {
		name           string
		filter         BlameFilter
		expectedLeaves []string
	}{
		{
			name: "Filter by exact leaf name",
			filter: BlameFilter{
				LeafName:  "ambulance",
				Deviation: false,
			},
			expectedLeaves: []string{"false"},
		},
		{
			name: "Filter by exact leaf name and deviation",
			filter: BlameFilter{
				LeafName:  "ambulance",
				Deviation: true,
			},
			expectedLeaves: []string{},
		},
		{
			name: "Filter by deviation",
			filter: BlameFilter{
				Deviation: true,
			},
			expectedLeaves: []string{"6000"},
		},
	}

	client := &ConfigClient{}
	tree := createTestBlameTree()
	require.NotNil(t, tree, "Expected sample tree loaded - start test from client pkg folder")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.filterBlameTree(tree, []string{}, tt.filter)

			if len(tt.expectedLeaves) == 0 {
				assert.Nil(t, result, "Expected no results for filter: (%s) %s", tt.name, tt.filter.LeafName)
				return
			}

			require.NotNil(t, result, "Expected results for filter: (%s) %s", tt.name, tt.filter.LeafName)

			// Collect all leaves to compare to expected ones (values)
			leaves := tu.CollectLeaves(result)
			leafValues := make([]string, len(leaves))
			for i, leaf := range leaves {
				if leaf.DeviationValue != nil {
					leafValues[i] = leaf.DeviationValue.ToString()
				} else {
					leafValues[i] = leaf.Value.ToString()
				}

			}

			assert.ElementsMatch(t, tt.expectedLeaves, leafValues)
		})
	}
}

func TestFilterBlameTree_FilterByOwner(t *testing.T) {
	tests := []struct {
		name           string
		filter         BlameFilter
		expectedLeaves []string
	}{
		{
			name: "Filter by exact owner",
			filter: BlameFilter{
				Owner: "test-system.intent-emergency",
			},
			expectedLeaves: []string{"ambulance", "digits", "fire-brigade", "police", "name"},
		},
		{
			name: "Filter by wildcard owner",
			filter: BlameFilter{
				Owner: "*config-running*",
			},
			expectedLeaves: []string{"appl-type", "timeout", "host"},
		},
		{
			name: "Filter by partial owner match",
			filter: BlameFilter{
				Owner: "*emergency*",
			},
			expectedLeaves: []string{"ambulance", "digits", "fire-brigade", "police", "name"},
		},
		{
			name: "No matches",
			filter: BlameFilter{
				Owner: "nonexistent-owner",
			},
			expectedLeaves: []string{},
		},
	}

	client := &ConfigClient{}
	tree := createTestBlameTree()
	require.NotNil(t, tree, "Expected sample tree loaded - start test from client pkg folder")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.filterBlameTree(tree, []string{}, tt.filter)

			if len(tt.expectedLeaves) == 0 {
				assert.Nil(t, result, "Expected no results for owner filter: %s", tt.filter.Owner)
				return
			}

			require.NotNil(t, result, "Expected results for owner filter: %s", tt.filter.Owner)

			leaves := tu.CollectLeaves(result)
			leafNames := make([]string, len(leaves))
			for i, leaf := range leaves {
				leafNames[i] = leaf.Name
			}

			assert.ElementsMatch(t, tt.expectedLeaves, leafNames)
		})
	}
}

func TestFilterBlameTree_FilterByPath(t *testing.T) {
	tests := []struct {
		name           string
		filter         BlameFilter
		expectedLeaves []string
	}{
		{
			name: "Filter by exact path",
			filter: BlameFilter{
				Path: "test-device/config/service/emergency/num-list/emergency-list/item/112/ambulance",
			},
			expectedLeaves: []string{"ambulance"},
		},
		{
			name: "Filter by wildcard path",
			filter: BlameFilter{
				Path: "*/emergency/*",
			},
			expectedLeaves: []string{"ambulance", "digits", "fire-brigade", "police", "name"},
		},
		{
			name: "Filter by path endings",
			filter: BlameFilter{
				Path: "*/timeout*",
			},
			expectedLeaves: []string{"timeout"},
		},
		{
			name: "Filter by path beginning",
			filter: BlameFilter{
				Path: "test-device/config/network/*",
			},
			expectedLeaves: []string{"appl-type", "timeout", "host", ""},
		},
	}

	client := &ConfigClient{}
	tree := createTestBlameTree()
	require.NotNil(t, tree, "Expected sample tree loaded - start test from client pkg folder")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.filterBlameTree(tree, []string{}, tt.filter)

			if len(tt.expectedLeaves) == 0 {
				assert.Nil(t, result, "Expected no results for path filter: %s", tt.filter.Path)
				return
			}

			require.NotNil(t, result, "Expected results for path filter: %s", tt.filter.Path)

			leaves := tu.CollectLeaves(result)
			leafNames := make([]string, len(leaves))
			for i, leaf := range leaves {
				leafNames[i] = leaf.Name
			}

			assert.ElementsMatch(t, tt.expectedLeaves, leafNames)
		})
	}
}

func TestFilterBlameTree_CombinedFilters(t *testing.T) {
	tests := []struct {
		name           string
		filter         BlameFilter
		expectedLeaves []string
	}{
		{
			name: "Filter by leaf name and owner",
			filter: BlameFilter{
				LeafName: "*brigade*",
				Owner:    "*emergency*",
			},
			expectedLeaves: []string{"fire-brigade"},
		},
		{
			name: "Filter by all criteria",
			filter: BlameFilter{
				LeafName: "timeout",
				Owner:    "*config-running*",
				Path:     "*/diameter/*",
			},
			expectedLeaves: []string{"timeout"},
		},
		{
			name: "Conflicting filters - no match",
			filter: BlameFilter{
				LeafName: "ambulance",
				Owner:    "*config-running*", // ambulance does not have this owner
			},
			expectedLeaves: []string{},
		},
	}

	client := &ConfigClient{}
	tree := createTestBlameTree()
	require.NotNil(t, tree, "Expected sample tree loaded - start test from client pkg folder")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.filterBlameTree(tree, []string{}, tt.filter)

			if len(tt.expectedLeaves) == 0 {
				assert.Nil(t, result, "Expected no results for combined filter")
				return
			}

			require.NotNil(t, result, "Expected results for combined filter")

			leaves := tu.CollectLeaves(result)
			leafNames := make([]string, len(leaves))
			for i, leaf := range leaves {
				leafNames[i] = leaf.Name
			}

			assert.ElementsMatch(t, tt.expectedLeaves, leafNames)
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	client := &ConfigClient{}

	tests := []struct {
		text     string
		pattern  string
		expected bool
	}{
		{"ambulance", "ambulance", true},
		{"ambulance", "*lance", true},
		{"ambulance", "amb*", true},
		{"ambulance", "*bul*", true},
		{"fire-brigade", "*brigade*", true},
		{"fire-brigade", "*police*", false},
		{"test.example.com", "*.example.*", true},
		{"test.example.com", "*.org", false},
		{"", "", true},
		{"test", "", true}, // Empty Pattern means all
	}

	for _, tt := range tests {
		t.Run(tt.text+"_"+tt.pattern, func(t *testing.T) {
			result := client.matchesPattern(tt.text, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}
