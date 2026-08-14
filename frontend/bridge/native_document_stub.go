//go:build !windows && !conductino_cgo

package bridge

func nativeDocumentExtract(path string) (string, error) {
	return "", errNativeUnavailable
}
