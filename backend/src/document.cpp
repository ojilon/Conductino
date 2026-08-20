/**
 * Local document extract + import + simple text cache.
 * No network. PDF/DOCX full extract deferred to optional third-party libs.
 */

#include "conductino/core.h"

#include <algorithm>
#include <cctype>
#include <chrono>
#include <cstring>
#include <filesystem>
#include <fstream>
#include <sstream>
#include <string>
#include <vector>

namespace fs = std::filesystem;

namespace conductino::detail {
fs::path data_dir();
bool ready();
bool write_text_file(const fs::path& path, const std::string& body);
bool read_text_file(const fs::path& path, std::string& out);
} // namespace conductino::detail

namespace {

std::string to_lower(std::string s) {
    std::transform(s.begin(), s.end(), s.begin(), [](unsigned char c) {
        return static_cast<char>(std::tolower(c));
    });
    return s;
}

std::string extension_of(const fs::path& p) {
    return to_lower(p.extension().string());
}

bool is_text_ext(const std::string& ext) {
    static const char* k[] = {
        ".txt", ".md", ".markdown", ".csv", ".tsv", ".json", ".jsonl",
        ".log", ".xml", ".html", ".htm", ".css", ".js", ".ts", ".py",
        ".c", ".cpp", ".h", ".hpp", ".go", ".rs", ".java", ".kt",
        ".yaml", ".yml", ".toml", ".ini", ".cfg", ".tex", ".bib",
    };
    for (auto* e : k) {
        if (ext == e) return true;
    }
    return false;
}

std::string cache_key_for(const fs::path& path) {
    std::error_code ec;
    auto sz = fs::file_size(path, ec);
    auto mt = fs::last_write_time(path, ec);
    auto mt_count = mt.time_since_epoch().count();
    std::ostringstream oss;
    oss << path.filename().string() << "_" << sz << "_" << mt_count;
    std::string s = oss.str();
    for (char& c : s) {
        if (!(std::isalnum(static_cast<unsigned char>(c)) || c == '_' || c == '-' || c == '.')) {
            c = '_';
        }
    }
    return s;
}

char* dup_string(const std::string& s, size_t* out_len) {
    char* buf = static_cast<char*>(::operator new(s.size() + 1));
    if (!buf) return nullptr;
    std::memcpy(buf, s.data(), s.size());
    buf[s.size()] = '\0';
    if (out_len) *out_len = s.size();
    return buf;
}

} // namespace

extern "C" {

CONDUCTINO_API int conductino_document_extract(
    const char* path,
    char** out_text,
    size_t* out_len) {
    if (!path || !out_text) return 1;
    *out_text = nullptr;
    if (out_len) *out_len = 0;
    if (!conductino::detail::ready()) return 2;

    fs::path p(path);
    std::error_code ec;
    if (!fs::exists(p, ec) || !fs::is_regular_file(p, ec)) {
        return 3;
    }

    const auto ext = extension_of(p);
    const fs::path cache_dir = conductino::detail::data_dir() / "cache" / "docs";
    fs::create_directories(cache_dir, ec);
    const fs::path cache_file = cache_dir / (cache_key_for(p) + ".txt");

    std::string cached;
    if (conductino::detail::read_text_file(cache_file, cached) && !cached.empty()) {
        *out_text = dup_string(cached, out_len);
        return *out_text ? 0 : 4;
    }

    if (is_text_ext(ext) || ext.empty()) {
        std::string body;
        if (!conductino::detail::read_text_file(p, body)) {
            return 5;
        }
        body.erase(std::remove(body.begin(), body.end(), '\0'), body.end());
        conductino::detail::write_text_file(cache_file, body);
        *out_text = dup_string(body, out_len);
        return *out_text ? 0 : 4;
    }

    if (ext == ".pdf" || ext == ".docx" || ext == ".doc") {
        std::ostringstream msg;
        msg << "[Conductino] Binary document detected (" << ext
            << "). Full extract requires an optional third-party library.\n"
            << "Path: " << p.string() << "\n"
            << "Workaround: export/copy text, or use Paste text in the study workspace.\n"
            << "See backend/features/document/README.md and docs/ai/ARCHITECTURE.md.";
        std::string body = msg.str();
        conductino::detail::write_text_file(cache_file, body);
        *out_text = dup_string(body, out_len);
        return *out_text ? 0 : 4;
    }

    return 6;
}

CONDUCTINO_API int conductino_document_import(
    const char* src_path,
    char** out_rel,
    size_t* out_len) {
    if (!src_path || !out_rel) return 1;
    *out_rel = nullptr;
    if (out_len) *out_len = 0;
    if (!conductino::detail::ready()) return 2;

    fs::path src(src_path);
    std::error_code ec;
    if (!fs::exists(src, ec) || !fs::is_regular_file(src, ec)) {
        return 3;
    }

    const fs::path imports = conductino::detail::data_dir() / "imports";
    fs::create_directories(imports, ec);

    auto now = std::chrono::system_clock::now().time_since_epoch().count();
    std::string name = std::to_string(now) + "_" + src.filename().string();
    for (char& c : name) {
        if (c == '/' || c == '\\' || c == ':') c = '_';
    }
    fs::path dest = imports / name;
    fs::copy_file(src, dest, fs::copy_options::overwrite_existing, ec);
    if (ec) return 4;

    std::string rel = (fs::path("imports") / name).generic_string();
    *out_rel = dup_string(rel, out_len);
    return *out_rel ? 0 : 5;
}

} // extern "C"
