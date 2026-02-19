package testutils

import (
	"fmt"
	"os"
	"strconv"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"gopkg.in/yaml.v2"
)

// ConfigBlame is the YAML struct
type ConfigBlame struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Status struct {
		Value *BlameTreeElementYAML `yaml:"value"`
	} `yaml:"status"`
}

// BlameTreeElementYAML is a YAML tree element
type BlameTreeElementYAML struct {
	Name           string                  `yaml:"name,omitempty"`
	Owner          string                  `yaml:"owner,omitempty"`
	Value          *ValueYAML              `yaml:"value,omitempty"`
	DeviationValue *ValueYAML              `yaml:"deviationValue,omitempty"`
	Childs         []*BlameTreeElementYAML `yaml:"childs,omitempty"`
}

// ValueYAML holds multiple value type
type ValueYAML struct {
	StringVal string `yaml:"stringVal,omitempty"`
	IntVal    string `yaml:"intVal,omitempty"`
	BoolVal   *bool  `yaml:"boolVal,omitempty"`
}

// LoadBlameTreeFromFile loads a blame tree from a YAML file
func LoadBlameTreeFromFile(filename string) (*BlameTreeElementYAML, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error while reading file: %v", err)
	}

	var configBlame ConfigBlame
	err = yaml.Unmarshal(data, &configBlame)
	if err != nil {
		return nil, fmt.Errorf("error while parsing YAML: %v", err)
	}

	return configBlame.Status.Value, nil
}

// ConvertValueYAMLToProtobuf converts ValueYAML to protobuf TypedValue
func ConvertValueYAMLToProtobuf(v *ValueYAML) *sdcpb.TypedValue {
	if v == nil {
		return nil
	}

	typedValue := &sdcpb.TypedValue{}

	switch {
	case v.StringVal != "":
		typedValue.Value = &sdcpb.TypedValue_StringVal{
			StringVal: v.StringVal,
		}
	case v.IntVal != "":
		// Convert to int64 if possible
		if intVal, err := strconv.ParseInt(v.IntVal, 10, 64); err == nil {
			typedValue.Value = &sdcpb.TypedValue_IntVal{
				IntVal: intVal,
			}
		} else {
			// Keep as string if conversion fails
			typedValue.Value = &sdcpb.TypedValue_StringVal{
				StringVal: v.IntVal,
			}
		}
	case v.BoolVal != nil:
		typedValue.Value = &sdcpb.TypedValue_BoolVal{
			BoolVal: *v.BoolVal,
		}
	}

	return typedValue
}

// ToProtobuf converts BlameTreeElementYAML to protobuf BlameTreeElement
func (b *BlameTreeElementYAML) ToProtobuf() *sdcpb.BlameTreeElement {
	if b == nil {
		return nil
	}

	node := &sdcpb.BlameTreeElement{
		Name:  b.Name,
		Owner: b.Owner,
	}

	if b.Value != nil {
		node.Value = ConvertValueYAMLToProtobuf(b.Value)
	}

	if b.DeviationValue != nil {
		node.DeviationValue = ConvertValueYAMLToProtobuf(b.DeviationValue)
	}

	if len(b.Childs) > 0 {
		node.Childs = make([]*sdcpb.BlameTreeElement, len(b.Childs))
		for i, child := range b.Childs {
			node.Childs[i] = child.ToProtobuf()
		}
	}

	return node
}

// CollectLeaves gathers all leaf nodes from a blame tree
func CollectLeaves(node *sdcpb.BlameTreeElement) []*sdcpb.BlameTreeElement {
	if node == nil {
		return nil
	}

	var leaves []*sdcpb.BlameTreeElement

	if len(node.Childs) == 0 {
		leaves = append(leaves, node)
	} else {
		for _, child := range node.Childs {
			leaves = append(leaves, CollectLeaves(child)...)
		}
	}

	return leaves
}

// CreateTestBlameTree creates a test blame tree from the standard test file
func CreateTestBlameTree(testFilePath string) *sdcpb.BlameTreeElement {
	tree, err := LoadBlameTreeFromFile(testFilePath)
	if err != nil {
		return nil
	}
	return tree.ToProtobuf()
}
