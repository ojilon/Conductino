# Text extraction & document structure

How Conductino turns files (and later pages) into text the Study workspace and AI chunker can use.

## Goals

1. **Local-first** — extract on device; no upload of documents to a server by default.
2. **Stable plain text** for LLMs — normalized newlines, no NULs, enough structure for chunking.
3. **Optional richer structure** — headings, pages, code ASTs — without blocking the simple path.

## Pipeline (current)

```
User picks file (Wails OpenFile dialog)
        │
        ▼
App.ExtractDocument(path)          ← Go binding
        │
        ├─ native C++ conductino_document_extract  (if DLL present)
        │         optional cache under data_dir/cache/docs/
        │
        └─ extractTextGo(path)     fallback
                  .txt/.md/… → read bytes, strip \0
                  .pdf/.docx → notice string until library linked
        │
        ▼
Study pane (React) + chunkText() in src/lib/ai/chunker.ts
```

Import copies the file into `conductino-data/imports/` via `ImportDocument`.

## Formats

| Format | Today | Direction |
|--------|--------|-----------|
| Plain / Markdown / code / JSON / YAML | Full text via Go | Keep |
| PDF | Notice + paste workaround | C++ or dedicated lib → page-aware text |
| DOCX | Notice | ZIP+XML or lib → paragraphs |
| HTML (saved file) | Raw read | Optional DOM → main content |
| Live page in iframe | Not extracted yet | Optional: selection transfer / readability later |

## Layers of structure

Think of three levels. Each builds on the previous; none requires the next.

### 1. Linear text

What the UI and LLM see first: a single string. Chunker splits on paragraphs / size (`chunkText`).

Good enough for notes, articles, and most summaries.

### 2. Page / block map

Metadata alongside text:

```json
{
  "pages": [
    { "index": 1, "charStart": 0, "charEnd": 3200 },
    { "index": 2, "charStart": 3200, "charEnd": 6400 }
  ]
}
```

Chunk objects already carry `approxPage` (heuristic from char offset). A real PDF extractor should fill exact ranges so citations can say “pp. 12–13”.

### 3. Tree / AST-style representation

For **structured** sources, prefer a tree over a flat string:

| Source | Natural tree |
|--------|----------------|
| Markdown / HTML | Heading hierarchy, lists, code fences |
| Source code | Language AST (functions, classes) |
| DOCX | Paragraphs + styles + tables |
| PDF | Pages → blocks → lines (layout-aware) |

Example conceptual node:

```ts
type DocNode = {
  kind: 'document' | 'section' | 'paragraph' | 'code' | 'table' | 'list_item'
  title?: string
  text?: string           // leaf text
  children?: DocNode[]
  span?: { start: number; end: number }  // offsets into linearized text
  meta?: Record<string, unknown>         // language, page, heading level
}
```

**Why trees help**

- Chunk by *section* instead of arbitrary character windows.
- Preserve heading path in prompts (`headingPath` on chunks is reserved for this).
- For code: summarize a function body without dragging the whole file.
- For study: jump from a knowledge citation back to a section node.

**Linearization:** always be able to flatten the tree to plain text for models that only accept strings. Keep `span` offsets so UI selection and citations stay aligned.

## Techniques (toolkit)

| Technique | Use when |
|-----------|----------|
| Raw decode + charset normalize | `.txt`, logs |
| Markdown / CommonMark parse | Structure without HTML noise |
| HTML DOM + readability / main-content heuristic | Saved pages, later live DOM |
| PDF text layer (library) | Born-digital PDFs |
| OCR | Scanned PDFs / images (heavy; optional) |
| DOCX XML (word/document.xml) | Office text + basic styles |
| Tree-sitter / language parsers | Code-aware study |
| AST from compiler front-ends | Deep code intelligence (later) |

Conductino should **not** invent a new PDF format parser if a maintained library fits behind the existing C ABI (`conductino_document_extract`).

## Interaction with Study UI

1. **Open file** → extract → left pane plain text (tree can power outline later).
2. **User selection** → exact transfer or summarize selection only.
3. **Chunk info** → token estimates from linear text today; later from tree sections.
4. **Knowledge blocks** store `sourceId`, `chunkIds`, optional `pageRange` for citations.

## Backend boundary

- Extraction and caching: **Go / C++ only** (offline).
- No network in `backend/` for fetching documents.
- Frontend never needs the binary format if extract returns UTF-8 text (+ optional JSON sidecar for the tree).

## Implementation checklist (future)

- [ ] PDF library behind same C ABI; page map JSON sidecar
- [ ] DOCX paragraph extract
- [ ] Optional `DocNode` JSON export from extract
- [ ] Chunker: prefer section boundaries when tree present
- [ ] Study outline view driven by heading nodes
- [ ] Code file: tree-sitter grammar pack (optional feature flag)

See also: [ai/ARCHITECTURE.md](../ai/ARCHITECTURE.md), [backend/features/document/README.md](../../backend/features/document/README.md).
