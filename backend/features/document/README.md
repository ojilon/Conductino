# features/document

**Purpose:** Offline document processing (PDF text extract, HTML structure for study tools) **after** content is available locally or from user export — not for live navigation.

**Current:** Empty stub.

**Next:**

1. Define C ABI for `extract_text(path) -> string`.
2. Optional: lexbor for local HTML files only.
3. Never use this module to fetch URLs.

**Recommended libs:** existing PDF extract ideas from old Go code; lexbor if needed for local HTML.
