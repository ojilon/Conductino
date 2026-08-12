# Conductino Desktop — Backend (C++23)

See also `backend/README.md` and `frontend/bridge/NATIVE.md`.

## Role

| Layer | Responsibility |
|-------|----------------|
| `frontend/` | Window, chrome, native page loads, thin IPC |
| `backend/` | Notes, settings, storage — **no browsing network** |

## Public C ABI

`backend/include/conductino/core.h`

## Linking from Go

**Windows default:** dynamic `LoadLibrary` of `conductino_core.dll`  
(search `backend/build`, `PATH`, `CONDUCTINO_CORE_DIR`).

**Optional:** `go build -tags conductino_cgo` for static cgo link.

```bat
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release

cd ..\frontend
go run .
```

Expect: `[native] loaded ... conductino_core.dll` then `[native] conductino_init ok`.

## Build demo

```bash
cd backend && cmake -B build && cmake --build build
./build/conductino_backend_demo
```
