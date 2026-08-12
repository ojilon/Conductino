## Build Instructions

> **Note (restructure branch):** The primary development path is moving to `frontend/` + `backend/`.  
> The instructions below still cover the older `frontend-ui` + `backend-core` tree.  
> They will be rewritten once the new skeleton is runnable.

### Current (legacy) path

#### 1. Prepare C libraries (Zig backend)

Download required libraries into `backend-core/third_party/` if you still use it:

- SQLite amalgamation: https://sqlite.org/amalgamation.html
- lexbor (or equivalent HTML parser)
- ini.h (optional)

#### 2. Build the Zig backend (optional / experimental)

```bash
cd backend-core
zig build
```

#### 3. Run the older Go frontend

```bash
cd frontend-ui
go mod tidy
go run .
# or
go build -o conductino
./conductino
```

### Upcoming (restructure)

```bash
# Frontend (Go + webview_go)
cd frontend
go mod tidy
go run .

# Backend (C++23) — once CMakeLists is in place
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

Exact flags, dependency list, and packaging steps will be filled in when the corresponding skeleton lands.  
See `docs/FOUNDATION.md` for the architectural target.
