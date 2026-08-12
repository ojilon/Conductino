# Conductino — Go ↔ JS Bridge & Tab Model

**Step 3 of the restructure.**

## Ownership

| Concern | Owner |
|---------|--------|
| Open tabs, active tab, per-tab history | Go `bridge.TabManager` |
| Native page load | Go `webview.Navigate` via `Host` |
| Chrome rendering (tab strip, omnibox, buttons) | JS `app.js` |
| User gestures | JS → `ConductinoBridge` → Go binds |

JS holds only a **snapshot** for painting. Mutations go through Go.

---

## Bindings (`webview.Bind`)

| Name | Direction | Purpose |
|------|-----------|---------|
| `hostNavigate(url)` | JS → Go | Record history on active tab + native Navigate |
| `hostGoBack` / `hostGoForward` | JS → Go | Tab history then native Navigate |
| `hostReload` | JS → Go | Reload current tab URL (or chrome if empty) |
| `hostTabNew` | JS → Go | Create + activate empty tab, show chrome |
| `hostTabClose(id)` | JS → Go | Close tab; navigate to next active URL or chrome |
| `hostTabActivate(id)` | JS → Go | Switch tab; native Navigate to that tab’s URL |
| `hostTabList` | JS → Go | JSON snapshot of all tabs |
| `hostShowChrome` | JS → Go | Navigate back to local chrome shell |
| `hostMinimize` / `hostMaximize` / `hostClose` | JS → Go | Window controls |
| `hostPing` | JS → Go | Health check |

Go → JS updates use `Eval` to call:

```js
window.ConductinoChrome.applyTabSnapshot([...])
```

---

## Files

```
frontend/bridge/
  tabs.go    # Tab + TabManager
  host.go    # Host, binds, PushTabsToChrome
frontend/web/js/
  bridge.js  # ConductinoBridge helpers
  app.js     # chrome UI + snapshot render
```

---

## Navigation flow

1. User submits omnibox → `app.js` normalizes URL or builds search URL.
2. `ConductinoBridge.navigate(url)` → `hostNavigate`.
3. `TabManager.Navigate` appends history, updates title/flags.
4. `Host.navigateNative` → `webview.Navigate(url)` (**native**, no iframe).
5. Snapshot pushed to chrome when still on chrome document; after remote load, chrome is replaced until `hostShowChrome`.

Tab switch / back / forward follow the same pattern: update model, then native Navigate.

---

## Extending

- **New bind:** add in `bridge/host.go` `Bind`, document here, optional wrapper in `bridge.js`.
- **Richer history:** still inside `TabManager` until multi-webview exists.
- **Persist tabs:** serialize `TabManager.Snapshot()` to the C++ backend later.

---

## Known limitation

Single webview_go surface: remote Navigate replaces chrome HTML. See `docs/GUI.md` § Content surface for the multi-webview path.
