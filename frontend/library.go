package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const libraryIndexVersion = 1

// Bookmark is a URL (and optional local download) under a library folder.
type Bookmark struct {
	ID        string `json:"id"`
	Folder    string `json:"folder"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	LocalPath string `json:"localPath"`
	CreatedAt string `json:"createdAt"`
}

// SummaryRef points at a markdown summary file for a folder/topic.
type SummaryRef struct {
	ID        string `json:"id"`
	Folder    string `json:"folder"`
	RelPath   string `json:"relPath"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updatedAt"`
}

type libraryIndex struct {
	Version   int          `json:"version"`
	Bookmarks []Bookmark   `json:"bookmarks"`
	Summaries []SummaryRef `json:"summaries"`
}

func (a *App) libraryRoot() string {
	return filepath.Join(a.dataDir, "library")
}

func (a *App) indexPath() string {
	return filepath.Join(a.libraryRoot(), "index.json")
}

func (a *App) ensureLibrary() error {
	if a.dataDir == "" {
		return fmt.Errorf("data dir not ready")
	}
	return os.MkdirAll(a.libraryRoot(), 0o755)
}

func (a *App) loadIndex() (*libraryIndex, error) {
	if err := a.ensureLibrary(); err != nil {
		return nil, err
	}
	p := a.indexPath()
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &libraryIndex{Version: libraryIndexVersion}, nil
		}
		return nil, err
	}
	var idx libraryIndex
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	if idx.Version == 0 {
		idx.Version = libraryIndexVersion
	}
	return &idx, nil
}

func (a *App) saveIndex(idx *libraryIndex) error {
	if err := a.ensureLibrary(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.indexPath(), b, 0o644)
}

func normalizeFolder(rel string) string {
	rel = strings.TrimSpace(rel)
	rel = strings.ReplaceAll(rel, "\\", "/")
	rel = strings.Trim(rel, "/")
	parts := strings.Split(rel, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || p == "." || p == ".." {
			continue
		}
		// Keep simple folder names only
		p = strings.Map(func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
				return r
			case r == '-' || r == '_' || r == ' ':
				return r
			default:
				return -1
			}
		}, p)
		p = strings.ReplaceAll(strings.TrimSpace(p), " ", "-")
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "/")
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// LibraryRoot returns the absolute library directory.
func (a *App) LibraryRoot() (string, error) {
	if err := a.ensureLibrary(); err != nil {
		return "", err
	}
	return a.libraryRoot(), nil
}

// ListLibraryFolders returns relative folder paths that exist under library/.
func (a *App) ListLibraryFolders() ([]string, error) {
	if err := a.ensureLibrary(); err != nil {
		return nil, err
	}
	var folders []string
	root := a.libraryRoot()
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if base == "downloads" || base == "summaries" || base == "previews" {
			return filepath.SkipDir
		}
		folders = append(folders, filepath.ToSlash(rel))
		return nil
	})
	if folders == nil {
		folders = []string{}
	}
	return folders, nil
}

// CreateLibraryFolder creates library/<rel> plus downloads/ and summaries/.
func (a *App) CreateLibraryFolder(rel string) (string, error) {
	rel = normalizeFolder(rel)
	if rel == "" {
		return "", fmt.Errorf("empty folder path")
	}
	base := filepath.Join(a.libraryRoot(), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Join(base, "downloads"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(base, "summaries"), 0o755); err != nil {
		return "", err
	}
	return rel, nil
}

// ListBookmarks returns bookmarks, optionally filtered by folder ("" = all).
func (a *App) ListBookmarks(folder string) ([]Bookmark, error) {
	idx, err := a.loadIndex()
	if err != nil {
		return nil, err
	}
	folder = normalizeFolder(folder)
	var out []Bookmark
	for _, b := range idx.Bookmarks {
		if folder == "" || normalizeFolder(b.Folder) == folder {
			out = append(out, b)
		}
	}
	if out == nil {
		out = []Bookmark{}
	}
	return out, nil
}

// AddBookmark stores a URL under a library folder.
func (a *App) AddBookmark(folder, url, title string) (Bookmark, error) {
	folder = normalizeFolder(folder)
	url = strings.TrimSpace(url)
	title = strings.TrimSpace(title)
	if folder == "" || url == "" {
		return Bookmark{}, fmt.Errorf("folder and url required")
	}
	if _, err := a.CreateLibraryFolder(folder); err != nil {
		return Bookmark{}, err
	}
	if title == "" {
		title = url
	}
	idx, err := a.loadIndex()
	if err != nil {
		return Bookmark{}, err
	}
	bm := Bookmark{
		ID:        newID("bm"),
		Folder:    folder,
		URL:       url,
		Title:     title,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	idx.Bookmarks = append(idx.Bookmarks, bm)
	if err := a.saveIndex(idx); err != nil {
		return Bookmark{}, err
	}
	return bm, nil
}

// RemoveBookmark deletes a bookmark by id.
func (a *App) RemoveBookmark(id string) error {
	id = strings.TrimSpace(id)
	idx, err := a.loadIndex()
	if err != nil {
		return err
	}
	filtered := idx.Bookmarks[:0]
	for _, b := range idx.Bookmarks {
		if b.ID != id {
			filtered = append(filtered, b)
		}
	}
	idx.Bookmarks = filtered
	return a.saveIndex(idx)
}

func (a *App) defaultSummaryRel(folder string) string {
	name := filepath.Base(filepath.FromSlash(folder))
	if name == "" || name == "." {
		name = "summary"
	}
	return folder + "/summaries/" + name + "-summary.md"
}

// EnsureSummary returns the primary summary ref for a folder (creates empty file if needed).
func (a *App) EnsureSummary(folder string) (SummaryRef, error) {
	folder = normalizeFolder(folder)
	if folder == "" {
		return SummaryRef{}, fmt.Errorf("folder required")
	}
	if _, err := a.CreateLibraryFolder(folder); err != nil {
		return SummaryRef{}, err
	}
	idx, err := a.loadIndex()
	if err != nil {
		return SummaryRef{}, err
	}
	for _, s := range idx.Summaries {
		if normalizeFolder(s.Folder) == folder {
			return s, nil
		}
	}
	rel := a.defaultSummaryRel(folder)
	abs := filepath.Join(a.libraryRoot(), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return SummaryRef{}, err
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		header := "# " + filepath.Base(filepath.FromSlash(folder)) + " — knowledge summary\n\n"
		if err := os.WriteFile(abs, []byte(header), 0o644); err != nil {
			return SummaryRef{}, err
		}
	}
	ref := SummaryRef{
		ID:        newID("sum"),
		Folder:    folder,
		RelPath:   rel,
		Title:     filepath.Base(filepath.FromSlash(rel)),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	idx.Summaries = append(idx.Summaries, ref)
	if err := a.saveIndex(idx); err != nil {
		return SummaryRef{}, err
	}
	return ref, nil
}

// AppendSummary appends markdown text to the folder's primary summary file.
func (a *App) AppendSummary(folder, sectionTitle, body string) (SummaryRef, error) {
	ref, err := a.EnsureSummary(folder)
	if err != nil {
		return SummaryRef{}, err
	}
	abs := filepath.Join(a.libraryRoot(), filepath.FromSlash(ref.RelPath))
	existing, _ := os.ReadFile(abs)
	block := "\n\n---\n\n"
	if strings.TrimSpace(sectionTitle) != "" {
		block += "## " + strings.TrimSpace(sectionTitle) + "\n\n"
	}
	block += strings.TrimSpace(body) + "\n"
	block += "\n<!-- appended " + time.Now().UTC().Format(time.RFC3339) + " -->\n"
	out := append(existing, []byte(block)...)
	if err := os.WriteFile(abs, out, 0o644); err != nil {
		return SummaryRef{}, err
	}
	ref.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	idx, err := a.loadIndex()
	if err != nil {
		return ref, nil
	}
	for i := range idx.Summaries {
		if idx.Summaries[i].ID == ref.ID {
			idx.Summaries[i].UpdatedAt = ref.UpdatedAt
		}
	}
	_ = a.saveIndex(idx)
	return ref, nil
}

// ReadSummaryFile returns the text of a summary by relative path under library/.
func (a *App) ReadSummaryFile(relPath string) (string, error) {
	relPath = strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if relPath == "" || strings.Contains(relPath, "..") {
		return "", fmt.Errorf("invalid path")
	}
	abs := filepath.Join(a.libraryRoot(), filepath.FromSlash(relPath))
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteSummaryFile overwrites a summary file (after user accepts AI edits).
func (a *App) WriteSummaryFile(relPath, content string) error {
	relPath = strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if relPath == "" || strings.Contains(relPath, "..") {
		return fmt.Errorf("invalid path")
	}
	abs := filepath.Join(a.libraryRoot(), filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

// ListSummaries returns summary refs, optionally filtered by folder.
func (a *App) ListSummaries(folder string) ([]SummaryRef, error) {
	idx, err := a.loadIndex()
	if err != nil {
		return nil, err
	}
	folder = normalizeFolder(folder)
	var out []SummaryRef
	for _, s := range idx.Summaries {
		if folder == "" || normalizeFolder(s.Folder) == folder {
			out = append(out, s)
		}
	}
	if out == nil {
		out = []SummaryRef{}
	}
	return out, nil
}

// MergeSummaryFiles concatenates two summary files into a new file under destFolder.
func (a *App) MergeSummaryFiles(relPathA, relPathB, destFolder, title string) (SummaryRef, error) {
	aText, err := a.ReadSummaryFile(relPathA)
	if err != nil {
		return SummaryRef{}, err
	}
	bText, err := a.ReadSummaryFile(relPathB)
	if err != nil {
		return SummaryRef{}, err
	}
	destFolder = normalizeFolder(destFolder)
	if destFolder == "" {
		return SummaryRef{}, fmt.Errorf("dest folder required")
	}
	if _, err := a.CreateLibraryFolder(destFolder); err != nil {
		return SummaryRef{}, err
	}
	if strings.TrimSpace(title) == "" {
		title = "merged-summary"
	}
	title = normalizeFolder(title)
	title = strings.ReplaceAll(title, "/", "-")
	rel := destFolder + "/summaries/" + title + ".md"
	abs := filepath.Join(a.libraryRoot(), filepath.FromSlash(rel))
	merged := "# " + title + "\n\n" +
		"<!-- merged " + time.Now().UTC().Format(time.RFC3339) + " -->\n\n" +
		"## From A\n\n" + strings.TrimSpace(aText) + "\n\n" +
		"## From B\n\n" + strings.TrimSpace(bText) + "\n"
	if err := os.WriteFile(abs, []byte(merged), 0o644); err != nil {
		return SummaryRef{}, err
	}
	ref := SummaryRef{
		ID:        newID("sum"),
		Folder:    destFolder,
		RelPath:   rel,
		Title:     title + ".md",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	idx, err := a.loadIndex()
	if err != nil {
		return ref, nil
	}
	idx.Summaries = append(idx.Summaries, ref)
	_ = a.saveIndex(idx)
	return ref, nil
}

// PickLibraryFolder opens a directory dialog starting at the library root.
// Returns a path relative to library root when possible.
func (a *App) PickLibraryFolder() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	if err := a.ensureLibrary(); err != nil {
		return "", err
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select library folder",
		DefaultDirectory:     a.libraryRoot(),
		CanCreateDirectories: true,
	})
	if err != nil || path == "" {
		return path, err
	}
	rel, err := filepath.Rel(a.libraryRoot(), path)
	if err != nil || strings.HasPrefix(rel, "..") {
		// Outside library — return absolute path marker
		return path, nil
	}
	return filepath.ToSlash(rel), nil
}
