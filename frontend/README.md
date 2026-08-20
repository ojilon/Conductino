# Conductino frontend (Wails + React + Tailwind)

Desktop study browser UI powered by [Wails v2](https://wails.io), React, Vite, and Tailwind CSS.

## Quick start

```bat
cd frontend
npm install
go mod tidy
wails dev
```

`wails dev` runs the Vite watcher and serves the React app. Production assets land in `dist/` via `npm run build`.

## Layout

| Path | Role |
|------|------|
| `main.go` | `wails.Run`, embed `dist/` |
| `app.go` | Bound Go API: OpenFile, ExtractDocument, ImportDocument, Greet, AppInfo |
| `src/` | React UI (components, AI lib, Tailwind) |
| `wailsjs/` | Generated JS bindings + runtime |
| `wails.json` | Vite install/build/dev + assetdir `dist` |
| `web/` | **Legacy** vanilla HTML/JS (unused; safe to delete later) |
| `frontend/frontend/` | **Legacy** nested asset dir (unused) |

## What Wails handles (no custom Go)

- Window minimise / maximise / quit → `wailsjs/runtime`
- Open URL in system browser → `BrowserOpenURL`

## What stays in Go

- Native file dialog (`OpenFile`)
- Document text extract / import (optional C++ DLL + Go fallback)

## Features

- Tabs, omnibox (system browser), sidebar, themes
- Study split view, open/paste documents, exact transfer
- AI summarize with local chunking + provider failover (keys in localStorage)
- Export knowledge pane as Markdown
