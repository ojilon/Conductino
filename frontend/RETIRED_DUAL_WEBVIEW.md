# Retired: dual-webview frontend

These packages were part of the pre-Wails shell and caused chrome/content overlap bugs:

- `frontend/bridge/` (except patterns reimplemented under Wails `app.go` / `native_*.go`)
- `frontend/shell/`
- `frontend/web/` (old asset root)
- `frontend/handlers/`, `frontend/digest_request/`, etc. tied to the old HTTP+webview host

**Active app entry:**

```
frontend/main.go
frontend/app.go
frontend/wails.json
frontend/frontend/     ← HTML/CSS/JS assets
```

Safe to delete the retired trees in a later cleanup PR once you no longer need them for reference. They are not compiled by the Wails `main` package.
