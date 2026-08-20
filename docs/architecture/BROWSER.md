# Browser / navigation model

## In-app content (default)

The omnibox resolves input to a URL (literal or search-engine template) and loads it in an **iframe** under the chrome (`BrowserView`). Tabs store `{ id, title, url }`.

- Chrome (tabs, toolbar, right sidebar) stays in the React tree.
- Single Wails webview hosts both chrome and the iframe.

### Limits

Many sites send `X-Frame-Options` or CSP `frame-ancestors` and will not render in an iframe. Use **Open external** (system browser via Wails `BrowserOpenURL`) for those.

Longer-term options (not required for this merge):

- Native child webview for content only (complex on Windows; dual-webview was retired once).
- Readability mode: fetch is **not** done in Go for display (backend stays offline); any future proxy must be explicit and documented.

## What Go does *not* do

- Download HTML to rewrite into an iframe (old `/api/proxy` — retired).
- Own a second full browser process unless we reintroduce a dedicated content surface.

## Related

- [GUI.md](GUI.md) — panels and chrome
- [BRIDGE.md](BRIDGE.md) — bindings vs runtime
