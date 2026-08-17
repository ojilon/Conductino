# Dual WebView2 (Windows)

## Architecture

```
┌─────────────────────────────────────┐
│ Tabs + omnibox + tools (Wails WV2)  │  chrome — never navigates away
├─────────────────────────────────────┤
│ Content WebView2 (child HWND)       │  real sites
└─────────────────────────────────────┘
```

Content WebView2 is created on the **parent window UI thread**:

1. Subclass Wails HWND (`SetWindowLongPtr` → custom `WndProc`)
2. `SendMessage(WM_APP+CREATE)` → handler runs on UI thread
3. `CreateWindowEx(WS_CHILD)` + `Chromium.Embed(host)` on that thread
4. Navigate / resize / visibility via further UI-thread messages

Fallback: if dual create fails, `Navigate` uses full-window load + floating bar.(not desirable)

## Logs to expect

```
[content] subclassed parent HWND=...
[content] UI-thread Embed host=...
[content] WebView2 ready on UI thread
[content] Navigate https://...
```
