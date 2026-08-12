/**
 * Storage helpers (file paths under data_dir).
 * Full SQLite integration lands later — see features/storage/README.md.
 */

#include <filesystem>
#include <fstream>
#include <string>

namespace conductino::detail {

std::filesystem::path data_dir();
bool ready();

std::filesystem::path path_under(const std::string& relative) {
    return data_dir() / relative;
}

bool write_text_file(const std::filesystem::path& path, const std::string& body) {
    if (!ready()) {
        return false;
    }
    try {
        if (path.has_parent_path()) {
            std::filesystem::create_directories(path.parent_path());
        }
        std::ofstream out(path, std::ios::binary | std::ios::trunc);
        if (!out) {
            return false;
        }
        out.write(body.data(), static_cast<std::streamsize>(body.size()));
        return static_cast<bool>(out);
    } catch (...) {
        return false;
    }
}

bool read_text_file(const std::filesystem::path& path, std::string& out) {
    try {
        std::ifstream in(path, std::ios::binary);
        if (!in) {
            return false;
        }
        out.assign(std::istreambuf_iterator<char>(in), std::istreambuf_iterator<char>());
        return true;
    } catch (...) {
        return false;
    }
}

} // namespace conductino::detail
