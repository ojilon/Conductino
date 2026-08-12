# Linking `conductino_core`

The Go frontend talks to the C++ backend through the C ABI in  
`backend/include/conductino/core.h`.

## Default on Windows (recommended)

Uses **LoadLibrary** — no MinGW/cgo required.

1. Build the DLL:

```bat
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

2. Ensure `conductino_core.dll` is found. Search order:

- `CONDUCTINO_CORE_DIR` env
- next to the exe
- `cwd`
- `backend/build`, `backend/build/Release`, `backend/build/Debug`
- system `PATH`

3. Run the frontend as usual:

```bat
cd frontend
go run .
```

Log on success:

```
[native] loaded ...\conductino_core.dll (0.1.0-restructure)
[native] conductino_init(...) ok
```

If the DLL is missing, the app still runs with the in-memory Go `NoteStore`.

## Optional: real cgo link

```bash
# after building libconductino_core
cd frontend
go build -tags conductino_cgo .
```

Requires a C toolchain that can link the C++ library (`-lstdc++` / MSVC).

## API used from Go

| Go | C |
|----|---|
| `NativeInit` | `conductino_init` |
| `NativeShutdown` | `conductino_shutdown` |
| `NativeNotesSaveJSON` | `conductino_notes_save_json` |
| `NativeNotesSearch` | `conductino_notes_search` |
| `NativeSettingsGet/Set` | `conductino_settings_*` |

`handlers.NoteStore` already prefers native when `NativeAvailable()` is true.
