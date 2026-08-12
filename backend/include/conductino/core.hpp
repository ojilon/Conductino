#pragma once

/**
 * C++23-facing helpers (optional). Hosts that prefer C++ can include this.
 * The stable ABI for Go remains core.h.
 */

#include "conductino/core.h"

#include <optional>
#include <string>
#include <string_view>

namespace conductino {

inline std::string_view version() {
    return conductino_version();
}

inline bool init(std::string_view data_dir) {
    return conductino_init(std::string(data_dir).c_str()) == 0;
}

inline void shutdown() {
    conductino_shutdown();
}

} // namespace conductino
