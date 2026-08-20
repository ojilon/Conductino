# Building Conductino

**Primary path:** Wails v2 + React / Vite / Tailwind + optional C++ backend.

## Prerequisites (Windows)

- Go 1.22+
- Node.js 18+ (npm)
- WebView2 runtime (usually installed with Edge)
- [Wails CLI](https://wails.io) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Optional: CMake 3.25+ for `conductino_core` (PDF/text extract cache)

## Development

```bat
cd frontend
npm install
go mod tidy
wails dev
```

Wails runs Vite (`frontend:dev:watcher`) and serves the React app. Config: `frontend/wails.json` (`frontend:dir` = `.`, `assetdir` = `dist`).

## Production binary

```bat
cd frontend
wails build
```

Runs `npm run build` into `dist/`, embeds it, outputs under `frontend/build/bin/` (**Conductino**).

## Optional C++ backend

```bat
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --config Release
```

DLL search order: `CONDUCTINO_CORE_DIR`, next to the exe, `backend/build`, `backend/build/Release`.

Without the DLL: Go still reads `.txt` / `.md` / etc.; PDF shows a notice. See [docs/concepts/TEXT_EXTRACTION.md](docs/concepts/TEXT_EXTRACTION.md).

## AI keys

Settings → AI providers → paste key → **Save AI config** (localStorage only).

## Retired paths

Do not use for new work:

- `frontend/web/` — legacy vanilla UI
- `frontend/frontend/` — nested asset dir
- Dual-webview notes in `docs/archive/`
