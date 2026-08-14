# Conductino AI Tooling Architecture

**Purpose:** Local-first academic research / knowledge-compression layer.
External LLMs only receive text (and alt-text descriptions). All parsing, chunking, file I/O, UI, citations, and verification orchestration stay inside the app.

This document is the long-term blueprint. Code that is deferred is marked **[FUTURE]** so you can implement it later without guessing the intended design.

---

## 1. Core principles (locked)

| Principle | Rule |
|-----------|------|
| API agnostic | One adapter interface. Switch provider by changing `endpoint`, `apiKey`, `model`. Application logic never branches on vendor. |
| No network in C++ | `backend/` never opens sockets. LLM calls live in the webview (JS `fetch`) or thin Go helpers. |
| Local heavy lifting | PDF/text extract, chunking, cache, citation list, DOM/markdown insertion, TTS, highlights → local. |
| Token efficiency | Chunk on device. Prefer short system prompts + structured JSON outputs. |
| Observable steps | Streaming responses + visible intermediate blocks in the right pane. |
| Auditable blocks | Every transferred/summarised block carries source id, page range, model, timestamp. |

---

## 2. Module map

```
frontend/ai/
  provider.js          # Unified LLM adapter (working)
  chunker.js           # Local semantic / size chunking (working)
  output_parser.js     # Citations, placeholders, DOM/md insert (working)
  orchestrator.js      # [FUTURE] multi-step workflows + failover
  workspace.js         # [FUTURE] split-view state controller

backend/features/document/
  (C++ extract + cache ABI)  # text working path; PDF lib later

backend/features/ai/         # optional Go/C++ helpers for queue / progress
```

---

## 3. Data schemas (JSON)

### 3.1 Chunk
```json
{
  "id": "chunk-00042",
  "sourceId": "doc-abc123",
  "index": 42,
  "text": "...",
  "approxPage": 7,
  "headingPath": ["Methods", "Sample preparation"],
  "tokenEstimate": 480,
  "charStart": 12040,
  "charEnd": 13410
}
```

### 3.2 Document metadata
```json
{
  "id": "doc-abc123",
  "title": "Cell wall structure in ...",
  "sourceType": "import" | "external" | "url",
  "pathOrUrl": "C:/Users/.../paper.pdf" | "imports/2026-08-14_paper.pdf",
  "importedAt": "2026-08-14T18:00:00Z",
  "pageCount": 18,
  "textHash": "sha256:...",
  "cachePath": "cache/docs/abc123.txt"
}
```

### 3.3 Image / figure placeholder
```json
{
  "id": "img-009",
  "sourceId": "doc-abc123",
  "approxPage": 4,
  "altText": "Electron micrograph showing ...",
  "placeholder": "[[IMG:img-009]]",
  "originalSrc": "cache/figs/abc123_p4_f2.png"
}
```

### 3.4 Summary / knowledge block (right pane)
```json
{
  "id": "block-778",
  "type": "summary" | "exact" | "note" | "outline",
  "text": "...",
  "sourceId": "doc-abc123",
  "chunkIds": ["chunk-00042", "chunk-00043"],
  "pageRange": [6, 8],
  "model": "gemini-2.0-flash",
  "provider": "google-ai-studio",
  "createdAt": "2026-08-14T18:12:00Z",
  "citations": [
    { "label": "[1]", "sourceId": "doc-abc123", "pages": "7-8" }
  ],
  "verified": false,
  "flags": []
}
```

### 3.5 Citation list (appended to knowledge doc)
```json
{
  "entries": [
    {
      "id": "cite-1",
      "label": "[1]",
      "title": "...",
      "sourceId": "doc-abc123",
      "pathOrUrl": "...",
      "pages": "7-8"
    }
  ]
}
```

---

## 4. Provider adapter contract

```js
// provider.js exports
createProvider(config) → {
  id, name,
  complete({ system, user, stream, signal }) → Promise<string> | AsyncIterable,
  isAvailable() → boolean
}
```

Config shape:
```json
{
  "id": "google",
  "name": "Google AI Studio",
  "endpoint": "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent",
  "apiKey": "...",
  "model": "gemini-2.0-flash",
  "headers": {},
  "maxTokens": 2048
}
```

Failover / teaming **[FUTURE orchestrator]**: try primary → on 429 / 5xx switch to next configured provider. Route specialised tasks (verify vs summarise vs alt-text) by capability tags in config.

---

## 5. Document processing pipeline

1. User picks file via native dialog (Go bind).
2. Choice: **Import** (copy into `data_dir/imports/`) or **Link** (store absolute path only).
3. C++ `conductino_document_extract(path)` → plain text + optional page map (cache under `data_dir/cache/docs/<hash>.txt`).
4. JS chunker turns text into Chunk[].
5. Payload assembly builds prompt windows that stay under free-tier limits.
6. Provider streams summary → output_parser inserts blocks + citations into right pane / markdown file.

**PDF libraries [FUTURE]**: CMake option that downloads a lightweight extractor (e.g. pdfium / poppler / mupdf) into `backend/third_party` once; keep repo light. Until then text/md and cached extracts work fully.

---

## 6. Vision / alt-text strategy

1. Dedicated (cheap) vision call or local heuristic produces detailed technical alt-text.
2. Alt-text is injected into the main summariser prompt as `[[IMG:id]] ... description ...`.
3. Summariser never “sees” pixels; it references the description.
4. UI later stitches the real image back at the placeholder location.

Full loop is **[FUTURE]**; schema and prompt template already defined in `PROMPTS.md`.

---

## 7. Multi-step verification

Pass 1 — Summarise chunks.  
Pass 2 — Evaluator prompt receives original chunk(s) + summary; returns JSON flags (`hallucination`, `missing_context`, `contradiction`).  
User can accept / edit / re-run.

Orchestrator + evaluator wiring is **[FUTURE]**; prompt lives in `PROMPTS.md`.

---

## 8. Citation & editing engine

`output_parser.js` already:
- Detects `[n]` / `(Author, year)` style markers and `[[IMG:...]]` placeholders.
- Builds a citation list object.
- Appends a References section to the target markdown / DOM.

Further academic styles (APA, Vancouver, BibTeX export) = **[FUTURE]**.

---

## 9. Split-view knowledge workspace

Left: source list + active document viewer (highlightable).  
Right: growing knowledge document (editable blocks with provenance).

Basic pane structure and data model land in the UI pass. Full multi-doc queue, progressive auto-summary, and TTS-sync are **[FUTURE]** improvements documented here so you can add them without redesign.

---

## 10. File open / import model

- Go bind `hostOpenFileDialog(filters)` → path string or empty.
- Follow-up choice UI: Import | Link external.
- Import copies file into `data_dir/imports/<timestamp>_<name>` and records the new relative path.
- Link stores the absolute path; app must tolerate missing files gracefully.
- Progress / sync markers for phone-PC later use the same document metadata store.

---

## 11. Improvement backlog (implement when needed)

- Full C++ PDF extract via downloaded third-party lib + page images.
- Orchestrator with capability-based routing and automatic provider failover.
- Verification agent loop with UI flags.
- Vision alt-text pass + image re-stitch.
- TTS read-along with sentence highlight + parallel summary stream.
- Batch multi-document comparative outline workflows driven by `.md` scripts.
- Persistent session state + phone/PC progress sync.
- SQLite-backed document / chunk / block index.
- Auth / sign-in notes (already flagged separately).

Each item above has enough schema and prompt detail in this folder to be added by an experienced beginner without architectural guesswork.
