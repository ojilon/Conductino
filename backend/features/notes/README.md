# features/notes

**Purpose:** Persist highlights and study notes created after a page is loaded in the native webview.

**Current:** Append-only `data_dir/notes.jsonl`; search is substring over lines.

**Next:**

1. Validate JSON against the same schema as `frontend/handlers/notes.go`.
2. Move to SQLite + FTS5 via `features/storage`.
3. Wire `POST /api/notes` in Go to `conductino_notes_save_json` via cgo.

**Recommended libs:** SQLite amalgamation (already used on Android path).
