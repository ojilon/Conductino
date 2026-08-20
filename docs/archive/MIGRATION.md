# Migration from frontend-ui (historical)

> **Archive.** Current UI is React under `frontend/src/`.

## Retired

- Go `/api/proxy` and HTML rewrite for iframe embedding of remote pages
- Go scraping of search-engine HTML for results UI
- Dual native content webview as the primary path

## Current navigation

```
Omnibox → React (URL or search) → tab.url → BrowserView iframe
```

Optional: system browser via **Open external** when framing is blocked.

Do not revive server-side page fetch for display.
