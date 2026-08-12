# Persistent chrome shell (critical UX fix)

## What you saw in the trial

1. Chrome loads correctly (tabs, omnibox, welcome).
2. You search `leaf` → log shows `hostNavigate → https://duckduckgo.com/?q=leaf`.
3. DuckDuckGo fills the **entire** window.
4. Tab strip / URL bar / sidebar are gone — not a tab-model bug; the chrome document was replaced.
5. New tabs still exist in Go (`tab new → 2`, `3`) but nothing can paint them until chrome is loaded again.

## Root cause

`webview_go` gives **one** native surface.  
`hostNavigate` → `webview.Navigate(url)` is correct for **content** (native network, Cloudflare, real browser identity) but wrong if that same surface is also the **chrome** HTML.

Android does not have this problem: XML layout keeps the toolbar; `WebView` is only the content region.

## Target architecture (same idea as Android)

```
┌── Window (OS) ────────────────────────────────────────┐
│  CHROME SURFACE (always visible)                              │
│  tabs · back/fwd/reload · omnibox · sidebar toggle            │
│  served from frontend/web  OR drawn by a UI toolkit           │
├───────────────────────────────────────────────────────────┤
│  CONTENT SURFACE (swappable)                                  │
│  · remote URL → native webview Navigate (never iframe)        │
│  · local states (welcome, settings) → local HTML templates    │
└───────────────────────────────────────────────────────────┘
```

Rules unchanged:

- Remote pages load **natively** on the content surface (no Go proxy, no iframe).
- Local templates only for app states (welcome, settings, stubs).
- Backend C++ still has **no** browsing network.

## Implementation options

| Option | Pros | Cons |
|--------|------|------|
| **A. Dual WebView2** (UI env + content env, like Microsoft WebView2Browser) | Chrome HTML stays; content is native | Windows-first; more code |
| **B. Wails v3 WebviewPanel** (when stable) | Chrome + panel API | Depends on Wails v3 panel maturity |
| **C. Native toolbar + one content webview** | Closest to Android XML | Rebuild chrome outside HTML |

**Chosen direction for Conductino desktop:** **A**, with package `frontend/shell` abstracting surfaces so we can swap backends.

## Work sequence (small steps)

1. **This commit** — document + `shell` interfaces/stubs; stop treating single Navigate as final UX.
2. **Next** — content surface controller: create/navigate/bounds for content-only webview (Windows WebView2 path first).
3. **Then** — chrome surface stays on local URL forever; omnibox only drives **content** Navigate.
4. **Then** — tab switch = content Navigate to that tab’s URL; chrome never unloads.
5. **Parallel** — cgo wire to `conductino_core` for notes/settings.

## Interim mitigation (until dual surface ships)

- `hostShowChrome` returns to the shell (already bound).
- Prefer not expanding features that assume chrome stays mounted after remote Navigate.

## Related

- `docs/GUI.md` — chrome widgets
- `docs/BRIDGE.md` — binds (will target content surface)
- `frontend/shell/` — code for this plan
