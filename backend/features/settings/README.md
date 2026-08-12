# features/settings

**Purpose:** Theme, default search engine, and other preferences.

**Current:** `settings.kv` key=value file via `conductino_settings_*`.

**Next:**

1. Align keys with chrome (`theme`, `search_engine`).
2. Load at Go startup and push into the webview.
3. Replace kv file with SQLite settings table when storage lands.
