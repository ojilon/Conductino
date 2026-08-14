package bridge

// NativeDocumentExtract is implemented per platform (windows LoadLibrary / cgo / stub).

func NativeDocumentExtract(path string) (string, error) {
	return nativeDocumentExtract(path)
}
