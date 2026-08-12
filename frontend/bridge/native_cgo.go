//go:build conductino_cgo

package bridge

/*
#cgo CFLAGS: -I${SRCDIR}/../../backend/include
#cgo windows LDFLAGS: -L${SRCDIR}/../../backend/build -L${SRCDIR}/../../backend/build/Release -lconductino_core
#cgo linux LDFLAGS: -L${SRCDIR}/../../backend/build -lconductino_core -lstdc++ -lm
#cgo darwin LDFLAGS: -L${SRCDIR}/../../backend/build -lconductino_core -lc++

#include "conductino/core.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

func NativeAvailable() bool { return true }

func NativeVersion() string {
	return C.GoString(C.conductino_version())
}

func NativeInit(dataDir string) error {
	cs := C.CString(dataDir)
	defer C.free(unsafe.Pointer(cs))
	if rc := C.conductino_init(cs); rc != 0 {
		return fmt.Errorf("conductino_init: %d", int(rc))
	}
	log.Printf("[native] cgo init ok version=%s", NativeVersion())
	return nil
}

func NativeShutdown() {
	C.conductino_shutdown()
}

func NativeNotesSaveJSON(json string) error {
	cs := C.CString(json)
	defer C.free(unsafe.Pointer(cs))
	if rc := C.conductino_notes_save_json(cs, C.size_t(len(json))); rc != 0 {
		return fmt.Errorf("notes_save_json: %d", int(rc))
	}
	return nil
}

func NativeNotesSearch(query string) (string, error) {
	cq := C.CString(query)
	defer C.free(unsafe.Pointer(cq))
	var out *C.char
	var n C.size_t
	if rc := C.conductino_notes_search(cq, &out, &n); rc != 0 {
		return "", fmt.Errorf("notes_search: %d", int(rc))
	}
	if out == nil {
		return "[]", nil
	}
	defer C.conductino_free(unsafe.Pointer(out))
	return C.GoStringN(out, C.int(n)), nil
}

func NativeSettingsGet(key string) (string, error) {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var out *C.char
	var n C.size_t
	rc := C.conductino_settings_get(ck, &out, &n)
	if rc < 0 {
		return "", fmt.Errorf("settings_get: %d", int(rc))
	}
	if rc == 1 {
		return "", nil
	}
	if out == nil {
		return "", nil
	}
	defer C.conductino_free(unsafe.Pointer(out))
	return C.GoStringN(out, C.int(n)), nil
}

func NativeSettingsSet(key, value string) error {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	cv := C.CString(value)
	defer C.free(unsafe.Pointer(cv))
	if rc := C.conductino_settings_set(ck, cv); rc != 0 {
		return fmt.Errorf("settings_set: %d", int(rc))
	}
	return nil
}
