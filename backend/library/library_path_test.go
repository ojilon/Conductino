package library

import (
	"path/filepath"
	"strings"
	"testing"
)

// sanitizeLibraryRel mirrors the rule library paths must follow:
// no absolute paths, no ".." segments. Adjust if library.go gains a shared helper.
func sanitizeLibraryRel(rel string) (string, bool) {
	rel = strings.TrimSpace(rel)
	rel = strings.ReplaceAll(rel, "\\", "/")
	if rel == "" {
		return "", false
	}
	if filepath.IsAbs(rel) {
		return "", false
	}
	parts := strings.Split(rel, "/")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		if p == ".." {
			return "", false
		}
		clean = append(clean, p)
	}
	if len(clean) == 0 {
		return "", false
	}
	return strings.Join(clean, "/"), true
}

func TestSanitizeLibraryRel(t *testing.T) {
	okCases := map[string]string{
		"plantphysiology/growth": "plantphysiology/growth",
		"a/b/c":                  "a/b/c",
		"./topic":                "topic",
	}
	for in, want := range okCases {
		got, ok := sanitizeLibraryRel(in)
		if !ok || got != want {
			t.Fatalf("%q: got %q ok=%v want %q", in, got, ok, want)
		}
	}
	bad := []string{"", "..", "../x", "a/../b", "/abs", `C:\abs`}
	for _, in := range bad {
		if _, ok := sanitizeLibraryRel(in); ok {
			t.Fatalf("expected reject %q", in)
		}
	}
}
