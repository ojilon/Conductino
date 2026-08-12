# Conductino Desktop — Backend (C++23)

**Step 5.** See also `backend/README.md`.

## Role

| Layer | Responsibility |
|-------|----------------|
| `frontend/` (Go + webview) | Window, chrome, **native page loads**, thin IPC |
| `backend/` (C++23) | Notes, settings, storage, document tools — **no browsing network** |

## Public API

C header: `backend/include/conductino/core.h`

- `conductino_init` / `conductino_shutdown`
- `conductino_notes_save_json` / `conductino_notes_search`
- `conductino_settings_get` / `conductino_settings_set`
- `conductino_free` / `conductino_version`

## Go bridge

`frontend/bridge/native.go` — currently a stub. After building `libconductino_core`, enable cgo and call the C API. Until then `handlers.NoteStore` remains the runtime store.

## Build

```bash
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

## Adding a feature

1. Create `backend/features/<name>/README.md`.
2. Implement in `src/` (or a new .cpp listed in `CMakeLists.txt`).
3. Expose only via `core.h` if Go must call it.
4. Never add HTTP client code for loading websites.
