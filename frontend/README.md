# Conductino frontend (Wails + React + Tailwind)

## Quick start

```bat
npm install
go mod tidy
wails dev
```

## Layout

| Path | Role |
|------|------|
| `main.go` | `wails.Run`, embed `dist/` |
| `app.go` | OpenFile, ExtractDocument, ImportDocument, … |
| `src/` | React UI |
| `src/components/BrowserView.tsx` | In-app iframe content |
| `src/lib/ai/` | Provider, chunker, output parser |
| `wailsjs/` | Generated bindings + runtime |
| `wails.json` | `frontend:dir` = `.`, Vite hooks, `assetdir` = `dist` |

## Legacy (unused)

- `web/` — old vanilla UI
- `frontend/` nested folder — do not use

Delete when convenient after merge.
