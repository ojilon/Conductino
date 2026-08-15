# Release process (Conductino)

First production-minded release can ship **without** a working dual content WebView2. Document known limits in the release notes.

---

## Version numbers (manual)

We use **manual** semver. No auto-bump script required.

### Where to change the version

| Location | Field |
|----------|--------|
| `frontend/wails.json` | `info.productVersion` |
| `frontend/app.go` | `AppInfo()` → `"version"` |
| Git tag | `vMAJOR.MINOR.PATCH` |

**Current first release target: `0.3.0`**

Rules of thumb:

- **PATCH** (0.3.1): fixes only  
- **MINOR** (0.4.0): features (library, study, extract improvements)  
- **MAJOR** (1.0.0): dual WebView stable + you are willing to support it  

---

## Pre-release checklist

1. Update versions in `wails.json` and `app.go`.  
2. Run `go test ./...` from `frontend/`.  
3. Manual smoke from `docs/TESTING.md`.  
4. Write release notes (known issue: dual content WebView2 — see `docs/DUAL_WEBVIEW.md`).  
5. Commit on the branch you will merge / tag from.

---

## Build the Windows executable

From `frontend/` (WebView2 runtime must be installed on the machine):

```bat
cd frontend
go mod tidy
wails build
```

Output (typical):

```text
frontend\build\bin\Conductino.exe
```

Optional production flags:

```bat
wails build -clean -upx
```

(`-upx` only if UPX is installed; skip if unsure.)

Copy next to the exe if you ship the native DLL:

- `backend\build\libconductino_core.dll` (if you rely on native extract)

Ship a folder, e.g.:

```text
Conductino-0.3.0-windows-amd64/
  Conductino.exe
  libconductino_core.dll   (optional)
  README.txt               (short: needs WebView2 runtime)
```

Zip that folder for GitHub Release assets.

---

## Git tag and GitHub Release

```bat
git checkout main
git pull
REM after merge of foundation branch:
git tag -a v0.3.0 -m "Conductino 0.3.0 — first packaged release"
git push origin v0.3.0
```

GitHub UI: **Releases → Draft a new release → choose tag v0.3.0**

- Title: `Conductino 0.3.0`  
- Attach the zip  
- Body template:

```markdown
## Conductino 0.3.0

First packaged study-browser build (Wails + Go).

### Included
- Study workspace (paste/open text, summarize via configured AI provider)
- Library folders, bookmarks, summary append/merge (disk under conductino-data)
- Settings for AI providers (Google / OpenRouter / Groq / custom)

### Known limitations
- Dual content WebView2 (chrome + page pane) is **not** production-ready; see docs/DUAL_WEBVIEW.md
- Omnibox navigation to a second content surface may show a blank pane until that is fixed
- PDF/DOCX full extract/render still limited (see docs/FEATURES_ROADMAP.md)

### Requirements
- Windows 10/11
- Microsoft Edge WebView2 Runtime
```

---

## Next versions

After dual WebView works, bump to `0.4.0` and remove the known-limitation bullet. Keep `docs/RELEASE.md` updated when you add signing (Authenticode) or an installer (Inno Setup / WiX).
