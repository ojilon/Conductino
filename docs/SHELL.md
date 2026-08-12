# Persistent chrome shell

## Problem (trial)

One `webview.Navigate` replaced the entire chrome document — tabs/omnibox vanished while Go still tracked tabs.

## Target

```
CHROME SURFACE  — always local shell (tabs, omnibox, sidebar)
CONTENT SURFACE — remote native WebView2  OR  local state panels
```

No iframe for remote pages. Backend still has no browsing network.

## Step 1 (done on Windows)

| Piece | Detail |
|-------|--------|
| `shell.DualHost` | Chrome via webview_go; content via separate controller |
| `shell.ContentWebView` | `wailsapp/go-webview2/pkg/edge` Chromium, top padding 128px |
| Separate user-data dir | `conductino-data/content-webview` |
| Fallback | Non-Windows / embed failure → single-surface + warning |

After pull:

```bash
cd frontend
go get github.com/wailsapp/go-webview2@latest
go mod tidy
go run .
```

Look for `[shell] dual surface active`. Search `leaf` — chrome band should remain.

## Later steps

2. Dynamic chrome height + WM_SIZE resize  
3. Tab switch only moves content surface  
4. cgo to `conductino_core`  

## Related

- `docs/GUI.md` · `docs/BRIDGE.md` · `frontend/shell/README.md`
