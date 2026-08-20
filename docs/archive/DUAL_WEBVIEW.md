# Dual WebView2 — historical failure analysis

> **Archive.** Active navigation uses an in-app iframe under React chrome. See [architecture/BROWSER.md](../architecture/BROWSER.md).

This document explains why a chrome + content dual WebView2 layout failed in earlier experiments. Kept for maintainers who may revisit a native content surface.

---

## Intent

Two WebView2 controllers: #1 Wails chrome, #2 child HWND for http(s). Omnibox → second controller Navigate. When #2 never painted, users saw a blank grey workspace.

## Failure modes observed

1. **COM / threading** — `CoInitialize` missing; then `0x8007139F` (invalid state) creating a second controller from binding / nested message-pump paths.
2. **Z-order** — Wails controller fills the full client area; sibling child HWND often invisible under composition.
3. **Optimistic ready** — `Embed` returned before controller completion.

## Why Edge/Chrome differ

Native chrome shell + first-class content bounds on the UI thread, one product owning environment lifecycle — not two stacks (Wails + hand-rolled go-webview2) on one HWND tree.

## If revisiting dual native content

1. Reliable parent HWND at startup.
2. Child host only on UI thread; non-zero size; clip siblings.
3. Z-order above host webview or shrink host PutBounds to chrome height (needs Wails access).
4. Unique user-data path per controller; ready only after controller completion.
5. No nested Embed message pump inside SendMessage.

Full narrative of the original debugging session lived in the pre-archive `docs/DUAL_WEBVIEW.md` commit history.
