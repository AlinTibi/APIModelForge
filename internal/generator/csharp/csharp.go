// Package csharp renders a schema type tree as C# classes using
// System.Text.Json-compatible nullable reference types.
package csharp

import (
	"fmt"
	"strings"

	"APIModelForge/internal/schema"
)

// Generate renders types (in dependency order, as returned by
// schema.CollectTypes) as a single C# source file.
func Generate(types []*schema.Node) string {
	var body strings.Builder
	needsList := false
	needsJSONAttr := false

	for i, t := range types {
		if i > 0 {
			body.WriteString("\n")
		}
		writeClass(&body, t, &needsList, &needsJSONAttr)
	}

	var header strings.Builder
	header.WriteString("#nullable enable\n")
	if needsList || needsJSONAttr {
		header.WriteString("\n")
		if needsList {
			header.WriteString("using System.Collections.Generic;\n")
		}
		if needsJSONAttr {
			header.WriteString("using System.Text.Json.Serialization;\n")
		}
	}
	header.WriteString("\n")

	return header.String() + body.String()
}

func writeClass(sb *strings.Builder, n *schema.Node, needsList, needsJSONAttr *bool) {
	fmt.Fprintf(sb, "public class %s\n{\n", n.Name)
	for _, f := range n.Fields {
		propName := schema.PascalCase(f.JSONName)
		typeStr := typeName(f.Node, needsList)

		if propName != f.JSONName {
			*needsJSONAttr = true
			fmt.Fprintf(sb, "    [JsonPropertyName(%s)]\n", quote(f.JSONName))
		}
		fmt.Fprintf(sb, "    public %s %s { get; set; }\n", typeStr, propName)
	}
	sb.WriteString("}\n")
}

func typeName(n *schema.Node, needsList *bool) string {
	base := baseType(n, needsList)
	if n.Nullable {
		return base + "?"
	}
	return base
}

func baseType(n *schema.Node, needsList *bool) string {
	switch n.Kind {
	case schema.KindBool:
		return "bool"
	case schema.KindInt:
		return "long"
	case schema.KindFloat:
		return "double"
	case schema.KindString:
		return "string"
	case schema.KindObject:
		return n.Name
	case schema.KindArray:
		*needsList = true
		return "List<" + typeName(n.Element, needsList) + ">"
	case schema.KindNull, schema.KindAny:
		return "object"
	default:
		return "object"
	}
}

func quote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
