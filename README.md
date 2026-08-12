# Conductino Study Browser (Desktop)

**Branch: `restructure`** — foundational rewrite in progress.

A modular research / study browser for the desktop.  
It is the PC counterpart of [Conductino-Android](https://github.com/ojilon/Conductino-Android).

### Stack

| Layer | Technology |
|-------|------------|
| Shell & chrome | Go + [webview_go](https://github.com/webview/webview_go) + vanilla HTML/CSS/JS |
| Content surface | **Native webview** (loads remote pages itself) |
| Native core | C++23 + CMake (`backend/`) |
| Experiments | Zig (`backend-core/`, optional) |

### Critical design rule

> The native webview is the only component that talks to the network for page loads.  
> Go does **not** fetch pages and inject them into an iframe.  
> The C++ backend does **not** perform network I/O.

This keeps the browser looking like a normal human-driven browser (Cloudflare, anti-bot systems, cookies, redirects, etc. work as expected) and keeps the architecture clean.

After a page is natively loaded and visible, the app may still inject scripts, extract text, or offer study tools (highlights, notes, reader view, …).

---

## Quick links

- Foundation & roadmap → [docs/FOUNDATION.md](docs/FOUNDATION.md)
- GUI skeleton & extension guide → [docs/GUI.md](docs/GUI.md)
- Docs index → [docs/README.md](docs/README.md)

---

## Project layout (target)

```
Conductino/
├── frontend/          # Go + webview_go + HTML/CSS/JS chrome
├── backend/           # C++23 + CMake (storage, document, … — no network)
├── backend-core/      # Zig experiments (kept for now)
├── config/            # search engines, settings schema, …
├── docs/              # foundation, GUI, themes, …
├── frontend-ui/       # OLD working tree (to be removed after migration)
└── …
```

See [docs/FOUNDATION.md](docs/FOUNDATION.md) for the full map and rules.

---

## Status (restructure)

| Area | Status |
|------|--------|
| Architecture rules & docs | Done |
| Chrome-like GUI skeleton | Next |
| Native webview as sole content surface | Next |
| Themes (dark/light) + multi search engines | Planned |
| C++23 backend skeleton | Planned |
| Migration of useful non-network logic from `frontend-ui` | Planned |

---

## Building (will be updated as the skeleton lands)

See [BUILDING.md](BUILDING.md). The old `frontend-ui` instructions still apply for the previous working tree; the new `frontend/` path will be documented as soon as the skeleton runs.

---

## Contributing / extending

- Prefer small, focused changes.
- Add a short README in any new feature directory.
- Follow the rules in `docs/FOUNDATION.md` (especially the native-webview network rule).
- GUI extension points are described in `docs/GUI.md`.

---

## License

Add a top-level `LICENSE` when you decide on terms.
