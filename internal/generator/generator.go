// Package generator dispatches an inferred schema.Node tree to the
// per-language renderers (csharp, typescript, kotlin, python, golang).
// Every renderer consumes the exact same tree; only the rendering rules
// differ.
package generator

import (
	"fmt"

	"APIModelForge/internal/generator/csharp"
	"APIModelForge/internal/generator/golang"
	"APIModelForge/internal/generator/kotlin"
	"APIModelForge/internal/generator/python"
	"APIModelForge/internal/generator/typescript"
	"APIModelForge/internal/schema"
)

// Language identifies one of the supported output targets.
type Language string

const (
	CSharp     Language = "csharp"
	TypeScript Language = "typescript"
	Kotlin     Language = "kotlin"
	Python     Language = "python"
	Go         Language = "go"
)

// All lists every supported language in the order they should appear in
// the UI.
var All = []Language{CSharp, TypeScript, Kotlin, Python, Go}

// Info describes a language for display purposes.
type Info struct {
	Language Language
	Label    string
	FileExt  string
}

var infoByLanguage = map[Language]Info{
	CSharp:     {CSharp, "C#", ".cs"},
	TypeScript: {TypeScript, "TypeScript", ".ts"},
	Kotlin:     {Kotlin, "Kotlin", ".kt"},
	Python:     {Python, "Python", ".py"},
	Go:         {Go, "Go", ".go"},
}

// InfoFor returns display metadata for lang.
func InfoFor(lang Language) Info {
	return infoByLanguage[lang]
}

// Result is the rendered output for one language.
type Result struct {
	Code       string
	ModelCount int
	FileName   string
}

// Generate renders root as source code for lang. rootName is used only to
// derive a sensible output file name; the type name itself was already
// assigned during schema.Infer.
func Generate(lang Language, root *schema.Node, rootName string) (Result, error) {
	types := schema.CollectTypes(root)
	info := InfoFor(lang)
	if info.Language == "" {
		return Result{}, fmt.Errorf("unsupported language: %s", lang)
	}

	var code string
	switch lang {
	case CSharp:
		code = csharp.Generate(types)
	case TypeScript:
		code = typescript.Generate(types)
	case Kotlin:
		code = kotlin.Generate(types)
	case Python:
		code = python.Generate(types)
	case Go:
		code = golang.Generate(types)
	default:
		return Result{}, fmt.Errorf("unsupported language: %s", lang)
	}

	fileBase := rootNameForFile(types, rootName)
	return Result{
		Code:       code,
		ModelCount: len(types),
		FileName:   fileBase + info.FileExt,
	}, nil
}

func rootNameForFile(types []*schema.Node, rootName string) string {
	if len(types) > 0 {
		return types[len(types)-1].Name
	}
	if rootName == "" {
		return "Root"
	}
	return rootName
}
