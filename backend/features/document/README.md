# features/document

**Purpose:** Offline document processing after content is available locally.

## Working now

| Capability | Status |
|------------|--------|
| `conductino_document_extract(path)` | Text-like files (`.txt`, `.md`, `.json`, source code, …) |
| Cache under `data_dir/cache/docs/` | Keyed by filename + size + mtime |
| `conductino_document_import(src)` | Copies into `data_dir/imports/` |
| PDF / DOCX | Returns a clear notice + caches it; full extract is optional |

## API (C ABI in `include/conductino/core.h`)

```c
int conductino_document_extract(const char* path, char** out_text, size_t* out_len);
int conductino_document_import(const char* src_path, char** out_rel, size_t* out_len);
```

Caller frees buffers with `conductino_free`.

## Adding a PDF library later

1. Prefer CMake `FetchContent` or a one-shot download into `backend/third_party/` (keep git light).
2. Implement real extract behind the same ABI — JS/Go do not change.
3. Optionally emit page-map JSON alongside text for better `approxPage` metadata.

See `docs/ai/ARCHITECTURE.md` for chunk / image / citation schemas.
