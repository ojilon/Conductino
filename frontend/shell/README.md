# frontend/shell

Persistent application frame: **chrome surface** + **content surface**.

| File | Role |
|------|------|
| `shell.go` | Interfaces (`Host`, `ChromeSurface`, `ContentSurface`) |
| `single.go` | Interim one-webview adapter (current behavior) |
| `dual_windows.go` | *Planned* — dual WebView2 on Windows |

See **docs/SHELL.md** for why this exists and the rollout steps.

Do not put remote page loads back into an iframe inside the chrome HTML.
