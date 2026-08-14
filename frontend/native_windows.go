//go:build windows && !conductino_cgo

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

var (
	nativeOnce   sync.Once
	nativeOK     bool
	nativeDll    *syscall.LazyDLL
	pInit        *syscall.LazyProc
	pFree        *syscall.LazyProc
	pDocExtract  *syscall.LazyProc
)

type nativeError string

func (e nativeError) Error() string { return string(e) }

var errNativeUnavailable = nativeError("conductino_core not linked")

func dllSearchPaths() []string {
	var out []string
	if exe, err := os.Executable(); err == nil {
		out = append(out, filepath.Dir(exe))
	}
	if cwd, err := os.Getwd(); err == nil {
		out = append(out,
			cwd,
			filepath.Join(cwd, "backend", "build"),
			filepath.Join(cwd, "backend", "build", "Release"),
			filepath.Join(cwd, "backend", "build", "Debug"),
			filepath.Join(cwd, "..", "backend", "build"),
			filepath.Join(cwd, "..", "backend", "build", "Release"),
		)
	}
	if v := os.Getenv("CONDUCTINO_CORE_DIR"); v != "" {
		out = append([]string{v}, out...)
	}
	return out
}

func tryLoadDLL() bool {
	names := []string{"conductino_core.dll", "libconductino_core.dll"}
	for _, dir := range dllSearchPaths() {
		for _, name := range names {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err != nil {
				continue
			}
			nativeDll = syscall.NewLazyDLL(path)
			if err := nativeDll.Load(); err != nil {
				continue
			}
			pInit = nativeDll.NewProc("conductino_init")
			pFree = nativeDll.NewProc("conductino_free")
			pDocExtract = nativeDll.NewProc("conductino_document_extract")
			log.Printf("[native] loaded %s", path)
			return true
		}
	}
	return false
}

func ensureNative() bool {
	nativeOnce.Do(func() {
		nativeOK = tryLoadDLL()
		if !nativeOK {
			log.Printf("[native] conductino_core.dll not found — using Go text extract")
		}
	})
	return nativeOK
}

func NativeInit(dataDir string) {
	if !ensureNative() || pInit == nil {
		return
	}
	_ = os.MkdirAll(dataDir, 0o755)
	ptr, err := syscall.BytePtrFromString(dataDir)
	if err != nil {
		return
	}
	r, _, _ := pInit.Call(uintptr(unsafe.Pointer(ptr)))
	if r == 0 {
		log.Printf("[native] conductino_init ok")
	}
}

func NativeDocumentExtract(path string) (string, error) {
	if !ensureNative() || pDocExtract == nil {
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
	if pFree != nil {
		defer pFree.Call(outPtr)
	}
	b := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	return string(b), nil
}
