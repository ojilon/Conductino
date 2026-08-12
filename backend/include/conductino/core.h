/**
 * Conductino backend — public C ABI
 *
 * Stable surface for Go (cgo) and other hosts.
 * Implementation is C++23; this header is pure C for linkage simplicity.
 *
 * RULE: this library does not perform network I/O.
 */
#ifndef CONDUCTINO_CORE_H
#define CONDUCTINO_CORE_H

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN32) && defined(CONDUCTINO_BUILDING)
#  define CONDUCTINO_API __declspec(dllexport)
#elif defined(_WIN32)
#  define CONDUCTINO_API __declspec(dllimport)
#else
#  define CONDUCTINO_API __attribute__((visibility("default")))
#endif

#include <stddef.h>
#include <stdint.h>

/** Library version string (static storage). */
CONDUCTINO_API const char* conductino_version(void);

/**
 * Initialize backend with a data directory path (UTF-8).
 * Creates the directory if needed. Returns 0 on success, non-zero on error.
 */
CONDUCTINO_API int conductino_init(const char* data_dir);

/** Shut down and flush. Safe to call multiple times. */
CONDUCTINO_API void conductino_shutdown(void);

/**
 * Save a note as JSON (UTF-8). Schema matches frontend NoteHighlightEvent.
 * Returns 0 on success.
 */
CONDUCTINO_API int conductino_notes_save_json(const char* json, size_t len);

/**
 * Search notes. Writes a JSON array into *out_json (malloc'd); caller frees
 * with conductino_free. Returns 0 on success.
 */
CONDUCTINO_API int conductino_notes_search(const char* query, char** out_json, size_t* out_len);

/** Free a buffer allocated by the library. */
CONDUCTINO_API void conductino_free(void* p);

/**
 * Get setting by key (UTF-8). Writes value into *out_value (malloc'd).
 * Returns 0 if found, 1 if missing, negative on error.
 */
CONDUCTINO_API int conductino_settings_get(const char* key, char** out_value, size_t* out_len);

/** Set setting key/value (UTF-8). Returns 0 on success. */
CONDUCTINO_API int conductino_settings_set(const char* key, const char* value);

#ifdef __cplusplus
}
#endif

#endif /* CONDUCTINO_CORE_H */
