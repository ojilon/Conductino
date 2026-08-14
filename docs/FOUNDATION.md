# Conductino — Foundation Map (Desktop)

**Status:** Active UI path is **Wails v2** (`frontend/main.go` + `frontend/frontend/`).  
The older `webview_go` dual-surface experiment is retired (see `frontend/RETIRED_DUAL_WEBVIEW.md`).

---

## Core architectural rules (locked)

| Rule | Detail |
|------|--------|
| **Wails owns the window** | One native webview surface for chrome + local study UI. |
| **Backend has no network** | `backend/` is C++23 + CMake: storage, document processing, notes, settings. Zero network I/O. |
| **LLM calls in JS** | `fetch` from the webview; API-agnostic adapter; local chunking. |
| **External pages** | Omnibox currently opens the system browser (`OpenURL`). In-app browsing can be added later without changing study tools. |

---

## Project layout (active)

```
Conductino/
├── frontend/                 # Wails app
│   ├── main.go, app.go
│   ├── wails.json
│   └── frontend/             # HTML/CSS/JS assets + ai/
├── backend/                  # C++23 core (optional DLL)
├── docs/
│   └── ai/                   # Architecture + prompt blueprints
└── BUILDING.md
```

---

## Roadmap (high level)

1. Wails shell + study workspace ✅
2. File open / extract / AI summarize ✅
3. In-app content surface (optional multi-webview or browser component)
4. PDF library via CMake `third_party`
5. Verification loop, vision alt-text, TTS (see `docs/ai/`)
6. Bookmarks / downloads / phone-PC sync

---

## How to extend

- New Go binding → method on `App` in `frontend/app.go`, call from `frontend/frontend/js/bridge.js`
- New UI panel → section in `index.html` + `showPanel` in `app.js`
- AI behaviour → `frontend/frontend/js/ai/*` and `docs/ai/PROMPTS.md`
