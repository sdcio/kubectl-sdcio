package blame

import (
	"testing"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

func TestOwnerFilter(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		nodeOwner string
		expected  bool
	}{
		{
			name:      "exact match",
			owner:     "alice",
			nodeOwner: "alice",
			expected:  true,
		},
		{
			name:      "no match",
			owner:     "alice",
			nodeOwner: "bob",
			expected:  false,
		},
		{
			name:      "wildcard match",
			owner:     "ali*",
			nodeOwner: "alice",
			expected:  true,
		},
		{
			name:      "case insensitive fallback",
			owner:     "Alice",
			nodeOwner: "alice",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := OwnerFilter(tt.owner)
			node := &sdcpb.BlameTreeElement{
				Owner: tt.nodeOwner,
			}
			if got := filter(node); got != tt.expected {
				t.Errorf("OwnerFilter(%q) = %v, want %v", tt.owner, got, tt.expected)
			}
		})
	}
}

func TestLeafNameFilter(t *testing.T) {
	tests := []struct {
		name     string
		leafName string
		nodeName string
		expected bool
	}{
		{
			name:     "exact match",
			leafName: "config",
			nodeName: "config",
			expected: true,
		},
		{
			name:     "no match",
			leafName: "config",
			nodeName: "state",
			expected: false,
		},
		{
			name:     "wildcard match",
			leafName: "conf*",
			nodeName: "config",
			expected: true,
		},
		{
			name:     "wildcard start match",
			leafName: "*fig",
			nodeName: "config",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := LeafNameFilter(tt.leafName)
			node := &sdcpb.BlameTreeElement{
				Name: tt.nodeName,
			}
			if got := filter(node); got != tt.expected {
				t.Errorf("LeafNameFilter(%q) = %v, want %v", tt.leafName, got, tt.expected)
			}
		})
	}
}

func TestDeviationFilter(t *testing.T) {
	tests := []struct {
		name          string
		value         *sdcpb.TypedValue
		deviatedValue *sdcpb.TypedValue
		expected      bool
	}{
		{
			name:          "deviated node",
			value:         &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "foo"}},
			deviatedValue: &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "bar"}},
			expected:      true,
		},
		{
			name:     "non-deviated node",
			value:    &sdcpb.TypedValue{Value: &sdcpb.TypedValue_StringVal{StringVal: "foo"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := DeviationFilter()
			node := &sdcpb.BlameTreeElement{
				Value:          tt.value,
				DeviationValue: tt.deviatedValue,
			}
			if got := filter(node); got != tt.expected {
				t.Errorf("DeviationFilter() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBlameFiltersMatches(t *testing.T) {
	tests := []struct {
		name     string
		filters  BlameFilters
		owner    string
		leafName string
		expected bool
	}{
		{
			name: "all filters match",
			filters: BlameFilters{
				OwnerFilter("alice"),
				LeafNameFilter("config"),
			},
			owner:    "alice",
			leafName: "config",
			expected: true,
		},
		{
			name: "first filter fails",
			filters: BlameFilters{
				OwnerFilter("alice"),
				LeafNameFilter("config"),
			},
			owner:    "bob",
			leafName: "config",
			expected: false,
		},
		{
			name: "second filter fails",
			filters: BlameFilters{
				OwnerFilter("alice"),
				LeafNameFilter("config"),
			},
			owner:    "alice",
			leafName: "state",
			expected: false,
		},
		{
			name:     "empty filters match",
			filters:  BlameFilters{},
			owner:    "anyone",
			leafName: "anything",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &sdcpb.BlameTreeElement{
				Owner: tt.owner,
				Name:  tt.leafName,
			}
			if got := tt.filters.Matches(node); got != tt.expected {
				t.Errorf("BlameFilters.Matches() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPathFilterFunc(t *testing.T) {
	tests := []struct {
		name       string
		filterPath *sdcpb.Path
		path       *sdcpb.Path
		expected   bool
	}{
		{
			name: "parent path matches child path",
			filterPath: &sdcpb.Path{Elem: []*sdcpb.PathElem{
				{Name: "interface"},
				{Name: "ethernet", Key: map[string]string{"name": "ethernet-1/1"}},
			}},
			path: &sdcpb.Path{Elem: []*sdcpb.PathElem{
				{Name: "interface"},
				{Name: "ethernet", Key: map[string]string{"name": "ethernet-1/1"}},
				{Name: "description"},
			}},
			expected: true,
		},
		{
			name: "different key does not match",
			filterPath: &sdcpb.Path{Elem: []*sdcpb.PathElem{
				{Name: "interface"},
				{Name: "ethernet", Key: map[string]string{"name": "ethernet-1/1"}},
			}},
			path: &sdcpb.Path{Elem: []*sdcpb.PathElem{
				{Name: "interface"},
				{Name: "ethernet", Key: map[string]string{"name": "ethernet-1/2"}},
				{Name: "description"},
			}},
			expected: false,
		},
		{
			name: "origin mismatch does not match",
			filterPath: &sdcpb.Path{
				Origin: "a",
				Elem: []*sdcpb.PathElem{
					{Name: "interface"},
				},
			},
			path: &sdcpb.Path{
				Origin: "b",
				Elem: []*sdcpb.PathElem{
					{Name: "interface"},
					{Name: "description"},
				},
			},
			expected: false,
		},
		{
			name: "nil child path does not match",
			filterPath: &sdcpb.Path{Elem: []*sdcpb.PathElem{
				{Name: "interface"},
			}},
			path:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := PathFilterFunc(tt.filterPath)
			if got := filter(tt.path); got != tt.expected {
				t.Errorf("PathFilterFunc() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPathFiltersMatches(t *testing.T) {
	path := &sdcpb.Path{Elem: []*sdcpb.PathElem{
		{Name: "interface"},
		{Name: "ethernet", Key: map[string]string{"name": "ethernet-1/1"}},
		{Name: "description"},
	}}

	matchingParent := PathFilterFunc(&sdcpb.Path{Elem: []*sdcpb.PathElem{
		{Name: "interface"},
	}})
	nonMatchingParent := PathFilterFunc(&sdcpb.Path{Elem: []*sdcpb.PathElem{
		{Name: "system"},
	}})

	tests := []struct {
		name     string
		filters  PathFilters
		expected bool
	}{
		{
			name: "single matching filter returns true",
			filters: PathFilters{
				matchingParent,
			},
			expected: true,
		},
		{
			name: "one filter matches and one fails returns true",
			filters: PathFilters{
				matchingParent,
				nonMatchingParent,
			},
			expected: true,
		},
		{
			name: "all filters fail returns false",
			filters: PathFilters{
				nonMatchingParent,
			},
			expected: false,
		},
		{
			name:     "empty filters match",
			filters:  PathFilters{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filters.Matches(path); got != tt.expected {
				t.Errorf("PathFilters.Matches() = %v, want %v", got, tt.expected)
			}
		})
	}
}
