package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"APIModelForge/internal/generator"
	"APIModelForge/internal/schema"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App holds the Wails runtime context and exposes the backend API consumed
// by the frontend. All generation logic lives in internal/schema and
// internal/generator; App is a thin binding layer.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GenerateRequest is the payload sent from the frontend to generate models
// for every supported language at once.
type GenerateRequest struct {
	JSON     string `json:"json"`
	RootName string `json:"rootName"`
}

// GeneratedFile is one language's rendered output.
type GeneratedFile struct {
	Language   string `json:"language"`
	Label      string `json:"label"`
	Code       string `json:"code"`
	FileName   string `json:"fileName"`
	ModelCount int    `json:"modelCount"`
}

// GenerateResponse carries either the generated files for every language,
// or a validation error with its approximate source location.
type GenerateResponse struct {
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Line    int             `json:"line,omitempty"`
	Column  int             `json:"column,omitempty"`
	Files   []GeneratedFile `json:"files,omitempty"`
}

// GenerateModels parses the given JSON and renders models for every
// supported output language in one call.
func (a *App) GenerateModels(req GenerateRequest) GenerateResponse {
	s, err := schema.Infer([]byte(req.JSON), req.RootName)
	if err != nil {
		return GenerateResponse{Success: false, Error: err.Error(), Line: parseErrorLine(err), Column: parseErrorColumn(err)}
	}

	files := make([]GeneratedFile, 0, len(generator.All))
	for _, lang := range generator.All {
		res, err := generator.Generate(lang, s.Root, req.RootName)
		if err != nil {
			return GenerateResponse{Success: false, Error: err.Error()}
		}
		info := generator.InfoFor(lang)
		files = append(files, GeneratedFile{
			Language:   string(lang),
			Label:      info.Label,
			Code:       res.Code,
			FileName:   res.FileName,
			ModelCount: res.ModelCount,
		})
	}

	return GenerateResponse{Success: true, Files: files}
}

// ValidationResult is the outcome of a standalone JSON validation check.
type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// ValidateJSON checks the given text without generating any code, for the
// toolbar's "Validate" action.
func (a *App) ValidateJSON(jsonText string) ValidationResult {
	s, err := schema.Infer([]byte(jsonText), "Root")
	if err != nil {
		return ValidationResult{Valid: false, Error: err.Error(), Line: parseErrorLine(err), Column: parseErrorColumn(err)}
	}

	count := len(schema.CollectTypes(s.Root))
	noun := "model"
	if count != 1 {
		noun = "models"
	}
	return ValidationResult{Valid: true, Message: fmt.Sprintf("Valid JSON — %d %s detected.", count, noun)}
}

func parseErrorLine(err error) int {
	if pe, ok := err.(*schema.ParseError); ok {
		return pe.Line
	}
	return 0
}

func parseErrorColumn(err error) int {
	if pe, ok := err.(*schema.ParseError); ok {
		return pe.Column
	}
	return 0
}

// OpenFileResult is the outcome of the native "open JSON file" dialog.
type OpenFileResult struct {
	Cancelled bool   `json:"cancelled"`
	Content   string `json:"content,omitempty"`
	FileName  string `json:"fileName,omitempty"`
	Error     string `json:"error,omitempty"`
}

// OpenJSONFile shows a native file picker and reads the selected file's
// contents.
func (a *App) OpenJSONFile() OpenFileResult {
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Open JSON file",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return OpenFileResult{Error: err.Error()}
	}
	if path == "" {
		return OpenFileResult{Cancelled: true}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return OpenFileResult{Error: err.Error()}
	}

	return OpenFileResult{Content: string(data), FileName: filepath.Base(path)}
}

// SaveFileResult is the outcome of the native "save generated code" dialog.
type SaveFileResult struct {
	Cancelled bool   `json:"cancelled"`
	Path      string `json:"path,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SaveTextFile shows a native save dialog defaulting to defaultFileName and
// writes content to the chosen path.
func (a *App) SaveTextFile(defaultFileName string, content string) SaveFileResult {
	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Save generated code",
		DefaultFilename: defaultFileName,
	})
	if err != nil {
		return SaveFileResult{Error: err.Error()}
	}
	if path == "" {
		return SaveFileResult{Cancelled: true}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return SaveFileResult{Error: err.Error()}
	}

	return SaveFileResult{Path: path}
}
