# Library / data directories

App data (relative to process CWD in dev) under `conductino-data/`:

| Path | Use |
|------|-----|
| `imports/` | Copied documents |
| `cache/docs/` | Extract cache (native) |

Library helpers live in Go module `conductino-backend/library`. UI may grow a picker later (`PickLibraryFolder` was sketched in older bindings; not required for current App surface).
