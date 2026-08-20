# Conductino Study Browser (Desktop)

Local-first research / study browser. PC counterpart of [Conductino-Android](https://github.com/ojilon/Conductino-Android).

## Stack

| Layer | Technology |
|-------|------------|
| App shell | **Wails v2** + Go |
| UI | **React 19** + **Vite** + **Tailwind CSS** (`frontend/src/`) |
| Native core (optional) | C++23 + CMake (`backend/`) — document extract/cache, **no network** |
| LLMs | API-agnostic TypeScript adapter (Google AI Studio, OpenRouter, Groq, …) |

## Design rules

- C++ / Go backend does **not** perform network I/O for page loads.
- Remote pages load **inside the app** (iframe content surface under chrome).
- LLM calls run in the webview (`fetch`) with local chunking to control tokens.
- Study workflow: open/paste sources → transfer or summarize into a knowledge document.

## Quick start

```bat
cd frontend
npm install
go mod tidy
wails dev
```

See [BUILDING.md](BUILDING.md) for binary builds and optional C++ backend.

## Project map

| Path | Role |
|------|------|
| [`frontend/`](frontend/) | Wails app: Go bindings + React UI |
| [`backend/`](backend/) | Optional C++ core (extract, storage stubs) |
| [`docs/`](docs/README.md) | Architecture, guides, concepts |
| [`tools/`](tools/) | Dev helpers |

Documentation index: **[docs/README.md](docs/README.md)**

Key concepts:

- [Text extraction & structure](docs/concepts/TEXT_EXTRACTION.md) — plain text, AST-style trees, PDF/DOCX path
- [AI architecture](docs/ai/ARCHITECTURE.md) — chunking, providers, knowledge blocks
- [GUI / chrome](docs/architecture/GUI.md)

## Status

| Area | Status |
|------|--------|
| Wails + React chrome (tabs, toolbar, right sidebar) | Working |
| In-app URL / search (iframe surface) | Working (sites may block framing) |
| Study split + AI summarize | Working |
| File open / text extract | Working (PDF notice until lib linked) |
| Full PDF / AST pipeline | Documented; partial |
| Verifier / vision / TTS | Documented in `docs/ai/` |

## License

Add a top-level `LICENSE` when you decide on terms.
