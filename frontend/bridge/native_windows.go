//go:build windows && !conductino_cgo

package bridge

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

// Dynamic load of conductino_core.dll (C ABI). No MinGW required.

var (
	nativeOnce sync.Once
	nativeOK   bool
	nativeDll  *syscall.LazyDLL

	pVersion   *syscall.LazyProc
	pInit      *syscall.LazyProc
	pShutdown  *syscall.LazyProc
	pNotesSave *syscall.LazyProc
	pNotesSearch *syscall.LazyProc
	pFree      *syscall.LazyProc
	pSettingsGet *syscall.LazyProc
	pSettingsSet *syscall.LazyProc
)

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
			// Load with full path so dependent CRT can resolve.
			h, err := syscall.LoadLibrary(path)
			if err != nil {
				log.Printf("[native] LoadLibrary %s: %v", path, err)
				continue
			}
			_ = h
			nativeDll = syscall.NewLazyDLL(path)
			if err := nativeDll.Load(); err != nil {
				log.Printf("[native] lazy load %s: %v", path, err)
				continue
			}
			pVersion = nativeDll.NewProc("conductino_version")
			pInit = nativeDll.NewProc("conductino_init")
			pShutdown = nativeDll.NewProc("conductino_shutdown")
			pNotesSave = nativeDll.NewProc("conductino_notes_save_json")
			pNotesSearch = nativeDll.NewProc("conductino_notes_search")
			pFree = nativeDll.NewProc("conductino_free")
			pSettingsGet = nativeDll.NewProc("conductino_settings_get")
			pSettingsSet = nativeDll.NewProc("conductino_settings_set")
			// Probe version export.
			r, _, err := pVersion.Call()
			if r == 0 {
				log.Printf("[native] conductino_version missing in %s (%v)", path, err)
				continue
			}
			ver := windowsCString(r)
			log.Printf("[native] loaded %s (%s)", path, ver)
			return true
		}
	}
	// Also try bare name on PATH.
	nativeDll = syscall.NewLazyDLL("conductino_core.dll")
	if err := nativeDll.Load(); err == nil {
		pVersion = nativeDll.NewProc("conductino_version")
		pInit = nativeDll.NewProc("conductino_init")
		pShutdown = nativeDll.NewProc("conductino_shutdown")
		pNotesSave = nativeDll.NewProc("conductino_notes_save_json")
		pNotesSearch = nativeDll.NewProc("conductino_notes_search")
		pFree = nativeDll.NewProc("conductino_free")
		pSettingsGet = nativeDll.NewProc("conductino_settings_get")
		pSettingsSet = nativeDll.NewProc("conductino_settings_set")
		r, _, _ := pVersion.Call()
		if r != 0 {
			log.Printf("[native] loaded conductino_core.dll from PATH (%s)", windowsCString(r))
			return true
		}
	}
	return false
}

func ensureNative() bool {
	nativeOnce.Do(func() {
		nativeOK = tryLoadDLL()
		if !nativeOK {
			log.Printf("[native] conductino_core.dll not found — build backend and place DLL on PATH or in backend/build")
		}
	})
	return nativeOK
}

func windowsCString(p uintptr) string {
	if p == 0 {
		return ""
	}
	var b []byte
	for i := 0; ; i++ {
		c := *(*byte)(unsafe.Pointer(p + uintptr(i)))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

func NativeAvailable() bool {
	return ensureNative()
}

func NativeVersion() string {
	if !ensureNative() {
		return ""
	}
	r, _, _ := pVersion.Call()
	return windowsCString(r)
}

func NativeInit(dataDir string) error {
	if !ensureNative() {
		log.Printf("[native] conductino_core not linked; using Go NoteStore")
		return nil
	}
	_ = os.MkdirAll(dataDir, 0o755)
	ptr, err := syscall.BytePtrFromString(dataDir)
	if err != nil {
		return err
	}
	r, _, callErr := pInit.Call(uintptr(unsafe.Pointer(ptr)))
	if r != 0 {
		return fmt.Errorf("conductino_init returned %d (%v)", r, callErr)
	}
	log.Printf("[native] conductino_init(%s) ok version=%s", dataDir, NativeVersion())
	return nil
}

func NativeShutdown() {
	if !ensureNative() {
		return
	}
	pShutdown.Call()
}

func NativeNotesSaveJSON(json string) error {
	if !ensureNative() {
		return errNativeUnavailable
	}
	ptr, err := syscall.BytePtrFromString(json)
	if err != nil {
		return err
	}
	r, _, callErr := pNotesSave.Call(uintptr(unsafe.Pointer(ptr)), uintptr(len(json)))
	if r != 0 {
		return fmt.Errorf("notes_save_json: %d (%v)", r, callErr)
	}
	return nil
}

func NativeNotesSearch(query string) (string, error) {
	if !ensureNative() {
		return "", errNativeUnavailable
	}
	qptr, err := syscall.BytePtrFromString(query)
	if err != nil {
		return "", err
	}
	var outPtr uintptr
	var outLen uintptr
	r, _, callErr := pNotesSearch.Call(
		uintptr(unsafe.Pointer(qptr)),
		uintptr(unsafe.Pointer(&outPtr)),
		uintptr(unsafe.Pointer(&outLen)),
	)
	if r != 0 {
		return "", fmt.Errorf("notes_search: %d (%v)", r, callErr)
	}
	if outPtr == 0 {
		return "[]", nil
	}
	defer pFree.Call(outPtr)
	b := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	return string(b), nil
}

func NativeSettingsGet(key string) (string, error) {
	if !ensureNative() {
		return "", errNativeUnavailable
	}
	kptr, err := syscall.BytePtrFromString(key)
	if err != nil {
		return "", err
	}
	var outPtr uintptr
	var outLen uintptr
	r, _, callErr := pSettingsGet.Call(
		uintptr(unsafe.Pointer(kptr)),
		uintptr(unsafe.Pointer(&outPtr)),
		uintptr(unsafe.Pointer(&outLen)),
	)
	if int(r) < 0 {
		return "", fmt.Errorf("settings_get: %d (%v)", r, callErr)
	}
	if r == 1 {
		return "", nil // missing
	}
	if outPtr == 0 {
		return "", nil
	}
	defer pFree.Call(outPtr)
	b := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), outLen)
	return string(b), nil
}

func NativeSettingsSet(key, value string) error {
	if !ensureNative() {
		return errNativeUnavailable
	}
	kptr, err := syscall.BytePtrFromString(key)
	if err != nil {
		return err
	}
	vptr, err := syscall.BytePtrFromString(value)
	if err != nil {
		return err
	}
	r, _, callErr := pSettingsSet.Call(uintptr(unsafe.Pointer(kptr)), uintptr(unsafe.Pointer(vptr)))
	if r != 0 {
		return fmt.Errorf("settings_set: %d (%v)", r, callErr)
	}
	return nil
}
