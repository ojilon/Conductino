# Bridge (Go ↔ frontend)

## Wails bindings (`App`)

Exposed from `frontend/app.go`:

| Method | Role |
|--------|------|
| `OpenFile` | Native open dialog |
| `ExtractDocument` | Local path → text |
| `ImportDocument` | Copy into app data |
| `Greet` / `AppInfo` | Smoke / diagnostics |

Generated JS: `frontend/wailsjs/go/main/`.

## Wails runtime (no custom Go)

From `frontend/wailsjs/runtime`:

- `WindowMinimise`, `WindowToggleMaximise`, `Quit`
- `BrowserOpenURL` — **fallback** when iframe cannot embed a site

Wrapper: `frontend/src/lib/wails.ts`.

## Navigation

In-app navigation is **not** a Go `Navigate` call. React sets the tab URL and mounts `BrowserView` (iframe).
