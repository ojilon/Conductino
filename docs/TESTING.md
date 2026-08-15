# Testing guide

Foundational tests exist so you can grow coverage without rewriting structure later.

---

## Layout

```
frontend/
  library_test.go      # library path helpers / pure logic (add more)
  navigate_test.go     # URL normalization helpers if extracted
  ...
backend/               # C++ tests later (Catch2/GTest) — not required for v0.3
```

Wails UI and WebView2 are **hard to unit-test** in pure Go. Prefer:

1. **Pure functions** in Go (paths, chunking ports, HTML strip) — unit tests.  
2. **App methods** that touch disk under `t.TempDir()` — integration-style.  
3. **Manual checklist** for WebView / omnibox / Study (below).

---

## Run tests

```bat
cd frontend
go test ./...
```

Verbose:

```bat
go test ./... -v -count=1
```

---

## What is covered now

| Test | Intent |
|------|--------|
| `TestHtmlToTextSmoke` | `FetchPageText` HTML strip does not panic and returns text |
| `TestLibraryFolderSanitize` | folder rel paths reject `..` when helper exists |

Expand by extracting pure helpers from `library.go` / `fetch.go` (e.g. `SanitizeRelPath`, `htmlToText`) so tests do not need Wails `ctx`.

---

## How to improve (later)

1. **Extract pure logic** from `App` methods into `package` functions → easy `go test`.  
2. **Table-driven tests** for URL detect (`looksLikeUrl` / `normalizeUrl`) — port small helpers to Go or test via thin wrappers.  
3. **Library integration:** create temp dir, set `dataDir`, call `CreateLibraryFolder` / `AddBookmark` / `AppendSummary`, assert files.  
4. **AI provider mocks:** HTTP test server returning fixed Gemini/OpenAI JSON; assert parser.  
5. **C++:** add `backend/tests` when extract is real; run in CI with DLL build.  
6. **CI:** GitHub Action `go test ./...` on push to `main`.  
7. Do **not** start with Selenium for Wails; stabilize dual WebView first, then consider end-to-end.

---

## Manual smoke (before each release)

- [ ] App starts (`wails dev` / built exe)  
- [ ] Settings: save AI key, status line updates  
- [ ] Library: create folder, add bookmark, ensure summary file  
- [ ] Study: paste text, chunk info, summarize (with key)  
- [ ] Omnibox: Enter navigates (note dual WebView status in `DUAL_WEBVIEW.md`)  
- [ ] Open file dialog returns a path  

---

## Naming convention

- Files: `*_test.go` next to code under test  
- Functions: `TestSomething_Situation`  
- Use `t.TempDir()` for any filesystem writes  
