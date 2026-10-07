package generator

import (
	"APIModelForge/internal/schema"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func regressionInput() string {
	keys := []string{"user-id", "user_id", "user id", "UserId", "userId", "UserId2", "123name", "class", "class_", "type", "default", "namespace", "quote\"", "slash\\", "line\n", "tab\t", "control\x00", "accenté", "漢字", "😀", "`", "", "a,b", "-", "__annotations__", "__dict__", "__post_init__", "MarshalJSON", "UnmarshalJSON", "List2"}
	var entries []string
	for _, key := range keys {
		quoted, _ := json.Marshal(key)
		entries = append(entries, string(quoted)+`:1`)
	}
	return "{" + strings.Join(entries, ",") + `,"nested":{"text":"hello","missing":null},"arrays":[[1,2],[3]],"mixed":[1,"a",null],"empty":{}}`
}

func regressionCode(t *testing.T, lang Language) string {
	t.Helper()
	s, err := schema.Infer([]byte(regressionInput()), "List")
	if err != nil {
		t.Fatal(err)
	}
	result, err := Generate(lang, s.Root, "List")
	if err != nil {
		t.Fatal(err)
	}
	return result.Code
}

func TestAllTargetsKeepUniqueDeterministicProperties(t *testing.T) {
	patterns := map[Language]string{CSharp: `(?m)^    public \S+ (\w+) \{`, Go: "(?m)^\\t(\\w+) [^\\n]+ `json:", Python: `(?m)^    (\w+):`, Kotlin: "(?m)^    val (`?\\w+`?):", TypeScript: `(?m)^  (.+?)\??: `}
	for _, lang := range All {
		t.Run(string(lang), func(t *testing.T) {
			code := regressionCode(t, lang)
			if code != regressionCode(t, lang) {
				t.Fatal("nondeterministic output")
			}
			fields := regexp.MustCompile(patterns[lang]).FindAllStringSubmatch(code, -1)
			if len(fields) != 36 {
				t.Fatalf("dropped properties: got %d, want 36\n%s", len(fields), code)
			}
			used := map[string]bool{}
			for _, field := range fields[len(fields)-34:] {
				if used[field[1]] {
					t.Fatalf("duplicate field %q", field[1])
				}
				used[field[1]] = true
			}
			if strings.ContainsRune(code, '\x00') || strings.Contains(code, "line\n") || strings.Contains(code, "tab\t") {
				t.Fatal("unescaped control character")
			}
		})
	}
}

func TestGeneratedGoCompilesAndPreservesWireNames(t *testing.T) {
	code := regressionCode(t, Go)
	if _, err := parser.ParseFile(token.NewFileSet(), "models.go", code, parser.AllErrors); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeTestFile(t, dir, "models.go", code)
	program := "package models\nimport(\"encoding/json\";\"reflect\";\"testing\")\nfunc TestWireNames(t *testing.T){var model List2; input:=[]byte(" + strconv.Quote(regressionInput()) + "); if err:=json.Unmarshal(input,&model);err!=nil{t.Fatal(err)}; output,err:=json.Marshal(model);if err!=nil{t.Fatal(err)};var a,b interface{};json.Unmarshal(input,&a);json.Unmarshal(output,&b);delete(a.(map[string]interface{})[\"nested\"].(map[string]interface{}),\"missing\");if !reflect.DeepEqual(a,b){t.Fatalf(\"wire names lost: %s\",output)}}"
	writeTestFile(t, dir, "models_test.go", program)
	cmd := exec.Command("go", "test", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=off", "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Go compile/roundtrip: %v\n%s", err, out)
	}
}

func TestGeneratedPythonCompilesAndLoads(t *testing.T) {
	python, err := exec.LookPath("python")
	if err != nil {
		t.Skip("Python not installed")
	}
	if _, err := exec.Command(python, "--version").CombinedOutput(); err != nil {
		t.Skip("Python unavailable")
	}
	dir := t.TempDir()
	writeTestFile(t, dir, "models.py", regressionCode(t, Python))
	cmd := exec.Command(python, "-c", "import runpy, dataclasses; scope = runpy.run_path('models.py'); model = scope['List2']; model(**{f.name: 1 for f in dataclasses.fields(model)})")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Python syntax/runtime: %v\n%s", err, out)
	}
}

func TestGeneratedTypeScriptCompiles(t *testing.T) {
	compiler, err := filepath.Abs("../../frontend/node_modules/typescript/bin/tsc")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(compiler); err != nil {
		t.Skip("run npm ci in frontend for compiler test")
	}
	dir := t.TempDir()
	writeTestFile(t, dir, "models.ts", regressionCode(t, TypeScript))
	cmd := exec.Command("node", compiler, "--noEmit", "--strict", "--skipLibCheck", "--target", "ES2022", "models.ts")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("TypeScript compile: %v\n%s", err, out)
	}
}

func TestGeneratedCSharpCompiles(t *testing.T) {
	dotnet, err := exec.LookPath("dotnet")
	if err != nil {
		t.Skip(".NET SDK not installed")
	}
	sdks, err := exec.Command(dotnet, "--list-sdks").CombinedOutput()
	if err != nil || !strings.Contains(string(sdks), "8.0.") {
		t.Skip(".NET 8 SDK not installed")
	}
	dir := t.TempDir()
	writeTestFile(t, dir, "Models.cs", regressionCode(t, CSharp))
	writeTestFile(t, dir, "Models.csproj", `<Project Sdk="Microsoft.NET.Sdk"><PropertyGroup><TargetFramework>net8.0</TargetFramework><Nullable>enable</Nullable></PropertyGroup></Project>`)
	writeTestFile(t, dir, "NuGet.Config", `<configuration><packageSources><clear /></packageSources></configuration>`)
	cmd := exec.Command(dotnet, "build", "--nologo", "--verbosity", "quiet")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("C# compile: %v\n%s", err, out)
	}
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}
