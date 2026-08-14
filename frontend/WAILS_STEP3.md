# Wails rewrite — Step 3

**Status:** Go bindings for URL open, native file dialog, document extract/import.

## Bindings (`App`)

| Method | Behavior |
|--------|----------|
| `OpenURL(url)` | System browser via `runtime.BrowserOpenURL` |
| `OpenFile()` | Native open dialog (txt/md/pdf/docx/…) |
| `ExtractDocument(path)` | C++ `conductino_core` if DLL present, else Go text reader |
| `ImportDocument(path)` | Copy into `conductino-data/imports/` |
| `WindowMinimise` / `WindowToggleMaximise` / `WindowClose` | Window controls |

## Try it

```bat
cd frontend
wails dev
```

1. Omnibox → submit a URL → opens in your default browser.  
2. Study → **Open file…** → pick a `.txt` or `.md` → text appears in the left pane.  
3. Dialog asks Import vs Link (OK = copy into app data).  
4. PDF/DOCX → notice text until C++ PDF lib is linked.

Optional native extract:

```bat
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

Place `conductino_core.dll` on PATH or under `backend/build/Release`.

## Next — Step 4

Wire AI modules (`provider` / `chunker` / `output_parser`) + Summarize & transfer.
