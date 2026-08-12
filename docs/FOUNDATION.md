# Conductino — Foundation Map (Desktop)

**Branch:** `restructure`  
**Status:** GUI skeleton in place. Next: harden navigation chrome-return, then backend C++23 stubs.

This document is the single source of truth for the desktop restructure.
It mirrors the spirit of `Conductino-Android/FOUNDATION.md` while adapting
to Go + webview_go + C++23.

---

## Core architectural rules (locked)

| Rule | Detail |
|------|--------|
| **Native webview owns the network** | The webview (via `webview_go`) loads URLs, receives responses, handles redirects, cookies, Cloudflare challenges, CORS, etc. **natively**. No Go-side HTTP client that fetches pages and injects them into an iframe. |
| **No iframe for remote content** | Remote pages are never the responsibility of an `<iframe>` inside the chrome document. |
| **Backend has no network** | `backend/` is almost purely C++23 + CMake. Storage, document processing, notes, history, bookmarks, settings persistence, heavy CPU work. Zero active/direct network I/O. |
| **Go helpers are thin** | Go hosts the webview, serves the chrome UI, and bridges to C++23. It must not become a second networking stack. |
| **Local state pages are allowed** | Welcome, settings, error, plain-text, search UI, downloads list, etc. render in the chrome document (`#content-host`). |
| **Post-load interaction only** | Extracting text, injecting scripts, highlighting, reader mode, etc. happen **after** the page has been natively loaded. |

---

## Target project layout

```
Conductino/
├── frontend/                 # Go + webview_go + vanilla HTML/CSS/JS
│   ├── main.go
│   ├── go.mod
│   ├── config.go / pathutil/
│   ├── bridge/               # (later) Go ↔ C++ helpers
│   └── web/
│       ├── index.html        # Chrome-like shell
│       ├── css/
│       └── js/
├── backend/                  # C++23 + CMake (primary native core)
├── backend-core/             # Zig experiments (kept; not the main path for now)
├── config/                   # search engines, settings schema (later)
├── docs/
└── frontend-ui/              # OLD tree — owner removes after migration
```

---

## Implemented so far (restructure)

| Area | Status | Location |
|------|--------|----------|
| Architecture rules | Done | this file |
| Docs skeleton | Done | `docs/` |
| Frontend chrome UI | Done | `frontend/web/` |
| Themes + search engines (UI) | Done | Settings panel + `app.js` |
| Native `hostNavigate` | Done | `frontend/main.go` |
| Chrome always-visible + remote | Interim | single webview trade-off — see `docs/GUI.md` |
| Backend C++23 skeleton | Planned | `backend/` |
| Migration of non-network `frontend-ui` logic | Planned | |

---

## Roadmap (ordered)

1. Foundation docs ✅
2. GUI skeleton ✅
3. Navigation UX (return-to-chrome / history meta) — next small slice
4. Clean old handlers (drop proxy path)
5. Backend C++23 + CMake feature stubs
6. Persistent settings via backend
7. Owner removes `frontend-ui/`

---

## How to extend

- Sidebar / local states / themes → `docs/GUI.md`
- New backend feature → `backend/features/<name>/` + README
- Always document-as-you-go

---

## Relationship to Conductino-Android

Same product ideas: native content surface, state-driven local UIs, settings (theme, search engines), tabs, modular features with READMEs.  
Different stack: Go + webview_go + C++23 vs Java + Android WebView + C.
