package blame

import (
	"strings"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

// BlameFilter defines a function type for filtering BlameTreeElements
type BlameFilter func(*sdcpb.BlameTreeElement) bool

// BlameFilters is a slice of BlameFilter functions
type BlameFilters []BlameFilter

// Matches returns true if the node matches all filters
func (b BlameFilters) Matches(node *sdcpb.BlameTreeElement) bool {
	for _, filter := range b {
		if !filter(node) {
			return false
		}
	}
	return true
}

// PathFilter defines a function type for filtering paths
type PathFilter func(*sdcpb.Path) bool

// PathFilters is a slice of PathFilter functions
type PathFilters []PathFilter

// Matches returns true if the path matches any filter
func (p PathFilters) Matches(path *sdcpb.Path) bool {
	if len(p) == 0 {
		return true
	}

	for _, filter := range p {
		if filter(path) {
			return true
		}
	}

	return false
}

// OwnerFilter returns a filter that matches nodes with the specified owner
func OwnerFilter(owner string) BlameFilter {
	return func(node *sdcpb.BlameTreeElement) bool {
		return matchesWildcard(node.Owner, owner)
	}
}

// LeafNameFilter returns a filter that matches nodes with the specified leaf name
func LeafNameFilter(leafName string) BlameFilter {
	return func(node *sdcpb.BlameTreeElement) bool {
		return matchesWildcard(node.Name, leafName)
	}
}

// PathFilterFunc returns a filter that matches paths with the specified parent path
func PathFilterFunc(filterPath *sdcpb.Path) PathFilter {
	return func(path *sdcpb.Path) bool {
		return filterPath.SharesPrefix(path)
	}
}

// DeviationFilter returns a filter that matches nodes that are deviated
func DeviationFilter() BlameFilter {
	return func(node *sdcpb.BlameTreeElement) bool {
		return node.IsDeviated()
	}
}

// OrFilter combines multiple filters with a logical OR
func OrFilter(filters ...BlameFilter) BlameFilter {
	return func(node *sdcpb.BlameTreeElement) bool {
		for _, filter := range filters {
			if filter(node) {
				return true
			}
		}
		return false
	}
}

// BuildFilters constructs a BlameFilters slice from raw filter values
// Multiple filters of the same type are combined with OR, different types are ANDed
func BuildFilters(leafNames, owners []string, filterDeviation bool) BlameFilters {
	var filters BlameFilters

	appendOrFilters(&filters, leafNames, LeafNameFilter)
	appendOrFilters(&filters, owners, OwnerFilter)

	if filterDeviation {
		filters = append(filters, DeviationFilter())
	}

	return filters
}

// appendOrFilters appends filters to the slice, using OrFilter if multiple items are provided
func appendOrFilters(filters *BlameFilters, items []string, filterFunc func(string) BlameFilter) {
	var itemFilters []BlameFilter
	for _, item := range items {
		itemFilters = append(itemFilters, filterFunc(item))
	}
	if len(itemFilters) > 1 {
		*filters = append(*filters, OrFilter(itemFilters...))
	} else if len(itemFilters) == 1 {
		*filters = append(*filters, itemFilters[0])
	}
}

// BuildPathFilters constructs a PathFilters slice from raw filter path strings
func BuildPathFilters(filterPaths []string) (PathFilters, error) {
	var pathFilters PathFilters
	for _, filterPath := range filterPaths {
		path, err := sdcpb.ParsePath(filterPath)
		if err != nil {
			return nil, err
		}
		pathFilters = append(pathFilters, PathFilterFunc(path))
	}
	return pathFilters, nil
}

// matchesWildcard performs case-insensitive wildcard matching
// Supports * as wildcard: "admin*" matches anything starting with admin
func matchesWildcard(text, pattern string) bool {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	// No wildcard - exact match
	if !strings.Contains(pattern, "*") {
		return text == pattern
	}

	// Handle wildcards
	parts := strings.Split(pattern, "*")

	// Check prefix
	if !strings.HasPrefix(text, parts[0]) {
		return false
	}
	text = text[len(parts[0]):]

	// Check suffix
	if len(parts) > 1 && parts[len(parts)-1] != "" {
		if !strings.HasSuffix(text, parts[len(parts)-1]) {
			return false
		}
		text = text[:len(text)-len(parts[len(parts)-1])]
	}

	// Check middle parts are in order
	for i := 1; i < len(parts)-1; i++ {
		if parts[i] == "" {
			continue
		}
		idx := strings.Index(text, parts[i])
		if idx == -1 {
			return false
		}
		text = text[idx+len(parts[i]):]
	}

	return true
}
