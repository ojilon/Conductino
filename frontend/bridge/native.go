package bridge

/*
Thin Go helper toward the C++23 backend (conductino_core).

When the shared library is built and cgo is enabled, implement the // #cgo
LDFLAGS/CFLAGS and call conductino_* from core.h.

Until then these functions are safe no-ops so the frontend keeps working with
handlers.NoteStore.

Build backend:
  cd backend && cmake -B build && cmake --build build

Then link, for example:
  #cgo LDFLAGS: -L${SRCDIR}/../../backend/build -lconductino_core
  #cgo CFLAGS: -I${SRCDIR}/../../backend/include
*/

import "log"

// NativeAvailable reports whether the C++ core is linked.
// Always false until cgo bindings are filled in.
func NativeAvailable() bool {
	return false
}

// NativeInit initializes the C++ core with a data directory.
func NativeInit(dataDir string) error {
	if !NativeAvailable() {
		log.Printf("[native] conductino_core not linked; using Go NoteStore")
		return nil
	}
	// TODO(cgo): call conductino_init
	return nil
}

// NativeShutdown flushes the C++ core.
func NativeShutdown() {
	if !NativeAvailable() {
		return
	}
	// TODO(cgo): call conductino_shutdown
}

// NativeNotesSaveJSON forwards a note JSON payload to C++.
func NativeNotesSaveJSON(json string) error {
	if !NativeAvailable() {
		return errNativeUnavailable
	}
	// TODO(cgo): conductino_notes_save_json
	return nil
}

// NativeNotesSearch returns JSON array of notes from C++.
func NativeNotesSearch(query string) (string, error) {
	if !NativeAvailable() {
		return "", errNativeUnavailable
	}
	// TODO(cgo): conductino_notes_search
	return "[]", nil
}

type nativeError string

func (e nativeError) Error() string { return string(e) }

const errNativeUnavailable = nativeError("conductino_core not linked")
