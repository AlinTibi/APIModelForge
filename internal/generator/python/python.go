// Package python renders a schema type tree as Python dataclasses.
package python

import (
	"fmt"
	"sort"
	"strings"

	"APIModelForge/internal/schema"
)

var reservedWords = map[string]bool{
	"False": true, "None": true, "True": true, "and": true, "as": true, "assert": true,
	"async": true, "await": true, "break": true, "class": true, "continue": true, "def": true,
	"del": true, "elif": true, "else": true, "except": true, "finally": true, "for": true,
	"from": true, "global": true, "if": true, "import": true, "in": true, "is": true,
	"lambda": true, "nonlocal": true, "not": true, "or": true, "pass": true, "raise": true,
	"return": true, "try": true, "while": true, "with": true, "yield": true,
}

// Generate renders types (in dependency order) as a single Python source
// file. `from __future__ import annotations` is always emitted so nested
// dataclasses can reference each other regardless of declaration order.
func Generate(types []*schema.Node) string {
	var body strings.Builder
	imports := make(map[string]bool)

	for i, t := range types {
		if i > 0 {
			body.WriteString("\n\n")
		}
		writeDataclass(&body, t, imports)
	}

	var header strings.Builder
	header.WriteString("from __future__ import annotations\n\n")
	header.WriteString("from dataclasses import dataclass\n")
	if len(imports) > 0 {
		names := make([]string, 0, len(imports))
		for name := range imports {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintf(&header, "from typing import %s\n", strings.Join(names, ", "))
	}
	header.WriteString("\n\n")

	return header.String() + body.String()
}

func writeDataclass(sb *strings.Builder, n *schema.Node, imports map[string]bool) {
	sb.WriteString("@dataclass(kw_only=True)\n")
	fmt.Fprintf(sb, "class %s:\n", n.Name)
	if len(n.Fields) == 0 {
		sb.WriteString("    pass\n")
		return
	}
	used := map[string]bool{"__annotations__": true, "__dict__": true, "__weakref__": true, "__slots__": true, "__class__": true, "__dataclass_fields__": true, "__dataclass_params__": true, "__init__": true, "__new__": true, "__repr__": true, "__eq__": true, "__hash__": true, "__module__": true, "__qualname__": true, "__match_args__": true}
	for _, f := range n.Fields {
		name, renamed := propertyName(f.JSONName)
		name = schema.UniqueIdentifier(name, used)
		renamed = renamed || name != f.JSONName
		typeStr := typeName(f.Node, imports)
		if f.Node.Nullable {
			imports["Optional"] = true
			typeStr = "Optional[" + typeStr + "]"
			fmt.Fprintf(sb, "    %s: %s = None", name, typeStr)
		} else {
			fmt.Fprintf(sb, "    %s: %s", name, typeStr)
		}
		if renamed {
			fmt.Fprintf(sb, "  # json: %s", quote(f.JSONName))
		}
		sb.WriteString("\n")
	}
}

// propertyName returns a valid Python identifier for jsonName, and whether
// it had to be changed from the original JSON key.
func propertyName(jsonName string) (string, bool) {
	candidate := schema.SanitizeIdentifier(jsonName)
	// Dunder fields can replace Python/dataclass hooks even when they are
	// syntactically valid identifiers (for example __post_init__).
	if strings.HasPrefix(candidate, "__") && strings.HasSuffix(candidate, "__") {
		candidate = "field" + candidate
	}
	if reservedWords[candidate] {
		candidate += "_"
	}
	return candidate, candidate != jsonName
}

func typeName(n *schema.Node, imports map[string]bool) string {
	switch n.Kind {
	case schema.KindBool:
		return "bool"
	case schema.KindInt:
		return "int"
	case schema.KindFloat:
		return "float"
	case schema.KindString:
		return "str"
	case schema.KindObject:
		return n.Name
	case schema.KindArray:
		return "list[" + typeName(n.Element, imports) + "]"
	case schema.KindNull, schema.KindAny:
		imports["Any"] = true
		return "Any"
	default:
		imports["Any"] = true
		return "Any"
	}
}

func quote(s string) string {
	return schema.QuoteString(s)
}
