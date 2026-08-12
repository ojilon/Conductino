# Migration from `frontend-ui` (Step 4)

This document records what was **kept**, what was **retired**, and where things live now.

The old tree (`frontend-ui/`) remains until you delete it. Do not wire new code to it.

---

## Retired (do not port)

| Old path | Why retired |
|----------|-------------|
| `handlers/api_proxy.go` (`/api/proxy`) | Fetched remote pages in Go and rewrote HTML for an iframe. Violates the native-webview rule. |
| `handlers/Plain_text_handler.go` network fetch | Downloaded search-engine HTML in Go to parse results. Network I/O for pages belongs in the webview. |
| `handlers/navigation_helper/rewrite.go` | Only existed to support proxy embedding. |
| `handlers/navigation_helper/browser.go` as page fetcher | Custom HTTP client for page loads is replaced by native Navigate. |
| iframe `#webview` in old `index.html` | Content surface is the native webview window. |

Hitting `/api/proxy`, `/api/plain_text`, or `/api/navigate` on the new server returns **410 Gone** with a pointer here.

---

## Kept / re-homed

| Concern | Old | New |
|---------|-----|-----|
| Note / highlight wire types | `handlers/api.go` `NoteHiglightEvent` | `frontend/handlers/notes.go` `NoteHighlightEvent` |
| Save note | `SaveNoteHandler` → Zig forward | `POST /api/notes` in-memory store (C++ backend later) |
| Search notes | `SearchHandler` → Zig | `GET /api/notes?query=` |
| URL vs search decision | `navigate.go` `DetectNavigation` | `frontend/handlers/navigate.go` (pure, no HTTP) + JS omnibox |
| Search engine list | `SearchEngines` map | `handlers.SearchEngines` + `app.js` `ENGINES` |
| Config / pathutil | both trees | `frontend/config.go`, `frontend/pathutil/` |
| Local state HTML ideas | `web/states_*` | Keep as reference; new local panels live under `frontend/web/` |

---

## New local API (no remote page fetch)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/health` | Liveness; reports proxy retired |
| POST | `/api/notes` | Save highlight JSON |
| GET | `/api/notes?query=` | Search in-memory notes |

Static chrome files are still served from `frontend/web/`.

---

## Navigation today

```
Omnibox → app.js (URL or search) → hostNavigate → TabManager → webview.Navigate
```

No Go handler downloads the page body for display.

Optional: call `handlers.DetectNavigation(input, engine)` from tests or future CLI; the chrome path does not need an HTTP `/api/navigate`.

---

## Plain-text / search results UI (later)

If you want a custom results page again:

1. Let the **native webview** load the search engine URL, **or**
2. Use a **local** state panel that receives structured data from a future backend feature — without Go scraping the engine HTML for the primary content surface.

Do not revive `/api/proxy`.

---

## What you can delete when ready

- Entire `frontend-ui/` directory (after you are satisfied nothing unique remains).
- Any JS that `fetch`es `/api/proxy?url=`.

---

## Next steps

- Step 5: C++23 `backend/` skeleton for durable notes/history/settings.
- Point `NoteStore` at that backend instead of memory.
