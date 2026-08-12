package bridge

// Native* API toward conductino_core (C ABI in backend/include/conductino/core.h).
// Platform files:
//   native_windows.go  — LoadLibrary (default on Windows)
//   native_cgo.go      — optional //go:build conductino_cgo
//   native_stub.go     — other OS / no library
//
// See NATIVE.md for build steps.

type nativeError string

func (e nativeError) Error() string { return string(e) }

const errNativeUnavailable = nativeError("conductino_core not linked")
