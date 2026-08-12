#include "conductino/core.h"

#include <cstdio>
#include <string>

int main() {
    std::printf("conductino_backend %s\n", conductino_version());

    if (conductino_init("./conductino-data") != 0) {
        std::fprintf(stderr, "init failed\n");
        return 1;
    }

    const char* sample =
        R"({"page_url":"https://example.com","page_title":"Example","selection":"hello","context":"","color":"#5ee7c4","coords":{"start_x":0,"start_y":0,"end_x":0,"end_y":0},"created_at":0})";
    if (conductino_notes_save_json(sample, std::char_traits<char>::length(sample)) != 0) {
        std::fprintf(stderr, "notes_save failed\n");
    }

    char* json = nullptr;
    size_t len = 0;
    if (conductino_notes_search("hello", &json, &len) == 0 && json != nullptr) {
        std::printf("search: %.*s\n", static_cast<int>(len), json);
        conductino_free(json);
    }

    conductino_settings_set("theme", "aurora-dark");
    char* theme = nullptr;
    size_t tlen = 0;
    if (conductino_settings_get("theme", &theme, &tlen) == 0 && theme != nullptr) {
        std::printf("theme: %.*s\n", static_cast<int>(tlen), theme);
        conductino_free(theme);
    }

    conductino_shutdown();
    return 0;
}
