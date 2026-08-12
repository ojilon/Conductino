# frontend/shell

Persistent frame: **chrome surface** + **content surface**.

| File | Role |
|------|------|
| `shell.go` | Interfaces |
| `single.go` | One-webview interim |
| `dual.go` | DualHost — chrome permanent, content separate |
| `content_windows.go` | WebView2 content controller (Windows) |
| `content_stub.go` | Non-Windows stub |

## Step 1 status

On **Windows**, after chrome loads, a second WebView2 is embedded on the same HWND with **top padding** (~128px) for the tab/toolbar band. Remote `hostNavigate` only drives that content controller — chrome HTML is not replaced.

```bash
cd frontend
go get github.com/wailsapp/go-webview2@latest
go mod tidy
go run .
```

Expect log: `dual surface active` then search should leave tabs/omnibox visible.

If attach fails, behaviour falls back to single-surface (chrome wiped on navigate).

## Next

- Measure chrome height from JS instead of fixed 128px
- Resize content bounds on WM_SIZE
- Hide content when showing settings/welcome only
