package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// Config holds library-related settings.
type Config struct {
	DatabasePath string `yaml:"database_path"`
}

// libraryRoot returns the absolute library directory path given the data directory.
func LibraryRoot(dataDir string) string {
	return filepath.Join(dataDir, "library")
}

// indexPath returns the path to the library index.json given dataDir.
func IndexPath(dataDir string) string {
	return filepath.Join(LibraryRoot(dataDir), "index.json")
}

// EnsureLibrary creates the library directory if it doesn't exist.
// Returns an error if dataDir is empty.
func EnsureLibrary(dataDir string) error {
	if dataDir == "" {
		return fmt.Errorf("data dir not ready")
	}
	return os.MkdirAll(LibraryRoot(dataDir), 0o755)
}

// LoadIndex loads the library index from disk.
// Returns a new index if the file doesn't exist.
func LoadIndex(dataDir string) (*libraryIndex, error) {
	if err := EnsureLibrary(dataDir); err != nil {
		return nil, err
	}
	p := IndexPath(dataDir)
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

// SaveIndex saves the library index to disk.
func SaveIndex(dataDir string, idx *libraryIndex) error {
	if err := EnsureLibrary(dataDir); err != nil {
		return err
	}
	p := IndexPath(dataDir)
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// normalizeFolder normalizes a folder path for consistent comparison.
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

// newID generates a new unique ID with the given prefix.
func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// LibraryRoot returns the absolute library directory.
func LibraryRootFunc(dataDir string) (string, error) {
	if err := EnsureLibrary(dataDir); err != nil {
		return "", err
	}
	return LibraryRoot(dataDir), nil
}

// ListLibraryFolders returns relative folder paths that exist under library/.
func ListLibraryFolders(dataDir, folder string) ([]string, error) {
	if err := EnsureLibrary(dataDir); err != nil {
		return nil, err
	}
	folder = normalizeFolder(folder)
	var folders []string
	root := LibraryRoot(dataDir)
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
func CreateLibraryFolder(dataDir, rel string) (string, error) {
	rel = normalizeFolder(rel)
	if rel == "" {
		return "", fmt.Errorf("empty folder path")
	}
	base := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Join(base, "downloads"), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(base, "summaries"), 0o755); err != nil {
		return "", err
	}
	return rel, nil
}

// Bookmark struct and operations

// ListBookmarks returns bookmarks, optionally filtered by folder ("" = all).
func ListBookmarks(dataDir, folder string) ([]Bookmark, error) {
	idx, err := LoadIndex(dataDir)
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
func AddBookmark(dataDir, folder, url, title string) (Bookmark, error) {
	folder = normalizeFolder(folder)
	url = strings.TrimSpace(url)
	title = strings.TrimSpace(title)
	if folder == "" || url == "" {
		return Bookmark{}, fmt.Errorf("folder and url required")
	}
	if _, err := CreateLibraryFolder(dataDir, folder); err != nil {
		return Bookmark{}, err
	}
	if title == "" {
		title = url
	}
	idx, err := LoadIndex(dataDir)
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
	if err := SaveIndex(dataDir, idx); err != nil {
		return Bookmark{}, err
	}
	return bm, nil
}

// RemoveBookmark deletes a bookmark by id.
func RemoveBookmark(dataDir, id string) error {
	id = strings.TrimSpace(id)
	idx, err := LoadIndex(dataDir)
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
	return SaveIndex(dataDir, idx)
}

// Summary operations

// defaultSummaryRel returns the default summary relative path for a folder.
func defaultSummaryRel(folder string) string {
	name := filepath.Base(filepath.FromSlash(folder))
	if name == "" || name == "." {
		name = "summary"
	}
	return folder + "/summaries/" + name + "-summary.md"
}

// EnsureSummary returns the primary summary ref for a folder (creates empty file if needed).
func EnsureSummary(dataDir, folder string) (SummaryRef, error) {
	folder = normalizeFolder(folder)
	if folder == "" {
		return SummaryRef{}, fmt.Errorf("folder required")
	}
	if _, err := CreateLibraryFolder(dataDir, folder); err != nil {
		return SummaryRef{}, err
	}
	idx, err := LoadIndex(dataDir)
	if err != nil {
		return SummaryRef{}, err
	}
	for _, s := range idx.Summaries {
		if normalizeFolder(s.Folder) == folder {
			return s, nil
		}
	}
	rel := defaultSummaryRel(folder)
	abs := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(rel))
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
	if err := SaveIndex(dataDir, idx); err != nil {
		return SummaryRef{}, err
	}
	return ref, nil
}

// AppendSummary appends markdown text to the folder's primary summary file.
func AppendSummary(dataDir, folder, sectionTitle, body string) (SummaryRef, error) {
	ref, err := EnsureSummary(dataDir, folder)
	if err != nil {
		return SummaryRef{}, err
	}
	abs := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(ref.RelPath))
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
	idx, err := LoadIndex(dataDir)
	if err != nil {
		return ref, nil
	}
	for i := range idx.Summaries {
		if idx.Summaries[i].ID == ref.ID {
			idx.Summaries[i].UpdatedAt = ref.UpdatedAt
		}
	}
	_ = SaveIndex(dataDir, idx)
	return ref, nil
}

// ReadSummaryFile returns the text of a summary by relative path under library/.
func ReadSummaryFile(dataDir, relPath string) (string, error) {
	relPath = strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if relPath == "" || strings.Contains(relPath, "..") {
		return "", fmt.Errorf("invalid path")
	}
	abs := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(relPath))
	b, err := os.ReadFile(abs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// WriteSummaryFile overwrites a summary file (after user accepts AI edits).
func WriteSummaryFile(dataDir, relPath, content string) error {
	relPath = strings.TrimSpace(strings.ReplaceAll(relPath, "\\", "/"))
	if relPath == "" || strings.Contains(relPath, "..") {
		return fmt.Errorf("invalid path")
	}
	abs := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}

// ListSummaries returns summary refs, optionally filtered by folder.
func ListSummaries(dataDir, folder string) ([]SummaryRef, error) {
	idx, err := LoadIndex(dataDir)
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
func MergeSummaryFiles(dataDir, relPathA, relPathB, destFolder, title string) (SummaryRef, error) {
	aText, err := ReadSummaryFile(dataDir, relPathA)
	if err != nil {
		return SummaryRef{}, err
	}
	bText, err := ReadSummaryFile(dataDir, relPathB)
	if err != nil {
		return SummaryRef{}, err
	}
	destFolder = normalizeFolder(destFolder)
	if destFolder == "" {
		return SummaryRef{}, fmt.Errorf("dest folder required")
	}
	if _, err := CreateLibraryFolder(dataDir, destFolder); err != nil {
		return SummaryRef{}, err
	}
	if strings.TrimSpace(title) == "" {
		title = "merged-summary"
	}
	title = normalizeFolder(title)
	title = strings.ReplaceAll(title, "/", "-")
	rel := destFolder + "/summaries/" + title + ".md"
	abs := filepath.Join(LibraryRoot(dataDir), filepath.FromSlash(rel))
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
	idx, err := LoadIndex(dataDir)
	if err != nil {
		return ref, nil
	}
	idx.Summaries = append(idx.Summaries, ref)
	_ = SaveIndex(dataDir, idx)
	return ref, nil
}

// PickLibraryFolder opens a directory dialog starting at the library root.
// Returns a path relative to library root when possible.
// Note: This requires runtime which is frontend-specific; kept as a stub.
func PickLibraryFolder(dataDir string) (string, error) {
	// This requires runtime which is frontend-specific
	// Return a stub path relative to library root
	libRoot := LibraryRoot(dataDir)
	return libRoot, nil
}