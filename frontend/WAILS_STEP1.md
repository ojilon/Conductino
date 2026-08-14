# Wails rewrite — Step 1

**Status:** Scaffold only. Proves a real OS window + JS↔Go binding.

## Why

The previous `webview_go` dual-surface layout broke chrome vs content (results area / sidebar overlap). Wails owns one native webview properly.

## Run (Windows)

```bat
cd frontend
go mod tidy
go run .
```

Or with Wails CLI (optional):

```bat
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
wails dev
```

You should see a dark window titled **Conductino Study Browser** with a **Ping Go** button. Click it — status should show a greeting from `App.Greet`.

## Layout

```
frontend/
  main.go              # wails.Run + embed
  app.go               # bound methods (Greet, AppInfo)
  wails.json
  go.mod
  frontend/            # asset root (Wails AssetServer)
    index.html
    js/boot.js
```

Old dual-webview code (`bridge/`, `shell/`, `web/`) is still in the tree until Step 5 cleanup. Do not use it for the new path.

## Next

**Step 2** — Port chrome (tabs, toolbar, sidebar, welcome/settings/study panels) into `frontend/frontend/`.

Tell me when Step 1 window works (or paste errors) and we continue.
