# Conductino documentation map

Start here. Docs are grouped by audience and lifetime.

## Guides

| Doc | Purpose |
|-----|---------|
| [../BUILDING.md](../BUILDING.md) | Install, `wails dev`, binary, C++ DLL |
| [../frontend/README.md](../frontend/README.md) | Frontend layout (React + Wails) |
| [../backend/README.md](../backend/README.md) | C++ core overview |

## Architecture

| Doc | Purpose |
|-----|---------|
| [architecture/FOUNDATION.md](architecture/FOUNDATION.md) | Rules, layout, non-goals |
| [architecture/GUI.md](architecture/GUI.md) | Chrome, panels, in-app content surface |
| [architecture/BRIDGE.md](architecture/BRIDGE.md) | Go ↔ JS (Wails bindings + runtime) |
| [architecture/BACKEND.md](architecture/BACKEND.md) | C ABI, feature modules |
| [architecture/LIBRARY.md](architecture/LIBRARY.md) | Local library / data dirs |
| [architecture/BROWSER.md](architecture/BROWSER.md) | Navigation model (in-app vs external) |

## Concepts

| Doc | Purpose |
|-----|---------|
| [concepts/TEXT_EXTRACTION.md](concepts/TEXT_EXTRACTION.md) | Document → text, AST-style trees, chunking inputs |
| [concepts/SECURITY.md](concepts/SECURITY.md) | Trust boundaries, keys, no backend network |

## AI

| Doc | Purpose |
|-----|---------|
| [ai/ARCHITECTURE.md](ai/ARCHITECTURE.md) | Chunk / provider / knowledge-block schemas |
| [ai/PROMPTS.md](ai/PROMPTS.md) | Summarizer / verifier templates |

## Product

| Doc | Purpose |
|-----|---------|
| [product/FEATURES_ROADMAP.md](product/FEATURES_ROADMAP.md) | Near-term and later features |
| [product/TESTING.md](product/TESTING.md) | How to smoke-test |
| [product/RELEASE.md](product/RELEASE.md) | Release checklist |

## Archive (historical)

Kept for context; **not** the active design.

| Doc | Notes |
|-----|--------|
| [archive/DUAL_WEBVIEW.md](archive/DUAL_WEBVIEW.md) | Retired dual-surface approach |
| [archive/MIGRATION.md](archive/MIGRATION.md) | frontend-ui → current shell |
| [archive/SHELL.md](archive/SHELL.md) | Old shell notes |

**Convention:** every non-trivial directory should have a short README describing purpose and how to extend it.
