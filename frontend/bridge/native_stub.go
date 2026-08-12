//go:build !windows && !conductino_cgo

package bridge

import "log"

func NativeAvailable() bool { return false }

func NativeInit(dataDir string) error {
	log.Printf("[native] conductino_core not linked; using Go NoteStore")
	return nil
}

func NativeShutdown() {}

func NativeNotesSaveJSON(json string) error {
	return errNativeUnavailable
}

func NativeNotesSearch(query string) (string, error) {
	return "", errNativeUnavailable
}

func NativeSettingsGet(key string) (string, error) {
	return "", errNativeUnavailable
}

func NativeSettingsSet(key, value string) error {
	return errNativeUnavailable
}

func NativeVersion() string { return "" }
