package handlers

import (
	"log"
	"net/http"
)

// API mounts local JSON endpoints that do NOT fetch remote pages.
// Page loading is exclusively the native webview's job (see docs/MIGRATION.md).
type API struct {
	Notes *NoteStore
}

func NewAPI() *API {
	return &API{Notes: NewNoteStore()}
}

// Register attaches routes on mux. Call before serving static files if you
// need API paths to take precedence, or use a dedicated prefix.
func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/notes", a.notesRouter)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"proxy":"retired"}`))
	})

	// Explicitly refuse the old proxy path so any leftover client code fails loudly.
	mux.HandleFunc("/api/proxy", retiredProxy)
	mux.HandleFunc("/api/plain_text", retiredProxy)
	mux.HandleFunc("/api/navigate", retiredProxy)
}

func (a *API) notesRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.Notes.SaveNoteHandler(w, r)
	case http.MethodGet:
		a.Notes.SearchNotesHandler(w, r)
	default:
		http.Error(w, "GET or POST required", http.StatusMethodNotAllowed)
	}
}

func retiredProxy(w http.ResponseWriter, r *http.Request) {
	log.Printf("[api] retired endpoint hit: %s %s", r.Method, r.URL.Path)
	http.Error(w,
		"This endpoint is retired. Remote pages load via the native webview (hostNavigate). See docs/MIGRATION.md.",
		http.StatusGone,
	)
}
