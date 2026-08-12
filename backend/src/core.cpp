#include "conductino/core.h"

#include <filesystem>
#include <mutex>
#include <string>

namespace {

std::mutex g_mu;
bool g_ready = false;
std::filesystem::path g_data_dir;

} // namespace

// Internal accessors used by other translation units.
namespace conductino::detail {

std::filesystem::path data_dir() {
    std::lock_guard lock(g_mu);
    return g_data_dir;
}

bool ready() {
    std::lock_guard lock(g_mu);
    return g_ready;
}

} // namespace conductino::detail

extern "C" {

CONDUCTINO_API const char* conductino_version(void) {
    return "0.1.0-restructure";
}

CONDUCTINO_API int conductino_init(const char* data_dir) {
    if (data_dir == nullptr || data_dir[0] == '\0') {
        return 1;
    }
    std::lock_guard lock(g_mu);
    try {
        g_data_dir = std::filesystem::path(data_dir);
        std::filesystem::create_directories(g_data_dir);
        g_ready = true;
        return 0;
    } catch (...) {
        g_ready = false;
        return 2;
    }
}

CONDUCTINO_API void conductino_shutdown(void) {
    std::lock_guard lock(g_mu);
    g_ready = false;
    g_data_dir.clear();
}

CONDUCTINO_API void conductino_free(void* p) {
    // Buffers from notes/settings use new[] / malloc-compatible pattern.
    ::operator delete(p);
}

} // extern "C"
