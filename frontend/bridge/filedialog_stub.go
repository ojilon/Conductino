//go:build !windows

package bridge

// Non-Windows: return empty so UI can fall back to path prompt / paste.
func OpenFileDialog() (string, error) {
	return "", nil
}
