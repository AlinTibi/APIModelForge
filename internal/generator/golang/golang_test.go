package golang

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"APIModelForge/internal/schema"
)

const sampleJSON = `{
	"id": 1,
	"full-name": "Ada Lovelace",
	"nickname": null,
	"tags": ["math", "logic"],
	"address": {"city": "London"},
	"friends": [{"id": 2}]
}`

func mustGenerate(t *testing.T) string {
	t.Helper()
	s, err := schema.Infer([]byte(sampleJSON), "Person")
	if err != nil {
		t.Fatalf("schema.Infer failed: %v", err)
	}
	return Generate(schema.CollectTypes(s.Root))
}

func TestGoGeneratesExpectedStructs(t *testing.T) {
	code := mustGenerate(t)

	for _, want := range []string{
		"package models",
		"type Person struct {",
		"type Address struct {",
		"type Friend struct {",
		"Id int64 `json:\"id\"`",
		"Tags []string `json:\"tags\"`",
		"Address Address `json:\"address\"`",
		"Friends []Friend `json:\"friends\"`",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected generated code to contain %q\n--- generated ---\n%s", want, code)
		}
	}
}

func TestGoFieldNamesAreExportedWithJSONTag(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "FullName string `json:\"full-name\"`") {
		t.Errorf("expected FullName field with original-name json tag\n--- generated ---\n%s", code)
	}
}

func TestGoNullableFieldRendersAsInterfaceWithOmitempty(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "Nickname interface{} `json:\"nickname,omitempty\"`") {
		t.Errorf("expected nickname to render as interface{} (always-null field)\n--- generated ---\n%s", code)
	}
}

func TestGoNullableScalarFieldIsPointer(t *testing.T) {
	s, err := schema.Infer([]byte(`[{"age": 30}, {"name": "no age here"}]`), "Item")
	if err != nil {
		t.Fatalf("schema.Infer failed: %v", err)
	}
	code := Generate(schema.CollectTypes(s.Root))

	if !strings.Contains(code, "Age *int64 `json:\"age,omitempty\"`") {
		t.Errorf("expected a sometimes-missing int field to render as *int64\n--- generated ---\n%s", code)
	}
}

// TestGoSyntaxIsValid parses the generated source with go/parser to prove
// it is syntactically valid Go, using only the standard library.
func TestGoSyntaxIsValid(t *testing.T) {
	code := mustGenerate(t)
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated_models.go", code, parser.AllErrors); err != nil {
		t.Fatalf("generated Go failed to parse: %v\n--- generated ---\n%s", err, code)
	}
}
