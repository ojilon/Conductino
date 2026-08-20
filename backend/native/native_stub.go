//go:build !windows || conductino_cgo

package native

// NativeInit initializes the native backend (C++ conductino_core DLL).
// On non-Windows platforms or when the C++ library is not linked, this is a no-op.
func NativeInit(dataDir string) {}

// NativeDocumentExtract extracts text from a document path.
// Returns an error indicating the C++ library is not linked when not available.
func NativeDocumentExtract(path string) (string, error) {
	return "", ErrNativeUnavailable
}

var errNativeUnavailable = "conductino_core not linked"

var ErrNativeUnavailable = "conductino_core not linked"

// nativeError implements the error interface.
type nativeError string

func (e nativeError) Error() string { return string(e) }