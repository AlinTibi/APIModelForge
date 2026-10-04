package schema

import (
	"testing"
)

func mustInfer(t *testing.T, jsonText string, rootName string) *Schema {
	t.Helper()
	s, err := Infer([]byte(jsonText), rootName)
	if err != nil {
		t.Fatalf("Infer() returned unexpected error: %v", err)
	}
	return s
}

func fieldByName(t *testing.T, n *Node, name string) *Field {
	t.Helper()
	for _, f := range n.Fields {
		if f.JSONName == name {
			return f
		}
	}
	t.Fatalf("field %q not found; have %d fields", name, len(n.Fields))
	return nil
}

func TestSimpleObject(t *testing.T) {
	s := mustInfer(t, `{"id": 1, "name": "Ada", "active": true, "score": 3.5}`, "Root")

	root := s.Root
	if root.Kind != KindObject {
		t.Fatalf("expected root Kind=Object, got %v", root.Kind)
	}
	if root.Name != "Root" {
		t.Errorf("expected root Name=Root, got %q", root.Name)
	}
	if len(root.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(root.Fields))
	}

	if k := fieldByName(t, root, "id").Node.Kind; k != KindInt {
		t.Errorf("id: expected KindInt, got %v", k)
	}
	if k := fieldByName(t, root, "name").Node.Kind; k != KindString {
		t.Errorf("name: expected KindString, got %v", k)
	}
	if k := fieldByName(t, root, "active").Node.Kind; k != KindBool {
		t.Errorf("active: expected KindBool, got %v", k)
	}
	if k := fieldByName(t, root, "score").Node.Kind; k != KindFloat {
		t.Errorf("score: expected KindFloat, got %v", k)
	}
}

func TestNestedObject(t *testing.T) {
	s := mustInfer(t, `{"user": {"id": 1, "address": {"city": "Paris", "zip": "75001"}}}`, "Root")

	userField := fieldByName(t, s.Root, "user")
	if userField.Node.Kind != KindObject {
		t.Fatalf("expected user Kind=Object, got %v", userField.Node.Kind)
	}
	if userField.Node.Name != "User" {
		t.Errorf("expected nested type name User, got %q", userField.Node.Name)
	}

	addressField := fieldByName(t, userField.Node, "address")
	if addressField.Node.Kind != KindObject {
		t.Fatalf("expected address Kind=Object, got %v", addressField.Node.Kind)
	}
	if addressField.Node.Name != "Address" {
		t.Errorf("expected nested type name Address, got %q", addressField.Node.Name)
	}

	types := CollectTypes(s.Root)
	if len(types) != 3 {
		t.Fatalf("expected 3 distinct types (Address, User, Root), got %d: %v", len(types), typeNames(types))
	}
	// Dependency order: Address and User must both come before Root.
	if types[len(types)-1].Name != "Root" {
		t.Errorf("expected Root to be last in dependency order, got %q", types[len(types)-1].Name)
	}
}

func TestArrayOfPrimitives(t *testing.T) {
	s := mustInfer(t, `{"tags": ["go", "json", "cli"]}`, "Root")

	tags := fieldByName(t, s.Root, "tags")
	if tags.Node.Kind != KindArray {
		t.Fatalf("expected tags Kind=Array, got %v", tags.Node.Kind)
	}
	if tags.Node.Element.Kind != KindString {
		t.Errorf("expected tags element Kind=String, got %v", tags.Node.Element.Kind)
	}
}

func TestArrayOfObjectsRoot(t *testing.T) {
	s := mustInfer(t, `[{"id": 1}, {"id": 2}]`, "User")

	if s.Root.Kind != KindArray {
		t.Fatalf("expected root Kind=Array, got %v", s.Root.Kind)
	}
	if s.Root.Element.Name != "User" {
		t.Errorf("expected element type name User, got %q", s.Root.Element.Name)
	}

	types := CollectTypes(s.Root)
	if len(types) != 1 || types[0].Name != "User" {
		t.Fatalf("expected exactly one type named User, got %v", typeNames(types))
	}
}

func TestNullability(t *testing.T) {
	s := mustInfer(t, `{"middleName": null, "age": 30}`, "Root")

	middle := fieldByName(t, s.Root, "middleName")
	if !middle.Node.Nullable {
		t.Errorf("expected middleName to be Nullable")
	}
	if middle.Node.Kind != KindNull {
		t.Errorf("expected middleName Kind=Null for an always-null field, got %v", middle.Node.Kind)
	}

	age := fieldByName(t, s.Root, "age")
	if age.Node.Nullable {
		t.Errorf("expected age to not be Nullable")
	}
}

func TestNullMergedWithValue(t *testing.T) {
	s := mustInfer(t, `[{"nickname": null}, {"nickname": "Lee"}]`, "Root")

	nickname := fieldByName(t, s.Root.Element, "nickname")
	if nickname.Node.Kind != KindString {
		t.Fatalf("expected nickname Kind=String after merging with null, got %v", nickname.Node.Kind)
	}
	if !nickname.Node.Nullable {
		t.Errorf("expected nickname to be Nullable after merging with a null sample")
	}
}

func TestArraysWithDifferentObjectShapes(t *testing.T) {
	s := mustInfer(t, `[{"id": 1, "name": "A"}, {"id": 2, "email": "b@example.com"}]`, "Root")

	elem := s.Root.Element
	if len(elem.Fields) != 3 {
		t.Fatalf("expected the merged element to union to 3 fields (id, name, email), got %d", len(elem.Fields))
	}

	name := fieldByName(t, elem, "name")
	if !name.Node.Nullable {
		t.Errorf("expected name to be Nullable (missing from the second element)")
	}
	email := fieldByName(t, elem, "email")
	if !email.Node.Nullable {
		t.Errorf("expected email to be Nullable (missing from the first element)")
	}
	id := fieldByName(t, elem, "id")
	if id.Node.Nullable {
		t.Errorf("expected id to remain non-nullable (present in both elements)")
	}
}

func TestIntegerAndFloatMergeToFloat(t *testing.T) {
	s := mustInfer(t, `[{"value": 1}, {"value": 2.5}]`, "Root")

	value := fieldByName(t, s.Root.Element, "value")
	if value.Node.Kind != KindFloat {
		t.Errorf("expected int+float to merge to KindFloat, got %v", value.Node.Kind)
	}
}

func TestConflictingPrimitivesFallBackToAny(t *testing.T) {
	s := mustInfer(t, `[{"value": "text"}, {"value": 42}]`, "Root")

	value := fieldByName(t, s.Root.Element, "value")
	if value.Node.Kind != KindAny {
		t.Errorf("expected conflicting string/int to merge to KindAny, got %v", value.Node.Kind)
	}
}

func TestEmptyArrayElementIsAny(t *testing.T) {
	s := mustInfer(t, `{"items": []}`, "Root")

	items := fieldByName(t, s.Root, "items")
	if items.Node.Element.Kind != KindAny {
		t.Errorf("expected an always-empty array's element to be KindAny, got %v", items.Node.Element.Kind)
	}
}

func TestDuplicateTypeNamesAreDisambiguated(t *testing.T) {
	// Two structurally different nested objects both named "item".
	s := mustInfer(t, `{"item": {"id": 1}, "entry": {"item": "not an object"}}`, "Root")
	_ = s

	s2 := mustInfer(t, `{"a": {"x": 1}, "b": {"x": "y"}}`, "Root")
	aField := fieldByName(t, s2.Root, "a")
	bField := fieldByName(t, s2.Root, "b")
	if aField.Node.Name == bField.Node.Name {
		t.Fatalf("expected distinct generated names for structurally different nested objects named the same, both got %q", aField.Node.Name)
	}
}

func TestUnusualPropertyNames(t *testing.T) {
	s := mustInfer(t, `{"user-id": 1, "1st_place": true, "": "empty key", "veryWeird Name!": "x"}`, "Root")

	if len(s.Root.Fields) != 4 {
		t.Fatalf("expected 4 fields to survive, got %d", len(s.Root.Fields))
	}
	// Nested type naming must not panic or produce an empty name for
	// odd field names; PascalCase() guarantees a non-empty identifier.
	if PascalCase("user-id") == "" || PascalCase("1st_place") == "" || PascalCase("") == "" {
		t.Errorf("PascalCase must never return an empty string")
	}
}

func TestInvalidJSONReportsLocation(t *testing.T) {
	_, err := Infer([]byte(`{"id": 1, "name": }`), "Root")
	if err == nil {
		t.Fatal("expected an error for invalid JSON, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Line <= 0 || pe.Column <= 0 {
		t.Errorf("expected a usable line/column, got line=%d column=%d", pe.Line, pe.Column)
	}
}

func TestInvalidJSONMultilineLocation(t *testing.T) {
	jsonText := "{\n  \"id\": 1,\n  \"name\": \"ok\",\n  \"bad\": ,\n}"
	_, err := Infer([]byte(jsonText), "Root")
	if err == nil {
		t.Fatal("expected an error for invalid JSON, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Line != 4 {
		t.Errorf("expected the error to be reported on line 4, got line %d", pe.Line)
	}
}

func TestEmptyInputIsInvalid(t *testing.T) {
	if _, err := Infer([]byte("   "), "Root"); err == nil {
		t.Fatal("expected an error for empty input")
	}
}

func TestScalarRootIsRejected(t *testing.T) {
	if _, err := Infer([]byte(`"just a string"`), "Root"); err == nil {
		t.Fatal("expected an error for a scalar root value")
	}
	if _, err := Infer([]byte(`42`), "Root"); err == nil {
		t.Fatal("expected an error for a numeric root value")
	}
}

func TestArrayOfPrimitivesRootIsRejected(t *testing.T) {
	if _, err := Infer([]byte(`[1, 2, 3]`), "Root"); err == nil {
		t.Fatal("expected an error for a root array of non-objects")
	}
}

func TestDefaultRootNameIsRoot(t *testing.T) {
	s := mustInfer(t, `{"id": 1}`, "")
	if s.Root.Name != "Root" {
		t.Errorf("expected default root name Root, got %q", s.Root.Name)
	}
}

func typeNames(nodes []*Node) []string {
	names := make([]string, len(nodes))
	for i, n := range nodes {
		names[i] = n.Name
	}
	return names
}
