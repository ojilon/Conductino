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
| `main.go` | `wails.Run`, embed assets from `web/` |
| `app.go` | Bound Go API (OpenURL, OpenFile, ExtractDocument, …) |
| `web/` | UI assets (index.html, css, js, ai modules) — **assetdir** |
| `wailsjs/` | Generated JS/TS bindings (wailsjsdir = `.`) |
| `wails.json` | Wails project config |

The nested `frontend/frontend/` directory is legacy and is no longer used for assets or bindings.

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
