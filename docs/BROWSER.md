# Dual webview browser (Windows)

## Layout

```
┌─────────────────────────────────────────┐
│ Tabs + omnibox + tools  (Wails chrome)  │  ← always visible
├─────────────────────────────────────────┤
│                                         │
│   Content WebView2 (child HWND)         │  ← real sites, Cloudflare, etc.
│                                         │
└─────────────────────────────────────────┘
```

- **Chrome** never navigates away for browsing.
- **Content pane** is a second WebView2 embedded under the chrome height (~82px).
- Opening Study / Library **hides** the content pane so local UI fills the workspace.

## Tools (chrome toolbar)

| Button | Action |
|--------|--------|
| ✨ | Copy selection from content page → Study + summarize |
| →📚 | Selection → Study |
| ☆ | Bookmark current omnibox URL |
| 📚 | Open Study (hides content pane) |

## Implementation notes

- `frontend/content_webview_windows.go` — child host HWND + `go-webview2` Chromium
- `ContentNavigate` / `ContentSetVisible` / `ContentResize` App bindings
- Non-Windows builds fall back to full-window navigate (`content_webview_stub.go`)

## If content pane fails to create

Check console for `[content] WebView2 ready`. HWND discovery uses the window title containing `Conductino`. Ensure WebView2 runtime is installed.
