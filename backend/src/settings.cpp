/**
 * Key/value settings as one JSON-ish file: data_dir/settings.kv
 * Lines: key=value (value may contain '='). Skeleton only.
 */

#include "conductino/core.h"

#include <filesystem>
#include <map>
#include <sstream>
#include <string>

namespace conductino::detail {
std::filesystem::path path_under(const std::string& relative);
bool write_text_file(const std::filesystem::path& path, const std::string& body);
bool read_text_file(const std::filesystem::path& path, std::string& out);
bool ready();
} // namespace conductino::detail

namespace {

std::filesystem::path settings_path() {
    return conductino::detail::path_under("settings.kv");
}

std::map<std::string, std::string> load_map() {
    std::map<std::string, std::string> m;
    std::string raw;
    if (!conductino::detail::read_text_file(settings_path(), raw)) {
        return m;
    }
    std::istringstream iss(raw);
    std::string line;
    while (std::getline(iss, line)) {
        auto pos = line.find('=');
        if (pos == std::string::npos) {
            continue;
        }
        m[line.substr(0, pos)] = line.substr(pos + 1);
    }
    return m;
}

bool save_map(const std::map<std::string, std::string>& m) {
    std::ostringstream oss;
    for (const auto& [k, v] : m) {
        oss << k << '=' << v << '\n';
    }
    return conductino::detail::write_text_file(settings_path(), oss.str());
}

} // namespace

extern "C" {

CONDUCTINO_API int conductino_settings_get(const char* key, char** out_value, size_t* out_len) {
    if (key == nullptr || out_value == nullptr || out_len == nullptr) {
        return -1;
    }
    *out_value = nullptr;
    *out_len = 0;
    if (!conductino::detail::ready()) {
        return -2;
    }
    auto m = load_map();
    auto it = m.find(key);
    if (it == m.end()) {
        return 1;
    }
    const std::string& v = it->second;
    char* buf = static_cast<char*>(::operator new(v.size() + 1));
    for (size_t i = 0; i < v.size(); ++i) {
        buf[i] = v[i];
    }
    buf[v.size()] = '\0';
    *out_value = buf;
    *out_len = v.size();
    return 0;
}

CONDUCTINO_API int conductino_settings_set(const char* key, const char* value) {
    if (key == nullptr || value == nullptr) {
        return 1;
    }
    if (!conductino::detail::ready()) {
        return 2;
    }
    auto m = load_map();
    m[key] = value;
    return save_map(m) ? 0 : 3;
}

} // extern "C"
