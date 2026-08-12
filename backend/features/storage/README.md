# features/storage

**Purpose:** Durable database and file layout under `data_dir`.

**Current:** Plain file helpers in `src/storage.cpp`.

**Next:**

1. Vendor SQLite amalgamation under `backend/third_party/sqlite/`.
2. Schema for notes, history, bookmarks.
3. Keep network out of this module.

**Recommended libs:** SQLite amalgamation only (no server).
