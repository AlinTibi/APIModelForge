// Package schema infers a language-agnostic type tree from arbitrary JSON
// values. Every output generator (C#, TypeScript, Kotlin, Python, Go)
// consumes the same tree, so inference and naming rules live here exactly
// once.
package schema

// Kind identifies the shape of a value observed in the source JSON.
type Kind int

const (
	KindNull Kind = iota
	KindBool
	KindInt
	KindFloat
	KindString
	KindObject
	KindArray
	// KindAny represents a value with no usable type information (an
	// always-empty array) or a position where incompatible primitive
	// kinds were observed across samples (e.g. a field that is
	// sometimes a string and sometimes a number). Every target language
	// maps this to its closest "any" equivalent.
	KindAny
)

func (k Kind) String() string {
	switch k {
	case KindNull:
		return "null"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindString:
		return "string"
	case KindObject:
		return "object"
	case KindArray:
		return "array"
	case KindAny:
		return "any"
	default:
		return "unknown"
	}
}

// Node is one position in the inferred type tree: a scalar, an object, or
// an array. Object and array-of-object nodes that end up representing a
// named type get a Name assigned by AssignNames.
type Node struct {
	Kind Kind

	// Nullable is true when this position was observed as JSON null at
	// least once, or (for object fields) was missing from at least one
	// sampled object that otherwise matched the same field set.
	Nullable bool

	// Fields holds an object node's properties in first-seen order.
	Fields []*Field

	// Element holds an array node's (merged) element type.
	Element *Node

	// Name is the generated type name for Object nodes (e.g. "Root",
	// "Address", "Item"). Empty for scalar/array nodes.
	Name string
}

// Field is a single named property of an object Node.
type Field struct {
	JSONName string
	Node     *Node
}

// Schema is the result of inferring and naming a complete JSON document.
type Schema struct {
	Root *Node
}
