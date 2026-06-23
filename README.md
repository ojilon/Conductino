# Conductino Study Browser

### NOTE: PROJECT STILL AT INITITIAL STAGES, NOT YET, ALOT OF THINGS SEEN HERE NOT YET ACHIEVED, THIS IS AS A GUIDE TO MYSELF TOO

A modular research browser built with **Go** (frontend shell + IPC router) and
**Zig/C** (high-performance document & storage backend). Also **plain HTML/CSS/JS** for the UI skin. The two runtimes communicate over a local REST API.

```
Basic project structure
```
```
Conductino-study-browser/
│
├── backend-core/          # Zig / C territory
│   ├── src/
│   │   ├── main.zig       # HTTP server + entry point
│   │   ├── document.zig   # Text / HTML / PDF processing
│   │   └── storage.zig    # SQLite (C amalgamation) wrapper
│   ├── build.zig          # Zig build configuration
│   └── third_party/       # C libraries like: sqlite3.c / sqlite3.h / lexbor / ini.h
│
├── frontend-ui/           # Go territory
│   ├── main.go            # Initializes WebView2 window + IPC server
│   ├── go.mod
│   ├── config.yaml
│   ├── handlers/
│   │   └── api.go         # Forwards JSON packets to the Zig backend
│   └── web/               # The "skin" of the browser (loaded by WebView2)
│       ├── index.html     # Control bar · WebView pane · Study sidebar
│       ├── style.css
│       └── app.js         # Captures text selections, calls /api/save_note
│
└── README.md

```

---

### Architecture
---

## Data Pipeline

```
[Raw Input]
    ↓
[Zig Memory Buffer Engine]
    ↓
[SQLite Engine]
```

```
┌──────────────────────────────────────────────────────────────┐
│  Frontend: Go + WebView2 (github.com/webview/webview_go)     │
│  - main.go      : native window + IPC server                 │
│  - handlers/api.go : JSON bridge to Zig backend              │
│  - web/         : HTML/CSS/JS browser skin                   │
└───────────────────────┬──────────────────────────────────────┘
                        │ REST API / JSON
                        │ eg using 127.0.0.1:8080  <->  127.0.0.1:8081
                        ▼
┌──────────────────────────────────────────────────────────────┐
│  Backend: Zig + C libraries                                  │
│  - main.zig      : HTTP server, routes requests              │
│  - document.zig  : MemoryBufferEngine + HTML parser (lexbor) │
│  - storage.zig   : SQLite C bindings + FTS5                  │
└───────────────────────┬──────────────────────────────────────┘
                        │ C FFI
                        ▼
┌──────────────────────────────────────────────────────────────┐
│  Embedded C libraries                                        │
│  - sqlite3       : notes/highlights database                 │
│  - lexbor        : HTML5 parsing & structural indexing       │
│  - ini.h         : backend configuration                     │
└──────────────────────────────────────────────────────────────┘
```

```


## Architectural rules

| Rule                       | Implementation                                              |
| -------------------------- | ----------------------------------------------------------- |
| Local Context Isolation    | Web content ↔ Backend logic communicate **only** over REST  |
| Data pipeline              | `[Raw Input] → [Zig Memory Buffer] → [SQLite Engine]`       |
| No external pkg managers   | Zig native build system, Go modules only                    |
| Manual memory in Zig       | like `std.mem.Allocator` usage                              |

## Library usage map

### Go (frontend-ui)
| Library                            | Purpose                                            |
| ---------------------------------- | -------------------------------------------------- |
| `github.com/webview/webview_go`    | Wraps Microsoft WebView2 (native window on Win11)  |
| `gopkg.in/yaml.v3`                 | Parses `config.yaml` at startup                    |
| `golang.org/x/net/html`            | Tokenizes data streams when archiving pages       |
| `github.com/ledongthuc/pdf`        | Extracts text maps from local PDF files            |

### Zig / C (backend-core)
| Library                            | Purpose                                            |
| ---------------------------------- | -------------------------------------------------- |
| `sqlite3` (C amalgamation)         | Embedded note/metadata storage + FTS5 search       |
| `lexbor` (pure C)                  | Ultra-fast HTML5 parsing & structural indexing     |
| `ini.h`                            | Tiny INI parser for backend config                 |

## Build & run

```bash
# 1. Build the Zig backend
cd backend-core
zig build -Doptimize=ReleaseFast

# 2. Run the Go frontend (it will spawn the Zig backend as a child process)
cd ../frontend-ui
go run main.go
```

## API endpoints (Zig backend)

| Method | Route             | Body / Query             | Description                       |
| ------ | ----------------- | ------------------------ | --------------------------------- |
| POST   | `/api/save_note`  | JSON `NoteHighlightEvent`| Persists a highlight into SQLite  |
| GET    | `/api/search`     | `?query=memory`          | FTS5 search over highlights       |


## Some short notes
- **Why manual memory in Zig?** It gives deterministic allocation and deallocation. There is no garbage collector that could pause the UI thread while parsing large files.
- **How do JSON packets bridge Go and Zig?** Both languages can serialize and deserialize JSON. The Go frontend builds a struct, marshals it to bytes, sends it over HTTP, and the Zig backend parses it back into a struct.
- **Why embed SQLite as C code?** SQLite is battle-tested, file-based, and requires no external database server. Zig imports `sqlite3.h` directly and links the amalgamation into one binary.