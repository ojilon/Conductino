# Building Conductino

**Primary path (current):** Wails v2 + vanilla HTML/CSS/JS + optional C++ backend.

## Prerequisites (Windows)

- Go 1.22+
- WebView2 runtime (usually already installed with Edge)
- Optional: [Wails CLI](https://wails.io) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Optional: CMake 3.25+ for the C++ `conductino_core` library (PDF/text extract cache)

## Run (development)

```bat
cd frontend
go mod tidy
wails dev
```

Or without the CLI:

```bat
cd frontend
go mod tidy
go run .
```

You should get a normal OS window with tabs, Settings, and Study workspace.

## Build a binary

```bat
cd frontend
wails build
```

Output is under `frontend/build/bin/` (name from `wails.json`: **Conductino**).

## Optional C++ backend

Used for richer document extract/cache when `conductino_core.dll` is found:

```bat
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

Search order for the DLL: `CONDUCTINO_CORE_DIR`, next to the exe, `backend/build`, `backend/build/Release`.

Without the DLL, the app still runs: Go reads `.txt` / `.md` / etc.; PDF shows a clear notice.

## AI keys

Settings → AI providers → paste key → **Save AI config** (replaces any previous key).  
Default Google model: `gemini-2.5-flash`.

## Retired path

The old `webview_go` dual-surface layout under `frontend/bridge/`, `frontend/shell/`, and `frontend/web/` is **retired**. Do not use it for new work. See `frontend/RETIRED_DUAL_WEBVIEW.md`.
