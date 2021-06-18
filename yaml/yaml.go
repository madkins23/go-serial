package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/madkins23/go-type/reg"
)

//////////////////////////////////////////////////////////////////////////

type Wrapper struct {
	TypeName string
	Contents interface{}
}

func WrapItem(item interface{}) (*Wrapper, error) {
	typeName, err := reg.NameFor(item)
	if typeName, err = reg.NameFor(item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", item, err)
	}

	return &Wrapper{
		TypeName: typeName,
		Contents: item,
	}, nil
}

func UnwrapItem(node *yaml.Node) (interface{}, error) {
	if wrapper, err := NodeAsMap(node); err != nil {
		return nil, fmt.Errorf("get wrapper map: %w", err)
	} else if typeNode, found := wrapper["typename"]; !found {
		return nil, fmt.Errorf("no type name in wrapper")
	} else if typeName, err := NodeAsString(typeNode); err != nil {
		return nil, fmt.Errorf("wrapper type name not string")
	} else if typeName == "" {
		return nil, fmt.Errorf("empty type name in wrapper")
	} else if itemNode, found := wrapper["contents"]; !found {
		return nil, fmt.Errorf("no wrapper contents")
	} else if item, err := reg.Make(typeName); err != nil {
		return nil, fmt.Errorf("make instance of type %s: %w", typeName, err)
	} else if err := itemNode.Decode(item); err != nil {
		return nil, fmt.Errorf("decode item node")
	} else {
		return item, nil
	}
}

//////////////////////////////////////////////////////////////////////////

func NodeAsArray(node *yaml.Node) ([]*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("node not array")
	}

	return node.Content, nil
}

func NodeAsMap(node *yaml.Node) (map[string]*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("node not map")
	} else if len(node.Content)%2 != 0 {
		return nil, fmt.Errorf("odd number of node contents")
	}

	nodeMap := make(map[string]*yaml.Node)
	for i := 0; i < len(node.Content); i += 2 {
		if key, err := NodeAsString(node.Content[i]); err != nil {
			return nil, fmt.Errorf("get key: %w", err)
		} else {
			nodeMap[key] = node.Content[i+1]
		}
	}

	return nodeMap, nil
}

func NodeAsString(node *yaml.Node) (string, error) {
	if node.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("node not scalar")
	}

	return node.Value, nil
}
