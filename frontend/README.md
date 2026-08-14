# Conductino frontend (Wails)

Desktop study browser UI powered by [Wails v2](https://wails.io).

## Quick start

```bat
go mod tidy
wails dev
```

## Layout

| Path | Role |
|------|------|
| `main.go` | `wails.Run`, embed assets |
| `app.go` | Bound Go API (OpenURL, OpenFile, ExtractDocument, …) |
| `native_windows.go` | Optional load of `conductino_core.dll` |
| `frontend/` | UI assets (index, css, js, ai modules) |
| `wails.json` | Wails project config |

## Features available now

- Tabs, omnibox (system browser), sidebar, themes
- Study split view (draggable splitter)
- Open/paste documents, transfer exact text
- AI summarize with local chunking + provider failover
- Export knowledge pane as Markdown

## Docs

- Repo root `BUILDING.md`
- `docs/ai/ARCHITECTURE.md` — long-term AI blueprints
- `docs/ai/PROMPTS.md` — system prompt templates
- `RETIRED_DUAL_WEBVIEW.md` — old shell (do not use)
