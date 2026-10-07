// Package golang renders a schema type tree as Go structs.
package golang

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

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
	for _, t := range types {
		if needsExplicitNames(t) {
			header.WriteString("import \"encoding/json\"\n\n")
			break
		}
	}

	return header.String() + body.String()
}

func writeStruct(sb *strings.Builder, n *schema.Node) {
	fmt.Fprintf(sb, "type %s struct {\n", n.Name)
	used := map[string]bool{}
	if needsExplicitNames(n) {
		used["MarshalJSON"] = true
		used["UnmarshalJSON"] = true
	}
	names := []string{}
	for _, f := range n.Fields {
		fieldName := schema.UniqueIdentifier(schema.PascalCase(f.JSONName), used)
		names = append(names, fieldName)
		typeStr := typeName(f.Node)
		tag := jsonTag(f.JSONName, f.Node.Nullable)
		if needsExplicitNames(n) {
			tag = "`json:\"-\"`"
		}
		fmt.Fprintf(sb, "\t%s %s %s\n", fieldName, typeStr, tag)
	}
	sb.WriteString("}\n")
	if needsExplicitNames(n) {
		writeJSONMethods(sb, n, names)
	}
}

// encoding/json cannot represent every JSON key in a struct tag: commas,
// an empty name, '-' and certain punctuation/control characters need explicit
// mapping. Keep these properties instead of silently changing wire names.
func needsExplicitNames(n *schema.Node) bool {
	for _, f := range n.Fields {
		if f.JSONName == "" || f.JSONName == "-" || strings.ContainsRune(f.JSONName, ',') {
			return true
		}
		for _, r := range f.JSONName {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("!#$%&()*+-./:;<=>?@[]^_{|}~ ", r) {
				return true
			}
		}
	}
	return false
}

func writeJSONMethods(sb *strings.Builder, n *schema.Node, names []string) {
	fmt.Fprintf(sb, "\nfunc (value %s) MarshalJSON() ([]byte, error) {\n\tfields := map[string]interface{}{}\n", n.Name)
	for i, f := range n.Fields {
		if f.Node.Nullable {
			fmt.Fprintf(sb, "\tif value.%s != nil {\n", names[i])
		}
		fmt.Fprintf(sb, "\tfields[%s] = value.%s\n", strconv.Quote(f.JSONName), names[i])
		if f.Node.Nullable {
			sb.WriteString("\t}\n")
		}
	}
	sb.WriteString("\treturn json.Marshal(fields)\n}\n")
	fmt.Fprintf(sb, "\nfunc (value *%s) UnmarshalJSON(data []byte) error {\n\tvar fields map[string]json.RawMessage\n\tif err := json.Unmarshal(data, &fields); err != nil { return err }\n", n.Name)
	for i, f := range n.Fields {
		fmt.Fprintf(sb, "\tif raw, ok := fields[%s]; ok {\n\t\tif err := json.Unmarshal(raw, &value.%s); err != nil { return err }\n\t}\n", strconv.Quote(f.JSONName), names[i])
	}
	sb.WriteString("\treturn nil\n}\n")
}

func jsonTag(jsonName string, nullable bool) string {
	opts := jsonName
	if nullable {
		opts += ",omitempty"
	}
	tag := "json:" + strconv.Quote(opts)
	if strings.ContainsRune(tag, '`') {
		return strconv.Quote(tag)
	}
	return "`" + tag + "`"
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
