# Conductino Desktop — GUI Skeleton & Extension Guide

This document describes the **Chrome-like UI** we are building and how to extend it.

The chrome (tabs, toolbar, sidebar, window controls) is pure HTML/CSS/JS.  
The **content area is the native webview** — never an iframe that loads remote sites.

---

## Target layout (visual)

```
┌───────────────────────────────────────────────────────────────────────┐
│  [tabs strip]                                    [ –  □  × ]  │  ← window controls
├───────────────────────────────────────────────────────────────────────┤
│  ◀  ▶  ↻  |  🔒  https://…                    |  ☰  │  ← toolbar + URL bar
├───────────────────────────────────────────────────────────────────────┤
│  Sidebar (optional)  │                                     │
│  · Settings          │   NATIVE WEBVIEW                    │
│  · Downloads (stub)  │   (content surface)                 │
│  · Bookmarks (stub)  │                                     │
│  · …                 │                                     │
└───────────────────────────────────────────────────────────────────────┘
```

### Elements we will implement in the skeleton

| Element | Responsibility | Notes |
|---------|----------------|-------|
| **Window controls** | Minimize / Maximize (or restore) / Close | Top-right. On platforms where the OS draws them we may hide our own. |
| **Tab strip** | Multiple tabs, new tab, close tab, switch | Basic model first; persistence later. |
| **Toolbar** | Back, Forward, Refresh | Drive the **native** webview navigation. |
| **URL / omnibox** | Type URL or search query | Search engines come from settings. |
| **Sidebar** | Settings (implemented), Downloads / Bookmarks (stubs) | Collapsible. |
| **Content area** | Native webview | The only place remote pages appear. |
| **Theme** | Dark / Light | CSS variables + settings. |

Undone pieces will have short TODOs and a pointer to this file or a feature README.

---

## File map (planned)

```
frontend/web/
├── index.html          # shell: tabs + toolbar + sidebar + content host
├── css/
│   ├── base.css
│   ├── toolbar.css
│   ├── tabs.css
│   ├── sidebar.css
│   ├── themes.css       # dark / light tokens
│   └── …
├── js/
│   ├── bridge.js        # talk to Go (and later C++)
│   ├── tabs.js
│   ├── navigation.js    # back / forward / reload / loadURL
│   ├── sidebar.js
│   ├── settings.js
│   ├── state.js
│   └── app.js           # bootstrap
└── states/              # local pages (not remote)
    ├── loading/
    ├── error/
    ├── plain_text/
    ├── search/
    └── …
```

Exact names may shift slightly while we implement; keep this doc in sync.

---

## How to add a new sidebar item

1. Add a button / entry in the sidebar markup (or generate it from a list in JS).
2. Give it a clear `data-action` or `id`.
3. In `sidebar.js` (or equivalent) handle the click:
   - For **Settings** → open the settings panel / state.
   - For a **stub** (Downloads, Bookmarks, …) → show a placeholder local page or toast "Coming soon".
4. Document the new item in this file or in a short `frontend/web/README.md`.

Only Settings needs to be fully functional in the first pass; one or two other entries can be stubs so the shape is visible.

---

## How to add a new local state page

Local states are for things that are **not** remote websites (error screens, plain-text viewer, search UI, downloads list, etc.).

1. Create `frontend/web/states/<name>/` with `index.html` (+ optional `style.css` / `logic.js`).
2. Register the state name in the Go host or in the JS state machine so it can be loaded into the content area (or a dedicated panel) when needed.
3. Add a one-paragraph README inside the state folder describing when it is shown and what data it expects.
4. Never use these folders for ordinary remote pages — those go through the native webview.

---

## Navigation rules (important)

- **Back / Forward / Refresh** must call the native webview APIs (via the Go bridge), not manipulate an iframe.
- **URL bar submit**:
  - If the input looks like a URL → `webview.Navigate(url)` (or equivalent).
  - If it looks like a search query → build a search URL from the **currently selected search engine** and navigate.
- Do **not** re-introduce a Go HTTP client that fetches the page and writes HTML into the DOM.

---

## Themes

- CSS custom properties for colors, radii, fonts.
- Two built-in themes: `dark` and `light` (names can be `aurora-dark` / `aurora-light` to stay close to Android).
- Preference stored in settings; applied by setting `data-theme` on `<html>` or a root class.
- Full token list will live in `docs/THEMES.md` once the CSS is written.

---

## Window controls

- Implement minimize / maximize-or-restore / close in the top-right of the chrome.
- On Windows with WebView2 these can map to the host window methods exposed by `webview_go`.
- On other platforms follow the same visual language; hide or adapt if the OS already draws the buttons.

---

## What is deliberately left for later

- Full tab persistence across restarts
- Drag-and-drop tab reordering
- Find-in-page chrome
- Reader mode / advanced content extraction UI
- Full downloads manager UI
- Bookmarks manager UI

Each of the above should get a short section or a feature README when work starts on it.

---

## Relation to the old `frontend-ui`

The old tree used an iframe (`#webview`) and a lot of Go-side request handling.  
That approach is **retired**. Useful non-network pieces (plain-text viewer, notes, local styling ideas) will be migrated carefully; the network path will not.
