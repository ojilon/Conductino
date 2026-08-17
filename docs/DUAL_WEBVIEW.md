# Dual WebView2 — failure analysis (Wails + Go)

This document explains **why** the chrome + content dual WebView2 layout is not working in Conductino today. It is written for someone who will fix it later. It does **not** describe alternate browsing UX.

**Note:**
**The dual webview problem has been fixed**
**The remaining problem is viewing the search results, summary of the issue is below:**
Omnibox Enter calls `App.Navigate` → `ContentNavigate` → second controller `Navigate(url)`. When #2 never paints, the user sees a **blank grey workspace** even though the omnibox URL updated.

---

## 1. What “dual WebView2” means here

Target layout:

```
┌──────────────────────────────────────────-----------┐
│  Tabs + omnibox + tool buttons                      │  ← WebView2 #1 (Wails host)
│  (HTML/CSS/JS chrome, never leaves app)             │
├─────────────────────────────────────────-----------─┤
│                                                     │
│  Page content (DuckDuckGo, website eg ResearchGate) │  ← WebView2 #2 (child HWND)
│                                                     │
└─────────────────────────────────────────-----------─┘
```

- **WebView2 #1**: owned by Wails. Serves `frontend/` assets, Go bindings (`window.go.main.App.*`), Study, Library, Settings.
- **WebView2 #2**: a second controller on a **child HWND** under the same top-level window, reserved for `http(s)` navigation so chrome stays visible.

Omnibox Enter calls `App.Navigate` → `ContentNavigate` → second controller `Navigate(url)`. When #2 never paints, the user sees a **blank grey workspace** even though the omnibox URL updated.

---

## 2. How Wails uses WebView2

Wails v2 on Windows embeds **one** WebView2 for the whole client area of the app window:

1. Creates a Win32 top-level window.
2. Creates a `CoreWebView2Environment` (user data folder under the app profile).
3. Creates a `CoreWebView2Controller` bound to that window’s client area.
4. Navigates that controller to the asset server (`http://wails.localhost` or the dev server).
5. Injects JS bridges so `window.go` can call Go methods on the UI / dispatcher thread.

Important properties of this design:

| Property | Implication for a second WebView |
|----------|----------------------------------|
| Single controller fills the **entire** client rect | A child HWND must sit **on top** of part of that surface and take input |
| Environment + user-data folder are already active | A second environment must use a **different** user-data path or compatible options |
| Controller APIs must run on the **window UI thread** | Creating #2 from a random Go goroutine or some binding paths fails |
| `go-webview2` `Chromium.Embed` **pumps messages** until `inited` | Calling `Embed` inside nested `SendMessage` / WndProc can re-enter the pump and hit invalid COM state |

There is **no** public Wails API to “split” the host controller into chrome-only bounds. Dual layout must be done with a **second** controller on a **separate** HWND.

---

## 3. What we implemented

Files:

- `frontend/content_webview_windows.go` — child host, subclass, UI-thread messages, `edge.Chromium`
- `frontend/content_webview_stub.go` — non-Windows stubs
- `frontend/navigate.go` — `useDualContent = true` → `ContentNavigate` first

Sequence attempted:

1. Find top-level HWND (title contains `"Conductino"`).
2. Subclass parent `WndProc` (`SetWindowLongPtr` / `GWLP_WNDPROC`).
3. On UI thread (`SendMessage` custom `WM_APP+*`):
   - `CreateWindowEx(WS_CHILD, …)` under parent
   - Size host below chrome (~82px)
   - `edge.NewChromium()` with DataPath `%LocalAppData%\Conductino\content-webview2`
   - `Chromium.Embed(hostHWND)`
4. Later: `Navigate`, `Resize`, `SetVisible` via the same UI-thread messages.

---

## 4. Observed failures (evidence)

### 4.1 `CoInitialize has not been called`

```
[WebView2] Environment creation failed … CoInitialize has not been called.
… Chromium.Embed ← ContentBrowser.initChromium ← ContentEnsure ← binding dispatcher
```

**Cause:** First attempts called `Embed` from the Wails **binding** path without COM initialized on that thread.

**Mitigation tried:** `CoInitializeEx(COINIT_APARTMENTTHREADED)` (and MTA fallback) before `Embed`.

**Result:** Environment creation then succeeded; problem moved downstream.

### 4.2 Environment OK, controller fails — `0x8007139F`

```
[WebView2] Environment created successfully
[WebView2 Error] error creating controller with 8007139f:
  The group or resource is not in the correct state to perform the requested operation.
```

HRESULT `0x8007139F` = `ERROR_INVALID_STATE`. In WebView2 this typically means controller creation was refused for the given HWND / environment / thread combination.

Known related causes in WebView2 community reports:

- Controller created off the thread that owns the parent HWND message loop
- Incompatible **environment options** vs an existing environment for a shared user-data folder
- HWND not ready (zero size, destroyed, wrong parent chain)
- Re-entrancy: message pump inside `Embed` while already inside a nested window procedure

### 4.3 Later runs: environment OK, no controller error, **blank surface**

```
[content] DataPath=…\content-webview2 host=… popup=false
[WebView2] Environment created successfully
```

Chrome (tabs/omnibox) visible; content area solid grey; page never appears.

Possible interpretations:

- Controller never finished (`inited` never set) while `ready` was set optimistically in our code
- Controller exists but bounds/visibility leave an empty surface under the Wails webview z-order
- Child HWND is covered by the host WebView2 (host paints full client area on top of siblings depending on z-order / composition)

### 4.4 Omnibox symptom

Typing a query and pressing Enter updates the tab title / omnibox string (JS chrome state) but **no page** appears because `ContentNavigate` targets WebView #2, which is not producing visible content.

---

## 5. Root causes (ordered by likelihood)

### A. Threading and message-pump re-entrancy (high)

`github.com/wailsapp/go-webview2` `Chromium.Embed`:

1. Starts async environment creation.
2. Runs a **local `GetMessage` / `DispatchMessage` loop** until `atomic inited != 0`.

If `Embed` is invoked from:

- A binding dispatcher callback, or
- A subclassed `WndProc` entered via `SendMessage`,

then that nested pump interacts with Wails’ own pump and COM callbacks. That matches both `CoInitialize` issues and `0x8007139F` during `CreateCoreWebView2Controller`.

**What a correct fix must guarantee:**

- Create the **child HWND** on the same thread that owns the parent window (UI thread).
- Create the **controller** on that UI thread.
- Do **not** nest a second unbounded message loop inside Wails’ critical sections if it races controller completion.
- Prefer posting work to the UI thread and completing via a callback/`inited` wait **outside** `SendMessage` when the caller is already the UI thread (detect with `GetWindowThreadProcessId` vs `GetCurrentThreadId`).

### B. Z-order / composition with Wails’ full-client controller (high)

Wails’ controller is sized to the **full client area**. A `WS_CHILD` sibling created later may:

- Sit under the WebView2 composition surface (never visible), or
- Receive no hit-testing while the host webview still thinks it owns the pixels.

Edge/Chrome do **not** put “HTML chrome WebView2” and “content WebView2” as two siblings under one Wails-managed HWND this way. They use a **native** chrome frame and one (or more) content views whose bounds the shell controls explicitly.

**What a correct fix must guarantee:**

- Either shrink the **host** WebView2 bounds to the chrome strip only (requires access to Wails’ controller `PutBounds` — not exposed today), **or**
- Host content WebView2 in a way that is guaranteed above the host surface (dedicated child with correct parent, z-order, and possibly `WS_CLIPSIBLINGS` + explicit `SetWindowPos` HWND insertion after host webview HWND), **or**
- Use a separate owned window whose position tracks the content rect (still dual WebView2, different parenting).

### C. Environment / user-data isolation (medium)

Second environment uses `%LocalAppData%\Conductino\content-webview2`. That is correct for isolation. If options ever differ while sharing a folder, WebView2 returns invalid state. Keep **one folder per controller role** and identical options for that folder.

### D. Optimistic `ready` flag (medium)

Our code sometimes set `ready = true` after `Embed` returned without verifying `controller != nil` / navigation capability. `Embed` can return after environment start while controller creation still fails asynchronously. A fix must only mark ready after controller completion (watch `inited` or a completion callback without assuming sleep is enough).

### E. HWND discovery (low–medium)

`findAppHWND("Conductino")` picks a visible top-level window whose title contains the substring. Wrong window (devtools, splash, secondary) would yield invalid parenting. Prefer storing HWND from Wails startup if a public or unsafe path exists, or match class name + PID more strictly.

---

## 6. Why Edge/Chrome “just work” but this does not

| Browser shell | Conductino (current) |
|---------------|----------------------|
| Native C++/WinUI chrome, not a full-window WebView2 for UI | Entire UI is WebView2 #1 (Wails) |
| Content view bounds set by the shell on the UI thread as first-class controls | Second controller bolted on after the fact onto Wails’ HWND tree |
| One product owns environment lifecycle end-to-end | Two stacks (Wails + hand-rolled `go-webview2`) share one process |
| No nested `Embed` message loop inside another framework’s dispatcher | `Embed` message loop + Wails dispatcher + subclass `SendMessage` |

Dual WebView2 **is** valid Win32 (multiple controllers, multiple HWNDs). The failure is integration with **Wails’ ownership of the top-level window and its single full-client controller**, not “WebView2 cannot do two views.”

---

## 7. Checklist for a real fix (dual only)

Use this when revisiting the feature:

1. **Obtain parent HWND** reliably at `OnStartup` / first paint (log class, title, TID).
2. **Create child host HWND** only on UI thread; non-zero size; `WS_CHILD | WS_CLIPSIBLINGS | WS_CLIPCHILDREN`.
3. **Z-order:** place content host above the Wails webview child; verify with Spy++ / WinSpy which HWND paints the grey region.
4. **Environment:** unique user-data path; log success of environment **and** controller completion separately.
5. **Ready gate:** block `Navigate` until controller completion; surface errors to logs and optional UI status line.
6. **No nested Embed pump inside SendMessage** if that path still shows `0x8007139F`; switch to PostMessage + completion channel from a non-blocking UI handler, or call create only when already on UI thread without `SendMessage` self-send.
7. **Resize:** on parent `WM_SIZE`, set content host rect to `client height - chromeHeight`; call controller `Resize` / `PutBounds`.
8. **Visibility:** hide content host when Study/Library panels need the full workspace; show on omnibox navigate.
9. **Optional hard path:** if Wails exposes (or you fork) host controller `PutBounds`, shrink WebView #1 to chrome height so #2 is the only surface in the content band — closest to a real browser shell.

---

## 8. Code map

| Symbol / file | Role |
|---------------|------|
| `useDualContent` in `navigate.go` | Feature flag for ContentNavigate-first |
| `ContentEnsure` / `ContentNavigate` | Public App bindings |
| `ensureContentBrowser` | HWND find, subclass, create |
| `uiCreateContent` | UI-thread child + Embed |
| `parentSubclassProc` | Handles `WM_APP` create/nav/resize |
| `contentDataPath` | Isolated user-data folder |

---

## 9. Log lines to capture on the next debugging session

```
[content] subclassed parent HWND=...
[content] UI-thread Embed host=...
[WebView2] Environment created successfully
[WebView2 Error] error creating controller with ...   # if still present
[content] WebView2 ready on UI thread
[content] Navigate https://...
```

Also note: whether Spy++ shows the content host HWND with non-zero size and visible style after Enter in the omnibox.

---

*Document version: 2026-08-15. Dual content remains the intended architecture; this file records why it is not production-ready yet.*
