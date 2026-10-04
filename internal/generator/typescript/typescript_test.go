package typescript

import (
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

func TestTypeScriptGeneratesExpectedInterfaces(t *testing.T) {
	code := mustGenerate(t)

	for _, want := range []string{
		"export interface Person {",
		"export interface Address {",
		"export interface Friend {",
		"id: number;",
		"tags: string[];",
		"address: Address;",
		"friends: Friend[];",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected generated code to contain %q\n--- generated ---\n%s", want, code)
		}
	}
}

func TestTypeScriptQuotesInvalidPropertyNames(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, `"full-name": string;`) {
		t.Errorf("expected full-name to be quoted\n--- generated ---\n%s", code)
	}
}

func TestTypeScriptNullableFieldIsOptional(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "nickname?: any;") {
		t.Errorf("expected nickname to be optional any\n--- generated ---\n%s", code)
	}
}

func TestTypeScriptBalancedBraces(t *testing.T) {
	code := mustGenerate(t)
	if strings.Count(code, "{") != strings.Count(code, "}") {
		t.Fatalf("unbalanced braces in generated code:\n%s", code)
	}
}
