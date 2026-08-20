# backend/ — Conductino native core (C++23)

Primary native core for the **desktop** app. **No network I/O** lives here.

```
backend/
  CMakeLists.txt
  cmake/CompilerFlags.cmake   # C++23 + strict warnings
  include/conductino/
    core.h                    # public C ABI (Go cgo)
    core.hpp                  # thin C++ wrappers
  src/
    core.cpp                  # init / shutdown / version
    storage.cpp               # file helpers under data_dir
    notes.cpp                 # notes.jsonl stub
    settings.cpp              # settings.kv stub
    demo_main.cpp             # optional demo binary
  features/                   # domain stubs + READMEs
  third_party/                # external libs (not committed)
```

## Build

```bash
cd backend
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build

# demo
./build/conductino_backend_demo   # or build/Release/... on multi-config
```

Requirements: CMake ≥ 3.25, a C++23 compiler (GCC 13+, Clang 17+, MSVC recent).

## Rules

| Do | Don't |
|----|--------|
| Storage, notes, settings, document parsing | HTTP clients, sockets for page load |
| Stable C ABI in `core.h` | Expose unstable C++ types to Go |
| Feature folder + README before large code | Dump everything in one .cpp |

Page loading is owned by **webview_go** in `frontend/`. This library only persists and processes data **after** the user has content.

