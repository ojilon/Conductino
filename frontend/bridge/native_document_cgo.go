//go:build conductino_cgo

package bridge

/*
#include "conductino/core.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func nativeDocumentExtract(path string) (string, error) {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	var out *C.char
	var n C.size_t
	if rc := C.conductino_document_extract(cs, &out, &n); rc != 0 {
		return "", fmt.Errorf("document_extract: %d", int(rc))
	}
	if out == nil {
		return "", nil
	}
	defer C.conductino_free(unsafe.Pointer(out))
	return C.GoStringN(out, C.int(n)), nil
}
