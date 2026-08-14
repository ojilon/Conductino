# Wails rewrite — Step 2

**Status:** Chrome shell ported into the Wails asset tree.

## What you should see

```bat
cd frontend
wails dev
```

- Tab strip (+ new / close)
- Toolbar (omnibox, study 📚, sidebar ☰)
- Welcome, Settings, Study split pane
- Sidebar menu items
- Paste text → source pane; Transfer exact → knowledge pane

Omnibox opens URLs via interim helper (system browser / Step 3 `OpenURL`).
Native file dialog + C++ extract + LLM summarize → Steps 3–4.

## Asset map

```
frontend/frontend/
  index.html
  css/base.css
  css/study.css
  js/bridge.js
  js/app.js
  js/study.js
```

## Next — Step 3

Go methods on `App`:

- `OpenURL(url)` — `runtime.BrowserOpenURL` or in-app strategy
- `OpenFile()` — native dialog
- `ExtractDocument(path)` — C++ backend when DLL present

Tell me when chrome looks right (or what’s still broken).
