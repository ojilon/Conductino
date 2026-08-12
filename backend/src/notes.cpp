/**
 * Notes persistence stub — append JSON lines to data_dir/notes.jsonl
 * Replace with SQLite FTS when features/storage is implemented.
 */

#include "conductino/core.h"

#include <filesystem>
#include <sstream>
#include <string>
#include <vector>

namespace conductino::detail {
std::filesystem::path path_under(const std::string& relative);
bool write_text_file(const std::filesystem::path& path, const std::string& body);
bool read_text_file(const std::filesystem::path& path, std::string& out);
bool ready();
} // namespace conductino::detail

namespace {

std::filesystem::path notes_path() {
    return conductino::detail::path_under("notes.jsonl");
}

} // namespace

extern "C" {

CONDUCTINO_API int conductino_notes_save_json(const char* json, size_t len) {
    if (!conductino::detail::ready() || json == nullptr) {
        return 1;
    }
    std::string line(json, len);
    // strip newlines so one record = one line
    for (char& c : line) {
        if (c == '\n' || c == '\r') {
            c = ' ';
        }
    }
    line.push_back('\n');

    std::string existing;
    (void)conductino::detail::read_text_file(notes_path(), existing);
    existing += line;
    return conductino::detail::write_text_file(notes_path(), existing) ? 0 : 2;
}

CONDUCTINO_API int conductino_notes_search(const char* query, char** out_json, size_t* out_len) {
    if (out_json == nullptr || out_len == nullptr) {
        return 1;
    }
    *out_json = nullptr;
    *out_len = 0;

    std::string file;
    if (!conductino::detail::read_text_file(notes_path(), file)) {
        // empty result
        const char empty[] = "[]";
        char* buf = static_cast<char*>(::operator new(sizeof(empty)));
        for (size_t i = 0; i < sizeof(empty); ++i) {
            buf[i] = empty[i];
        }
        *out_json = buf;
        *out_len = sizeof(empty) - 1;
        return 0;
    }

    // Very small filter: include lines that contain query (or all if query empty).
    // Not a JSON parser — good enough for the skeleton.
    std::string q = query ? query : "";
    std::ostringstream arr;
    arr << '[';
    bool first = true;
    std::istringstream iss(file);
    std::string line;
    while (std::getline(iss, line)) {
        if (line.empty()) {
            continue;
        }
        if (!q.empty() && line.find(q) == std::string::npos) {
            continue;
        }
        if (!first) {
            arr << ',';
        }
        first = false;
        arr << line;
    }
    arr << ']';
    std::string body = arr.str();
    char* buf = static_cast<char*>(::operator new(body.size() + 1));
    for (size_t i = 0; i < body.size(); ++i) {
        buf[i] = body[i];
    }
    buf[body.size()] = '\0';
    *out_json = buf;
    *out_len = body.size();
    return 0;
}

} // extern "C"
