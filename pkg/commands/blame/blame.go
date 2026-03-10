package blame

import (
	"context"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type BlameFilterClient interface {
	GetBlameTree(ctx context.Context, namespace string, device string) (*sdcpb.BlameTreeElement, error)
}

func Run(ctx context.Context, bfc BlameFilterClient, namespace string, device string, pathFilter PathFilters, filter BlameFilters) (*sdcpb.BlameTreeElement, error) {
	tree, err := bfc.GetBlameTree(ctx, namespace, device)
	if err != nil {
		return nil, err
	}

	// return full tree if no filters provided
	if len(filter) == 0 && len(pathFilter) == 0 {
		return tree, nil
	}

	return filterBlameTree(tree, nil, pathFilter, filter), nil
}

// filterBlameTree filter blame tree keeping the whole path
func filterBlameTree(node *sdcpb.BlameTreeElement, path *sdcpb.Path, pathFilter PathFilters, filter BlameFilters) *sdcpb.BlameTreeElement {
	if node == nil {
		return nil
	}

	result := node.Copy()

	// if the node is a leaf (has value), check if it matches the filter
	if node.GetValue() != nil || node.GetDeviationValue() != nil {
		// if the filter does not match, return nil to exclude this leaf
		if !filter.Matches(node) {
			return nil
		}
	}

	// if node has childs, recursive call on each one
	var filteredChilds []*sdcpb.BlameTreeElement

	for _, child := range node.Childs {
		childPath := child.GetPath(path)
		if !pathFilter.Matches(childPath) {
			continue
		}

		filteredChild := filterBlameTree(child, childPath, pathFilter, filter)
		if filteredChild != nil {
			filteredChilds = append(filteredChilds, filteredChild)
		}
	}

	result.Childs = filteredChilds

	// if one child matches, keep the node
	if len(filteredChilds) == 0 && node.GetValue() == nil && node.GetDeviationValue() == nil {
		return nil
	}

	return result
}
