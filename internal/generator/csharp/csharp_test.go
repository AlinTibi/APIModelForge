package csharp

import (
	"strings"
	"testing"

	"APIModelForge/internal/schema"
)

const sampleJSON = `{
	"id": 1,
	"full-name": "Ada Lovelace",
	"nickname": null,
	"score": 3.5,
	"tags": ["math", "logic"],
	"address": {"city": "London", "zip": "W1"},
	"friends": [{"id": 2, "name": "Grace"}]
}`

func mustGenerate(t *testing.T) string {
	t.Helper()
	s, err := schema.Infer([]byte(sampleJSON), "Person")
	if err != nil {
		t.Fatalf("schema.Infer failed: %v", err)
	}
	types := schema.CollectTypes(s.Root)
	return Generate(types)
}

func TestCSharpGeneratesExpectedClasses(t *testing.T) {
	code := mustGenerate(t)

	for _, want := range []string{
		"public class Person",
		"public class Address",
		"public class Friend",
		"public long Id { get; set; }",
		"public double Score { get; set; }",
		"public List<string> Tags { get; set; }",
		"public Address Address { get; set; }",
		"public List<Friend> Friends { get; set; }",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected generated code to contain %q\n--- generated ---\n%s", want, code)
		}
	}
}

func TestCSharpUsesJsonPropertyNameWhenNameDiffers(t *testing.T) {
	code := mustGenerate(t)

	if !strings.Contains(code, `[JsonPropertyName("full-name")]`) {
		t.Errorf("expected a JsonPropertyName attribute for full-name\n--- generated ---\n%s", code)
	}
	if !strings.Contains(code, "public string FullName { get; set; }") {
		t.Errorf("expected FullName property\n--- generated ---\n%s", code)
	}
	if !strings.Contains(code, "using System.Text.Json.Serialization;") {
		t.Errorf("expected the JsonPropertyName using directive to be present")
	}
}

func TestCSharpNullableAlwaysNullFieldIsNullableObject(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "public object? Nickname { get; set; }") {
		t.Errorf("expected nickname to render as object?\n--- generated ---\n%s", code)
	}
}

func TestCSharpBalancedBraces(t *testing.T) {
	code := mustGenerate(t)
	if strings.Count(code, "{") != strings.Count(code, "}") {
		t.Fatalf("unbalanced braces in generated code:\n%s", code)
	}
}
