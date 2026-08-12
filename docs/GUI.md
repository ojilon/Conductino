# Conductino Desktop — GUI Skeleton & Extension Guide

This document describes the **Chrome-like UI** and how to extend it.

The chrome (tabs, toolbar, sidebar, window controls) is pure HTML/CSS/JS.  
Remote pages are loaded by the **native webview** (`webview.Navigate`) — not by feeding HTML into an iframe.

---

## Current skeleton (restructure Step 2)

Implemented in `frontend/web/` + `frontend/main.go`:

| Element | Status |
|---------|--------|
| Title bar + window controls (min / max / close) | UI + Go binds (`hostMinimize`, `hostMaximize`, `hostClose`) |
| Tab strip (new / close / switch) | JS model live |
| Toolbar (back / forward / reload) | UI + Go binds |
| Omnibox (URL or search) | Live; search engines from settings |
| Sidebar (Settings + Downloads/Bookmarks stubs) | Live |
| Themes (Aurora Dark / Light) | Live (`data-theme`) |
| Multi search engines | Live (DuckDuckGo, Google, Bing, Startpage) |
| Local state panels (welcome, settings, stub) | Live in `#content-host` |

---

## Target visual layout

```
┌──────────────────────────────────────────────────────────────┐
│  ◆ Conductino                         [ ─  □  × ]           │  ← title bar
├──────────────────────────────────────────────────────────────┤
│  [Tab] [Tab] [+]                                             │  ← tab strip
├──────────────────────────────────────────────────────────────┤
│  ◀  ▶  ⟳  |  🔒  https://…                         |  ☰    │  ← toolbar
├──────────────┬───────────────────────────────────────────────┤
│  Sidebar     │   Content host / native webview surface       │
│  · Settings  │                                               │
│  · Downloads │                                               │
│  · Bookmarks │                                               │
└──────────────┴───────────────────────────────────────────────┘
```

---

## Content surface (important)

### Rule
Remote pages must be loaded **natively** by the webview (same as Android WebView).  
Do **not** reintroduce a Go HTTP client that fetches a page and injects it into an iframe.

### What webview_go gives us today
One native webview = one window surface. When the chrome calls `hostNavigate(url)`, Go runs `webview.Navigate(url)`. That is correct and native — the browser identity, cookies, Cloudflare challenges, etc. are handled by the platform webview.

**Trade-off:** navigating away replaces the chrome HTML with the remote page.  
`hostShowChrome` navigates back to the local shell.

### Path forward (not blocking the skeleton)
To keep chrome always visible *and* a native content surface:

1. **Preferred later:** multi-webview / panel composition (e.g. Wails WebviewPanel, or platform child WebView2 / WKWebView embedded under the chrome region).
2. **Interim UX:** after a remote load, inject a small floating “← Chrome” control via `Eval` that calls `hostShowChrome`.
3. **Never:** put remote documents in an `<iframe>` inside `index.html` as the primary content path.

Local states (welcome, settings, error, plain-text, downloads list, …) continue to render inside `#content-host` in the chrome document.

---

## File map

```
frontend/
├── main.go              # webview host, binds, static server
├── config.go
├── go.mod
├── pathutil/
└── web/
    ├── index.html       # chrome shell
    ├── css/base.css     # layout + aurora themes
    └── js/app.js        # tabs, nav, sidebar, settings
```

---

## How to add a new sidebar item

1. Add a button in `#sidebar .sidebar-nav` with `data-action="your-id"`.
2. Handle it in `app.js` (sidebar click handler).
3. Either open a local state panel under `#content-host` or call a Go binding.
4. For stubs, reuse `#stub-panel` via `showStub(title, body)`.
5. Document the item here or in a short README next to the feature.

---

## How to add a new local state page

1. Add a `<section id="…" class="state-panel">` in `index.html` (or a folder under `web/states/` later).
2. Show/hide it from `showPanel` in `app.js`.
3. Keep remote pages out of these panels — they go through `hostNavigate`.

---

## Navigation rules

- **Omnibox submit**
  - Looks like URL → normalize and `hostNavigate(url)`.
  - Otherwise → build search URL from the selected engine and navigate.
- **Back / Forward / Reload** → `hostGoBack` / `hostGoForward` / `hostReload` (currently `history.*` / `location.reload` via Eval on the active surface).
- Do not add a Go-side page proxy.

---

## Themes & search engines

- Theme: `data-theme="aurora-dark" | "aurora-light"` on `<html>`, toggled from Settings.
- Search engines: defined in `app.js` (`ENGINES`). Preference in `localStorage` for now; will move to backend settings later.
- Full token table will live in `docs/THEMES.md` when polished.

---

## Window controls

Bound as `hostMinimize`, `hostMaximize`, `hostClose`.  
Close is implemented; min/max log for now and can be wired via the native window handle from `webview.Window()` per platform.

---

## What is deliberately left for later

- Always-visible chrome + native content (multi-webview / panel)
- Tab persistence across restarts
- Real back/forward history flags from the webview
- Downloads / bookmarks implementations
- Find-in-page, reader mode
- C++ backend settings storage

Each should get a short note here or a feature README when work starts.
