# Foundation

## Rules

1. **Offline core** — `backend/` and Go extract paths must not depend on the network for primary features.
2. **One Wails window** — React owns chrome; content is iframe or local panels.
3. **Secrets in the webview** — API keys stay in `localStorage`; not written to C++ data dir by default.
4. **Bindings stay thin** — file dialog, extract, import; window/URL helpers use Wails runtime where possible.

## Layout

```
Conductino/
  frontend/          Wails project root (go.mod, main.go, src/, dist/)
  backend/           Optional C++ conductino_core
  docs/              This tree
  tools/             Scripts
```

## Roadmap pointers

- Text extraction depth → [concepts/TEXT_EXTRACTION.md](../concepts/TEXT_EXTRACTION.md)
- AI schemas → [ai/ARCHITECTURE.md](../ai/ARCHITECTURE.md)
- Product backlog → [product/FEATURES_ROADMAP.md](../product/FEATURES_ROADMAP.md)
