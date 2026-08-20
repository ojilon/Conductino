cmake_minimum_required(VERSION 3.25)
project(conductino_backend LANGUAGES CXX)

list(APPEND CMAKE_MODULE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/cmake")
include(CompilerFlags)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

option(CONDUCTINO_BUILD_DEMO "Build small demo executable" ON)
option(CONDUCTINO_BUILD_SHARED "Build shared library (for Go cgo)" ON)

# ---------------------------------------------------------------------------
# Core library — no network I/O by design (see docs/FOUNDATION.md)
# ---------------------------------------------------------------------------
set(CONDUCTINO_CORE_SOURCES
    src/core.cpp
    src/storage.cpp
    src/notes.cpp
    src/settings.cpp
    src/document.cpp
)

if(CONDUCTINO_BUILD_SHARED)
    add_library(conductino_core SHARED ${CONDUCTINO_CORE_SOURCES})
else()
    add_library(conductino_core STATIC ${CONDUCTINO_CORE_SOURCES})
endif()

target_include_directories(conductino_core
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(conductino_core PUBLIC project_warnings)

set_target_properties(conductino_core PROPERTIES
    OUTPUT_NAME "conductino_core"
    CXX_VISIBILITY_PRESET hidden
    VISIBILITY_INLINES_HIDDEN ON
)

target_compile_definitions(conductino_core PRIVATE CONDUCTINO_BUILDING=1)

if(CONDUCTINO_BUILD_DEMO)
    add_executable(conductino_backend_demo src/demo_main.cpp)
    target_link_libraries(conductino_backend_demo PRIVATE conductino_core project_warnings)
endif()

install(TARGETS conductino_core
    ARCHIVE DESTINATION lib
    LIBRARY DESTINATION lib
    RUNTIME DESTINATION bin
)
install(DIRECTORY include/ DESTINATION include)
