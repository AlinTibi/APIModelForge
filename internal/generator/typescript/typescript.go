// Package typescript renders a schema type tree as TypeScript interfaces.
package typescript

import (
	"fmt"
	"strings"

	"APIModelForge/internal/schema"
)

// Generate renders types (in dependency order) as a single TypeScript
// source file.
func Generate(types []*schema.Node) string {
	var sb strings.Builder
	for i, t := range types {
		if i > 0 {
			sb.WriteString("\n")
		}
		writeInterface(&sb, t)
	}
	return sb.String()
}

func writeInterface(sb *strings.Builder, n *schema.Node) {
	fmt.Fprintf(sb, "export interface %s {\n", n.Name)
	for _, f := range n.Fields {
		name := propertyName(f.JSONName)
		optional := ""
		if f.Node.Nullable {
			optional = "?"
		}
		fmt.Fprintf(sb, "  %s%s: %s;\n", name, optional, typeName(f.Node))
	}
	sb.WriteString("}\n")
}

func propertyName(jsonName string) string {
	if schema.IsValidBareIdentifier(jsonName) {
		return jsonName
	}
	return quote(jsonName)
}

func typeName(n *schema.Node) string {
	switch n.Kind {
	case schema.KindBool:
		return "boolean"
	case schema.KindInt, schema.KindFloat:
		return "number"
	case schema.KindString:
		return "string"
	case schema.KindObject:
		return n.Name
	case schema.KindArray:
		elem := typeName(n.Element)
		if strings.Contains(elem, " ") {
			return "(" + elem + ")[]"
		}
		return elem + "[]"
	case schema.KindNull, schema.KindAny:
		return "any"
	default:
		return "any"
	}
}

func quote(s string) string {
	return schema.QuoteString(s)
}
