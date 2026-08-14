# Conductino Library — study folders, bookmarks, summaries

This is the **second-last phase** model before merge readiness.

## Mental model (your workflow)

```
plantphysiology/                 ← course unit (you create)
  growth/                        ← topic
    bookmarks                    ← ResearchGate / paper links
    downloads/                   ← optional local PDF copies
    summaries/
      growth-summary.md          ← one living summary for all papers on this topic
  hormonalsecretion/
    ...
```

- You **create the folder tree** (app helpers + native dialogs).
- Bookmarks and downloads **belong to a folder**.
- Summaries for papers in that topic **append into the same summary file**.
- You **review AI edits** (diff / accept) — active participation.
- Later: merge two topic summaries into a **new** document.

## On disk

```
conductino-data/
  library/
    index.json              # bookmarks + summary registry
    plantphysiology/
      growth/
        downloads/
        summaries/
          growth-summary.md
```

`index.json` shape (simplified):

```json
{
  "version": 1,
  "bookmarks": [
    {
      "id": "bm-…",
      "folder": "plantphysiology/growth",
      "url": "https://…",
      "title": "Paper title",
      "localPath": "",
      "createdAt": "…"
    }
  ],
  "summaries": [
    {
      "id": "sum-…",
      "folder": "plantphysiology/growth",
      "relPath": "plantphysiology/growth/summaries/growth-summary.md",
      "title": "growth-summary",
      "updatedAt": "…"
    }
  ]
}
```

## Implemented now (Go + UI)

| Capability | Status |
|------------|--------|
| Create / list folders under library root | ✅ |
| Bookmark URL → choose folder | ✅ |
| List bookmarks in folder | ✅ |
| Default summary file per folder + append | ✅ |
| Read / write summary text | ✅ |
| Merge two summary files → new file | ✅ basic |
| Right-hand sidebar | ✅ |
| Library panel in UI | ✅ |
| PDF render / thumbnails / in-app web highlight | Documented below |
| AI edit review (diff accept/reject) | Documented below |

## Deferred (working notes for you)

### PDF / DOCX extract + render

- **Extract:** link a real PDF library in C++ (`backend/features/document`) or pure Go (`pdfcpu` / `ledongthuc/pdf`) — prefer one path and cache text under `cache/docs/`.
- **Render left pane:** WebView can show PDF via browser engine for *local* `file://` or blob URLs; or embed pdf.js in the Wails assets.
- **Thumbnails:** render first page with pdf.js or C++ raster; store PNG under `library/.../previews/`.
- **Right pane as PDF/DOCX:** keep **Markdown as canonical** for edit + AI; export PDF/DOCX on demand (Go libraries later).

### In-page highlight → summarize

Needs either:
1. In-app browser surface (second webview / Browser component), or
2. Browser extension / bookmarklet that posts selection to Conductino.

Until then: **Paste** or **Open file** in Study remains the path.

### AI edit review

1. AI proposes new section or rewrite.
2. UI shows **side-by-side or unified diff** (old vs proposed).
3. Buttons: **Accept / Reject / Edit then accept**.
4. Only accepted text is written to the summary file.

Implement when summarize-to-folder is stable.

### Downloads

`DownloadToFolder(url, folder)` — Go HTTP fetch into `library/<folder>/downloads/` and register `localPath` on the bookmark. Not in this slice.

## API surface (`App` methods)

See `frontend/library.go`:

- `LibraryRoot`, `ListLibraryFolders`, `CreateLibraryFolder`
- `ListBookmarks`, `AddBookmark`, `RemoveBookmark`
- `EnsureSummary`, `AppendSummary`, `ReadSummaryFile`, `WriteSummaryFile`
- `MergeSummaryFiles`
