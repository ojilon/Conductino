//go:build windows

package bridge

import (
	"syscall"
	"unsafe"
)

// Minimal GetOpenFileNameW wrapper (comdlg32).

type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

const (
	ofnExplorer  = 0x00080000
	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800
	ofnHidereadonly = 0x00000004
)

func OpenFileDialog() (string, error) {
	comdlg := syscall.NewLazyDLL("comdlg32.dll")
	proc := comdlg.NewProc("GetOpenFileNameW")

	buf := make([]uint16, 32768)
	filter, _ := syscall.UTF16PtrFromString("Documents\x00*.txt;*.md;*.pdf;*.docx;*.json;*.csv\x00All\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Open document")

	of := openFileNameW{
		lStructSize:  uint32(unsafe.Sizeof(openFileNameW{})),
		lpstrFilter:  filter,
		nFilterIndex: 1,
		lpstrFile:    &buf[0],
		nMaxFile:     uint32(len(buf)),
		lpstrTitle:   title,
		flags:        ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnHidereadonly,
	}

	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&of)))
	if r == 0 {
		return "", nil // cancelled
	}
	return syscall.UTF16ToString(buf), nil
}
