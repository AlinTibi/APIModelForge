// Package kotlin renders a schema type tree as Kotlin data classes.
package kotlin

import (
	"fmt"
	"strings"

	"APIModelForge/internal/schema"
)

var reservedWords = map[string]bool{
	"class": true, "object": true, "interface": true, "fun": true, "val": true, "var": true,
	"is": true, "in": true, "as": true, "when": true, "if": true, "else": true, "for": true,
	"while": true, "do": true, "break": true, "continue": true, "return": true, "throw": true,
	"try": true, "catch": true, "finally": true, "null": true, "true": true, "false": true,
	"this": true, "super": true, "typeof": true, "package": true, "import": true,
	"typealias": true, "constructor": true, "init": true, "companion": true, "out": true, "by": true,
}

// Generate renders types (in dependency order) as a single Kotlin source
// file.
func Generate(types []*schema.Node) string {
	var sb strings.Builder
	for i, t := range types {
		if i > 0 {
			sb.WriteString("\n")
		}
		writeDataClass(&sb, t)
	}
	return sb.String()
}

func writeDataClass(sb *strings.Builder, n *schema.Node) {
	fmt.Fprintf(sb, "data class %s(\n", n.Name)
	for i, f := range n.Fields {
		name := propertyName(f.JSONName)
		typeStr := typeName(f.Node)
		line := fmt.Sprintf("    val %s: %s", name, typeStr)
		if f.Node.Nullable {
			line += " = null"
		}
		if i < len(n.Fields)-1 {
			line += ","
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString(")\n")
}

func propertyName(jsonName string) string {
	if schema.IsValidBareIdentifier(jsonName) && !reservedWords[jsonName] {
		return jsonName
	}
	return "`" + strings.ReplaceAll(jsonName, "`", "") + "`"
}

func typeName(n *schema.Node) string {
	base := baseType(n)
	if n.Nullable {
		return base + "?"
	}
	return base
}

func baseType(n *schema.Node) string {
	switch n.Kind {
	case schema.KindBool:
		return "Boolean"
	case schema.KindInt:
		return "Long"
	case schema.KindFloat:
		return "Double"
	case schema.KindString:
		return "String"
	case schema.KindObject:
		return n.Name
	case schema.KindArray:
		return "List<" + typeName(n.Element) + ">"
	case schema.KindNull, schema.KindAny:
		return "Any"
	default:
		return "Any"
	}
}
