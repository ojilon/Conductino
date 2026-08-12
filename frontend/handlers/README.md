# frontend/handlers

Local HTTP helpers for the chrome process.

**Allowed:** notes API, health, pure navigation classification, future settings JSON.  
**Forbidden:** fetching remote pages for display (see `docs/MIGRATION.md`).

| File | Role |
|------|------|
| `api.go` | Route registration; 410 on retired proxy paths |
| `notes.go` | In-memory note store + `/api/notes` |
| `navigate.go` | `DetectNavigation` (no network) |
