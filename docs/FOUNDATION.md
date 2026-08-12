# Conductino — Foundation Map (Desktop)

**Branch:** `restructure`  
**Status:** Foundation in progress. Chrome skeleton and backend stubs come next.

This document is the single source of truth for the desktop restructure.
It mirrors the spirit of `Conductino-Android/FOUNDATION.md` while adapting
to Go + webview_go + C++23.

---

## Core architectural rules (locked)

| Rule | Detail |
|------|--------|
| **Native webview owns the network** | The webview (via `webview_go`) loads URLs, receives responses, handles redirects, cookies, Cloudflare challenges, CORS, etc. **natively**. No Go-side HTTP client that fetches pages and injects them into an iframe. |
| **No iframe for remote content** | The content surface **is** the native webview. Local UI chrome (tabs, toolbar, sidebar) is HTML/CSS/JS living in the same process, but remote pages are never loaded into an `<iframe>`. |
| **Backend has no network** | `backend/` is almost purely C++23 + CMake. Storage, document processing, notes, history, bookmarks, settings persistence, heavy CPU work. Zero active/direct network I/O. |
| **Go helpers are thin** | Go exists to host the webview, serve the chrome UI, and provide a low-cost bridge to the C++23 core. It must not become a second networking stack. |
| **Local state pages are allowed** | Loading, error, plain-text viewer, search-results UI, settings, downloads list, etc. can still be local HTML/CSS/JS served by the Go host or loaded into the webview as `data:` / local HTTP. They are **not** remote pages. |
| **Post-load interaction only** | Extracting text, injecting scripts, highlighting, reader mode, etc. happen **after** the page has been natively loaded and is visible to the user. |

These rules exist so the browser is identified as a normal human-driven browser
(Cloudflare, anti-bot systems, etc.) and so the architecture stays clean.

---

## Target project layout

```
Conductino/
├── frontend/                 # Go + webview_go + vanilla HTML/CSS/JS
│   ├── main.go
│   ├── go.mod
│   ├── config.go / pathutil/
│   ├── bridge/                # Go ↔ JS and Go ↔ C++ helpers
│   ├── handlers/              # Only non-network logic (settings, notes, local states)
│   ├── web/
│   │   ├── index.html          # Chrome-like shell (tabs, toolbar, sidebar)
│   │   ├── css/
│   │   ├── js/
│   │   └── states/             # local loading / error / plain-text / search / …
│   └── …
├── backend/                  # C++23 + CMake (primary native core)
│   ├── CMakeLists.txt
│   ├── include/
│   ├── src/
│   ├── features/              # tabs, storage, document, … (stubs + READMEs)
│   └── third_party/
├── backend-core/             # Zig experiments (kept; not the main path for now)
├── config/                   # search_engines.json, settings schema, themes
├── docs/                     # this file, GUI.md, THEMES.md, …
├── frontend-ui/              # OLD working tree — will be removed by owner after migration
└── …
```

---

## Implemented so far (restructure)

| Area | Status | Location / notes |
|------|--------|------------------|
| Architecture rules | Done | this file |
| Docs skeleton | Done | `docs/FOUNDATION.md`, `docs/GUI.md` |
| Frontend chrome | Next | tabs, toolbar, URL bar, sidebar, window controls |
| Native webview content surface | Next | replace any iframe-based loading |
| Theme (dark / light) | Planned | settings + CSS variables |
| Multi search engines | Planned | config + settings UI |
| Backend C++23 skeleton | Planned | `backend/` |
| Thin Go ↔ C++ bridge | Planned | |
| Migration of useful `frontend-ui` logic | Planned | settings, notes, local states only |

---

## Roadmap (ordered)

1. **Foundation docs** ✅ (this commit)
2. **GUI skeleton** — Chrome-like shell in HTML/CSS/JS + webview_go hosting the native content surface
3. **Bridge & tab model** — Go ↔ JS, tab management, navigation controls that drive the native webview
4. **Clean old handlers** — keep only non-network pieces; delete proxy / fetch-and-inject approach
5. **Backend C++23 + CMake** — feature stubs + READMEs
6. **Settings, themes, search engines** + remaining stubs + extension docs
7. **Owner removes `frontend-ui/`** when ready

---

## How to extend (quick pointers)

- **New sidebar item** → see `docs/GUI.md`
- **New local state page** (error, plain-text, …) → `frontend/web/states/` + short README
- **New backend feature** → `backend/features/<name>/` with a README describing purpose and recommended libraries
- **Theme tokens** → `docs/THEMES.md` (to be added with the theme work)
- **Search engine** → config + Settings UI (documented when implemented)

Always add or update a small README next to new feature directories.  
Document-as-you-go is the convention (same as Android).

---

## Relationship to Conductino-Android

The desktop version is the PC counterpart of the Android app.  
Concepts that transfer directly:

- State-driven local UIs
- Native webview as the content surface
- Settings (theme, search engines)
- Tabs, history, bookmarks, downloads as first-class modules
- Native core for heavy / reusable logic
- Feature directories with READMEs

Differences are mainly technology (Go + webview_go + C++23 vs Java + Android WebView + C) and the absence of any Go-side page fetching.
