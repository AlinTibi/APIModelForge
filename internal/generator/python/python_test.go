package python

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"APIModelForge/internal/schema"
)

const sampleJSON = `{
	"id": 1,
	"full-name": "Ada Lovelace",
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

func TestPythonGeneratesExpectedDataclasses(t *testing.T) {
	code := mustGenerate(t)

	for _, want := range []string{
		"from __future__ import annotations",
		"from dataclasses import dataclass",
		"@dataclass(kw_only=True)",
		"class Person:",
		"class Address:",
		"class Friend:",
		"id: int",
		"tags: list[str]",
		"address: Address",
		"friends: list[Friend]",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("expected generated code to contain %q\n--- generated ---\n%s", want, code)
		}
	}
}

func TestPythonSanitizesInvalidAndReservedFieldNames(t *testing.T) {
	code := mustGenerate(t)

	if !strings.Contains(code, `full_name: str  # json: "full-name"`) {
		t.Errorf("expected full-name to become full_name with an original-name comment\n--- generated ---\n%s", code)
	}
	if !strings.Contains(code, `class_: str  # json: "class"`) {
		t.Errorf("expected the reserved word 'class' to become class_\n--- generated ---\n%s", code)
	}
}

func TestPythonNullableFieldUsesOptionalWithNoneDefault(t *testing.T) {
	code := mustGenerate(t)
	if !strings.Contains(code, "nickname: Optional[Any] = None") {
		t.Errorf("expected nickname to be Optional[Any] = None\n--- generated ---\n%s", code)
	}
}

// TestPythonSyntaxIsValid compiles the generated source with the system
// Python interpreter when one is available, as a strong correctness check
// beyond string matching. It skips (not fails) when no interpreter is on
// PATH, since CI/build environments are not guaranteed to have one.
func TestPythonSyntaxIsValid(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		python, err = exec.LookPath("python")
	}
	if err != nil {
		t.Skip("no python interpreter found on PATH; skipping syntax validation")
	}
	// On Windows, "python" can resolve to a Microsoft Store app-execution
	// alias stub even when no real interpreter is installed; probe before
	// trusting it.
	if out, err := exec.Command(python, "--version").CombinedOutput(); err != nil || strings.Contains(string(out), "was not found") {
		t.Skip("python on PATH is a non-functional stub; skipping syntax validation")
	}

	code := mustGenerate(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "generated_models.py")
	if err := os.WriteFile(file, []byte(code), 0o644); err != nil {
		t.Fatalf("failed to write generated file: %v", err)
	}

	cmd := exec.Command(python, "-m", "py_compile", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated Python failed to compile: %v\n%s\n--- generated ---\n%s", err, out, code)
	}
}
