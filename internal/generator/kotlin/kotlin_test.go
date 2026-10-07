package kotlin

import (
	"strings"
	"testing"

	"APIModelForge/internal/schema"
)

const sampleJSON = `{
	"id": 1,
	"class": "VIP",
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

func TestKotlinGeneratesExpectedDataClasses(t *testing.T) {
	code := mustGenerate(t)

	for _, want := range []string{
		"data class Person(",
		"data class Address(",
		"data class Friend(",
		"val id: Long",
		"val tags: List<String>",
		"val address: Address",
		"val friends: List<Friend>",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected generated code to contain %q\n--- generated ---\n%s", want, code)
		}
	}
}

func TestKotlinEscapesReservedWordWithBackticks(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "val `class`: String") {
		t.Errorf("expected the reserved word 'class' to be backtick-escaped\n--- generated ---\n%s", code)
	}
}

func TestKotlinNullableFieldHasNullDefault(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "val nickname: Any? = null") {
		t.Errorf("expected nickname to be nullable Any with null default\n--- generated ---\n%s", code)
	}
}

func TestKotlinBalancedParens(t *testing.T) {
	code := mustGenerate(t)
	if strings.Count(code, "(") != strings.Count(code, ")") {
		t.Fatalf("unbalanced parens in generated code:\n%s", code)
	}
}

func TestKotlinAvoidsJVMGetterCollisions(t *testing.T) {
	s, err := schema.Infer([]byte(`{"userId":1,"UserId":2,"UserId2":3,"user-id":4,"user_id":5,"isReady":true,"IsReady":false,"is_ready":true,"Is_ready":false}`), "Root")
	if err != nil {
		t.Fatal(err)
	}
	code := Generate(schema.CollectTypes(s.Root))
	for _, want := range []string{"val userId: Long", "val UserId2: Long", "val UserId22: Long", "val user_id: Long", "val user_id2: Long", "val isReady: Boolean", "val IsReady: Boolean", "val is_ready: Boolean", "val Is_ready: Boolean"} {
		if !strings.Contains(code, want) {
			t.Errorf("missing property %q\n%s", want, code)
		}
	}
}
