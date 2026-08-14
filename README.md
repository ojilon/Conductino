# Conductino Study Browser (Desktop)

Local-first research / study browser. PC counterpart of [Conductino-Android](https://github.com/ojilon/Conductino-Android).

### Stack (current)

| Layer | Technology |
|-------|------------|
| App shell | **Wails v2** + Go |
| UI | Vanilla HTML / CSS / JS |
| Native core (optional) | C++23 + CMake (`backend/`) — document extract/cache, no network |
| LLMs | API-agnostic JS adapter (Google AI Studio, OpenRouter, Groq, …) |

### Design rules

- C++ backend does **not** perform network I/O.
- LLM calls run in the webview (`fetch`) with local chunking to control tokens.
- Study workflow: open/paste sources on the left → transfer or summarize into a knowledge document on the right.

### Quick start

```bat
cd frontend
go mod tidy
wails dev
```

See [BUILDING.md](BUILDING.md) for binary builds and optional C++ backend.

### Docs

- [BUILDING.md](BUILDING.md)
- [frontend/README.md](frontend/README.md)
- [docs/ai/ARCHITECTURE.md](docs/ai/ARCHITECTURE.md) — deferred AI features & schemas
- [docs/ai/PROMPTS.md](docs/ai/PROMPTS.md) — summarizer / verifier prompt templates

### Status

| Area | Status |
|------|--------|
| Wails window + chrome UI | Working |
| Study split + resize | Working |
| File open / text extract | Working (PDF notice until lib linked) |
| AI summarize + key management | Working |
| In-app web content (not system browser) | Future |
| Full PDF extract lib | Future |
| Verifier / vision / TTS | Documented in `docs/ai/` |

### License

Add a top-level `LICENSE` when you decide on terms.
