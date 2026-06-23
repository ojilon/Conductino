## Build Instructions

### 1. Prepare C libraries

Download the required libraries into `backend-core/third_party/`:

- SQLite amalgamation: https://sqlite.org/amalgamation.html
- lexbor: https://github.com/lexbor/lexbor (or gumbo-parser)
- ini.h: https://github.com/rxi/ini

See `backend-core/third_party/README.md` for details.

### 2. Build the Zig backend

```bash
cd backend-core
zig build
```

Run it:

```bash
zig build run
# or
./zig-out/bin/research-backend
```

### 3. Build the Go frontend

```bash
cd frontend-ui
go mod tidy
go build -o research-browser.exe
```

Run it:

```bash
./research-browser.exe
```

---