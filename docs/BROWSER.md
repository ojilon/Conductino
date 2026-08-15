# In-app browser

## Behaviour

- Omnibox navigates **inside** Conductino (`#browser-frame` iframe), not Edge by default.
- Toolbar on the browser panel:
  - **→ Study** — selection or clipboard → Study left pane
  - **Summarize selection** — same + run summarize
  - **Fetch page text** — Go downloads HTML and strips to text (works when iframe is blank)
  - **System browser** — optional external open

## Why selection is limited

Most research sites (ResearchGate, many publishers) send `X-Frame-Options` / CSP `frame-ancestors`, so the iframe is empty or cross-origin. Browsers **forbid** reading `getSelection()` across origins.

### Practical workflow today

1. Open URL in-app (or system browser if embed blocked).
2. Select text → **Ctrl+C**.
3. **Summarize selection** (reads clipboard) **or** **Fetch page text** for a full extract.

## Future upgrades (not in this slice)

- Native WebView2 child control (true second webview) with script injection for selection — requires Windows native host code beyond plain Wails iframe.
- Optional local proxy / readability pipeline for paywalled HTML (legal/ToS caution).
- Per-tab iframe history (back/forward stacks).
