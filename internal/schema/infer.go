package schema

import (
	"encoding/json"
	"errors"
	"strings"
)

// Infer parses raw JSON and builds a named, merged type tree for it.
// rootName becomes the generated name of the root object type (or the
// element type, when the root is an array of objects).
func Infer(raw []byte, rootName string) (*Schema, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return nil, errors.New("JSON input is empty")
	}

	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}

	root := inferValue(value)

	if err := validateRootKind(root); err != nil {
		return nil, err
	}

	if strings.TrimSpace(rootName) == "" {
		rootName = "Root"
	}

	assignNames(root, rootName, make(map[string]*Node))

	return &Schema{Root: root}, nil
}

func validateRootKind(root *Node) error {
	switch root.Kind {
	case KindObject:
		return nil
	case KindArray:
		if root.Element != nil && root.Element.Kind == KindObject {
			return nil
		}
		return errors.New("root JSON array must contain objects to generate models from")
	default:
		return errors.New("root JSON value must be an object or an array of objects")
	}
}

func inferValue(v interface{}) *Node {
	switch t := v.(type) {
	case nil:
		return &Node{Kind: KindNull, Nullable: true}
	case bool:
		return &Node{Kind: KindBool}
	case json.Number:
		if isInteger(t) {
			return &Node{Kind: KindInt}
		}
		return &Node{Kind: KindFloat}
	case string:
		return &Node{Kind: KindString}
	case []interface{}:
		return inferArray(t)
	case *orderedObject:
		return inferObject(t)
	default:
		return &Node{Kind: KindAny}
	}
}

func inferArray(items []interface{}) *Node {
	var element *Node
	for _, item := range items {
		element = Merge(element, inferValue(item))
	}
	if element == nil {
		element = &Node{Kind: KindAny}
	}
	return &Node{Kind: KindArray, Element: element}
}

func inferObject(obj *orderedObject) *Node {
	node := &Node{Kind: KindObject}
	for _, key := range obj.keys {
		node.Fields = append(node.Fields, &Field{
			JSONName: key,
			Node:     inferValue(obj.values[key]),
		})
	}
	return node
}

// isInteger reports whether a JSON number literal has no fractional or
// exponent part, e.g. "42" or "-7" but not "3.14" or "1e9".
func isInteger(n json.Number) bool {
	s := string(n)
	return !strings.ContainsAny(s, ".eE")
}
