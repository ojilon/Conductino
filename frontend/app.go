package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"conductino-backend/bridge"
	"conductino-backend/library"
	"conductino-backend/native"
)

// App is the Wails-bound application API.
type App struct {
	ctx     context.Context
	dataDir string
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{}
}

// startup initializes the application data directory and native backend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	a.dataDir = filepath.Join(cwd, "conductino-data")
	_ = os.MkdirAll(filepath.Join(a.dataDir, "cache", "docs"), 0o755)
	_ = os.MkdirAll(filepath.Join(a.dataDir, "imports"), 0o755)
	native.NativeInit(a.dataDir)
	library.EnsureLibrary(a.dataDir)
}

// Greet returns a greeting string.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "researcher"
	}
	return bridge.Greet(name)
}

// AppInfo returns application information map.
func (a *App) AppInfo() map[string]string {
	return bridge.AppInfo()
}

// WindowMinimise minimizes the application window.
func (a *App) WindowMinimise() {
	if a.ctx == nil {
		return
	}
	runtime.WindowMinimise(a.ctx)
}

// WindowToggleMaximise toggles the application window maximization.
func (a *App) WindowToggleMaximise() {
	if a.ctx == nil {
		return
	}
	runtime.WindowToggleMaximise(a.ctx)
}

// WindowClose closes the application window.
func (a *App) WindowClose() {
	if a.ctx == nil {
		return
	}
	runtime.Quit(a.ctx)
}

// OpenURL opens a URL in the system default browser.
func (a *App) OpenURL(url string) error {
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("empty url")
	}
	runtime.BrowserOpenURL(a.ctx, url)
	return nil
}

// OpenFile shows a native file dialog and returns the selected path (or "").
func (a *App) OpenFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open document",
		Filters: []runtime.FileFilter{
			{DisplayName: "Documents (txt, md, pdf, docx, json)", Pattern: "*.txt;*.md;*.markdown;*.pdf;*.docx;*.doc;*.json;*.csv;*.tex"},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// ExtractDocument returns plain text for a local path.
func (a *App) ExtractDocument(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	// Try native C++ backend first, fall back to Go text extraction
	if text, err := native.NativeDocumentExtract(path); err == nil && text != "" {
		return text, nil
	}
	return extractTextGo(path)
}

// ImportDocument copies src into data_dir/imports and returns the new absolute path.
func (a *App) ImportDocument(src string) (string, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return "", fmt.Errorf("empty path")
	}
	if a.dataDir == "" {
		return "", fmt.Errorf("data dir not ready")
	}
	in, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("%d_%s", os.Getpid(), filepath.Base(src))
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' {
			return '_'
		}
		return r
	}, name)
	dest := filepath.Join(a.dataDir, "imports", name)
	if err := os.WriteFile(dest, in, 0o644); err != nil {
		return "", err
	}
	return dest, nil
}

// extractTextGo extracts text from a file based on its extension.
func extractTextGo(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".txt": true, ".md": true, ".markdown": true, ".csv": true, ".tsv": true,
		".json": true, ".jsonl": true, ".log": true, ".xml": true, ".html": true,
		".htm": true, ".css": true, ".js": true, ".ts": true, ".py": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true, ".go": true,
		".rs": true, ".java": true, ".yaml": true, ".yml": true, ".toml": true,
		".ini": true, ".tex": true, ".bib": true,
	}
	if textExts[ext] || ext == "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		s := string(b)
		s = strings.ReplaceAll(s, "\x00", "")
		return s, nil
	}
	if ext == ".pdf" || ext == ".docx" || ext == ".doc" {
		return fmt.Sprintf(
			"[Conductino] Binary document (%s).\nPath: %s\n\n"+
				"Full PDF/DOCX extract needs the C++ backend library or an export to text.\n"+
				"Workaround: copy text from the PDF and use Paste text in Study.\n"+
				"See backend/features/document/README.md.",
			ext, path,
		), nil
	}
	return "", fmt.Errorf("unsupported file type: %s", ext)
}