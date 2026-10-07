package main

import "testing"

func TestValidationBlocksMalformedGeneration(t *testing.T) {
	for _, input := range []string{`{"a":1,}`, `{"a:1}`, `{"a":1 "b":2}`, `{"a":[1}`, ``, `{} garbage`, `{} /`, `{} {`, `{"a":1,"a":2}`} {
		t.Run(input, func(t *testing.T) {
			app := NewApp()
			result := app.GenerateModels(GenerateRequest{JSON: input, RootName: "Root"})
			if result.Success || result.Error == "" || len(result.Files) != 0 {
				t.Fatalf("invalid generation: %+v", result)
			}
			if validation := app.ValidateJSON(input); validation.Valid || validation.Error == "" {
				t.Fatalf("invalid validation: %+v", validation)
			}
		})
	}
}

func TestValidGenerationAllLanguages(t *testing.T) {
	for _, input := range []string{`{"id":1}`, `[{"id":1},{"id":null}]`} {
		result := NewApp().GenerateModels(GenerateRequest{JSON: input, RootName: "Root"})
		if !result.Success || len(result.Files) != 5 {
			t.Fatalf("generation failed: %+v", result)
		}
	}
}
