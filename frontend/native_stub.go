//go:build !windows || conductino_cgo

package main

// Fallback when Windows LoadLibrary path is not used.

func NativeInit(dataDir string) {}

func NativeDocumentExtract(path string) (string, error) {
	return "", errNativeUnavailable
}

var errNativeUnavailable = nativeError("conductino_core not linked")

type nativeError string

func (e nativeError) Error() string { return string(e) }
