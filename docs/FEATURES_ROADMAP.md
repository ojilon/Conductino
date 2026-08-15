# Feature implementation guides

Practical notes for adding study features after the dual WebView2 content pane is reliable. Written for an experienced beginner.

---

## 1. Real PDF / DOCX extract + render + thumbnails

### Goal

- Extract text (and optionally images) from local PDF/DOCX for Study / summarize.
- Render pages in-app (left knowledge source).
- Show folder thumbnails (small page-1 previews).

### Suggested layout

```
backend/features/document/     # C++ extract (existing skeleton)
frontend/                      # Go bindings already call NativeDocumentExtract
frontend/frontend/js/pdf_view.js   # viewer UI (new)
```

### Extract (text)

1. **C++ (preferred for production)**  
   - PDF: PDFium or MuPDF (text layer + page count).  
   - DOCX: unzip `word/document.xml` + strip tags, or a small library (e.g. libdocx).  
   - Expose `conductino_extract(path) -> { text, pages[], meta }` via existing `libconductino_core.dll`.

2. **Go fallback**  
   - PDF: `github.com/ledongthuc/pdf` or call a CLI only if you accept the dependency.  
   - DOCX: `archive/zip` + XML parse of `word/document.xml`.

3. Wire through `App.ExtractDocument` (already exists) to return real text instead of the binary placeholder.

### Render

- **PDF.js** in the chrome WebView (left pane): load file as blob URL from Go (`ReadFile` binding) or `file://` if allowed.  
- Or PDFium bitmap per page from C++ → PNG bytes → `<img>` / canvas.

### Thumbnails

- On import/download: render page 1 at low DPI → PNG under `conductino-data/library/<folder>/thumbs/`.  
- Library UI: `<img src="...">` grid when a folder is selected.

### Acceptance

- Open a multi-page PDF → text appears in Study source.  
- Thumbnail appears in library folder view.  
- Large files: chunk extract; do not load entire binary into the LLM.

---

## 2. In-page highlight on websites

### Goal

User selects text on a live page → Summarize / Transfer without leaving the app.

### Depends on

Working **content** WebView2 (dual pane). Cross-origin selection is readable **inside** that page’s JS world, not from the chrome WebView.

### Implementation outline

1. On content WebView `NavigationCompleted`, inject script (WebView2 `AddScriptToExecuteOnDocumentCreated` or `ExecuteScript`):
   - `mouseup` / `selectionchange` → optional floating mini-toolbar, **or** only react to chrome buttons.
2. Chrome buttons **✨ / →📚**:
   - `ContentEval` → `getSelection().toString()` then `chrome.webview.postMessage` **or** write clipboard and read from chrome (clipboard is a workable bridge).
3. Prefer **WebMessage** from content → Go → EventsEmit to chrome JS for cleaner flow once dual WebView is stable.

### Right-click menu (later)

- WebView2 `ContextMenuRequested` / custom menu: Summarize selection, Bookmark, Download target if `a[href]` ends with `.pdf`.

### Acceptance

- Select paragraph on Wikipedia in content pane → Summarize → text in Study with citation metadata.

---

## 3. AI edit accept / reject diff UI

### Goal

When AI appends or rewrites a summary, user reviews before it becomes permanent.

### Model

```text
PendingEdit {
  id, folder, summaryRelPath,
  beforeText, afterText,  // or unified diff hunks
  source: "summarize" | "merge" | "rewrite",
  createdAt
}
```

### UI

- Right Study pane: show pending block with **Accept** / **Reject** / **Edit manually**.  
- Store pending in memory + optional `conductino-data/library/.../pending/*.json`.  
- Accept → `WriteSummaryFile` / `AppendSummary`. Reject → discard.

### LLM side

- Ask model for structured JSON: `{ "summary": "...", "claims": [] }` via existing provider + output_parser.  
- Never auto-write library files without Accept (configurable later).

### Acceptance

- Summarize → folder creates pending unit; disk unchanged until Accept.

---

## 4. Download PDFs into `folder/downloads/`

### Goal

Save a PDF (or any file) into the library tree, e.g. `plantphysiology/growth/downloads/`.

### Flow

1. User on a page or with a PDF URL.  
2. Action: **Download to library folder** (prompt folder or use selected library folder).  
3. Go: `http.Get` / WebView download API → write under `LibraryRoot()/rel/downloads/filename.pdf`.  
4. Register bookmark or `index.json` entry with `type: "file"`, path relative to library.

### Implementation tips

- Prefer content WebView `DownloadStarting` event when dual WebView works (intercept and redirect path).  
- Until then: explicit URL + Go download binding `DownloadToLibrary(url, folderRel)`.  
- Sanitize filenames; refuse path traversal (`..`).

### Acceptance

- File exists on disk under the chosen folder; Library refresh lists it; Open file uses extract pipeline.

---

## Dependency order

1. Dual content WebView2 (see `docs/DUAL_WEBVIEW.md`)  
2. In-page highlight + download intercept  
3. PDF render/thumbnails (can start in parallel with extract C++)  
4. Accept/reject diff (works on Study pane even before dual WebView)
