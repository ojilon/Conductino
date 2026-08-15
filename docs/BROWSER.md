# Browser model (current)

## Working path (default)

- Omnibox navigates the **main WebView2** (same engine as Edge).
- Real sites work: DuckDuckGo, ResearchGate, Cloudflare, etc.
- A **floating tool bar** is injected on web pages:
  - Conductino Home
  - Copy selection
  - Selection → Study
  - Summarize selection
  - Hide
- Tools run **only when you click them**.

## Dual chrome + content WebView2 (deferred)

Code lives in `content_webview_windows.go` but is **disabled** (`useDualContent = false` in `navigate.go`).

Why deferred:
- Second WebView2 controller often fails or shows a blank pane when created from Wails binding threads (`0x8007139F` / empty surface).
- Needs create-on-UI-thread via `PostMessage` subclassing of the parent HWND — non-trivial native work.

### Future enable steps (for you or a later session)

1. Subclass parent HWND; handle `WM_APP+N` to create child host + `Embed` on the UI thread.
2. Unique DataPath (already set under `%LocalAppData%\Conductino\content-webview2`).
3. Set `useDualContent = true` in `navigate.go` when controller creation is reliable.
4. Keep chrome height sync via `ContentSetChromeHeight`.

Until then, full-window navigate + floating bar is the production path for study use.
