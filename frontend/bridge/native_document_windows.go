//go:build windows && !conductino_cgo

package bridge

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	docOnce   sync.Once
	pDocExtract *syscall.LazyProc
)

func ensureDocProcs() {
	docOnce.Do(func() {
		if !ensureNative() || nativeDll == nil {
			return
		}
		pDocExtract = nativeDll.NewProc("conductino_document_extract")
	})
}

func nativeDocumentExtract(path string) (string, error) {
	ensureDocProcs()
	if pDocExtract == nil {
		return "", errNativeUnavailable
	}
	pptr, err := syscall.BytePtrFromString(path)
	if err != nil {
		return "", err
	}
	var outPtr uintptr
	var outLen uintptr
	r, _, callErr := pDocExtract.Call(
		uintptr(unsafe.Pointer(pptr)),
		uintptr(unsafe.Pointer(&outPtr)),
		uintptr(unsafe.Pointer(&outLen)),
	)
	if r != 0 {
		return "", fmt.Errorf("document_extract: %d (%v)", r, callErr)
	}
	if outPtr == 0 {
		return "", nil
	}
	defer pFree.Call(outPtr)
	b := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	return string(b), nil
}
