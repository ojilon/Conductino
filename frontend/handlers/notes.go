package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"frontend/bridge"
)

// NoteHighlightEvent is the wire contract for study highlights/notes.
// Eventually persisted by the C++ backend; for now held in-process and
// forwarded to conductino_core when NativeAvailable().
type NoteHighlightEvent struct {
	PageURL   string `json:"page_url"`
	PageTitle string `json:"page_title"`
	Selection string `json:"selection"`
	Context   string `json:"context"`
	Color     string `json:"color"`
	Coords    Coords `json:"coords"`
	CreatedAt int64  `json:"created_at"`
}

type Coords struct {
	StartX int `json:"start_x"`
	StartY int `json:"start_y"`
	EndX   int `json:"end_x"`
	EndY   int `json:"end_y"`
}

type NoteStore struct {
	mu    sync.Mutex
	notes []NoteHighlightEvent
}

func NewNoteStore() *NoteStore {
	return &NoteStore{notes: make([]NoteHighlightEvent, 0)}
}

func (s *NoteStore) Add(n NoteHighlightEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes = append(s.notes, n)
}

func (s *NoteStore) Search(query string) []NoteHighlightEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if query == "" {
		out := make([]NoteHighlightEvent, len(s.notes))
		copy(out, s.notes)
		return out
	}
	var out []NoteHighlightEvent
	for _, n := range s.notes {
		if containsFold(n.Selection, query) || containsFold(n.Context, query) || containsFold(n.PageTitle, query) {
			out = append(out, n)
		}
	}
	return out
}

func containsFold(hay, needle string) bool {
	return len(needle) == 0 || (len(hay) > 0 && indexFold(hay, needle) >= 0)
}

func indexFold(s, substr string) int {
	ls, lsub := toLower(s), toLower(substr)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func (s *NoteStore) SaveNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var note NoteHighlightEvent
	if err := json.Unmarshal(body, &note); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if note.Selection == "" {
		http.Error(w, "selection required", http.StatusBadRequest)
		return
	}
	if note.CreatedAt == 0 {
		note.CreatedAt = time.Now().UnixMilli()
	}

	// Prefer C++ core when linked; always keep Go mirror for search while hybrid.
	if bridge.NativeAvailable() {
		clean, _ := json.Marshal(note)
		if err := bridge.NativeNotesSaveJSON(string(clean)); err != nil {
			log.Printf("[notes] native save failed: %v — falling back to memory", err)
		} else {
			log.Printf("[notes] saved via conductino_core")
		}
	}
	s.Add(note)
	log.Printf("[notes] saved selection=%q page=%s", trunc(note.Selection, 40), note.PageURL)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "created_at": note.CreatedAt, "native": bridge.NativeAvailable()})
}

func (s *NoteStore) SearchNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("query")

	if bridge.NativeAvailable() {
		if raw, err := bridge.NativeNotesSearch(q); err == nil && raw != "" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(raw))
			return
		}
	}

	results := s.Search(q)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"query":   q,
		"count":   len(results),
		"results": results,
	})
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
