// Package golang renders a schema type tree as Go structs.
package golang

import (
	"fmt"
	"strings"

	"APIModelForge/internal/schema"
)

// Generate renders types (in dependency order) as a single Go source file
// in package "models".
func Generate(types []*schema.Node) string {
	var body strings.Builder
	for i, t := range types {
		if i > 0 {
			body.WriteString("\n")
		}
		writeStruct(&body, t)
	}

	var header strings.Builder
	header.WriteString("package models\n\n")

	return header.String() + body.String()
}

func writeStruct(sb *strings.Builder, n *schema.Node) {
	fmt.Fprintf(sb, "type %s struct {\n", n.Name)
	for _, f := range n.Fields {
		fieldName := schema.PascalCase(f.JSONName)
		typeStr := typeName(f.Node)
		tag := jsonTag(f.JSONName, f.Node.Nullable)
		fmt.Fprintf(sb, "\t%s %s %s\n", fieldName, typeStr, tag)
	}
	sb.WriteString("}\n")
}

func jsonTag(jsonName string, nullable bool) string {
	opts := jsonName
	if nullable {
		opts += ",omitempty"
	}
	return "`json:\"" + opts + "\"`"
}

func typeName(n *schema.Node) string {
	switch n.Kind {
	case schema.KindBool:
		return pointerIfNullable(n, "bool")
	case schema.KindInt:
		return pointerIfNullable(n, "int64")
	case schema.KindFloat:
		return pointerIfNullable(n, "float64")
	case schema.KindString:
		return pointerIfNullable(n, "string")
	case schema.KindObject:
		return pointerIfNullable(n, n.Name)
	case schema.KindArray:
		return "[]" + typeName(n.Element)
	case schema.KindNull, schema.KindAny:
		// interface{} is already nil-able; no pointer needed.
		return "interface{}"
	default:
		return "interface{}"
	}
}

func pointerIfNullable(n *schema.Node, base string) string {
	if n.Nullable {
		return "*" + base
	}
	return base
}
