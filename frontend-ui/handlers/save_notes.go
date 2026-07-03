package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// ------/api/save/note ---------------------------------
//POST body: NoteHighlightEvent (JSON). Validate basic shape then forwards.
func (c *BackendClient) SaveNoteHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "POST required", http.StatusMethodNotAllowed)
        return
    }

    //Decode-validate-re-encode so a malformed packet never reaches Zig.
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    var note NoteHiglightEvent
    if err := json.Unmarshal(body, &note); err != nil {
        http.Error(w, "invalid JSON: " + err.Error(), http.StatusBadRequest)
        return
    }
    if note.CreatedAt == 0 {
        note.CreatedAt = time.Now().Local().UnixMilli()
    }

    //Re-encode the cleaned packet and forward to Zig.
    clean, _ := json.Marshal(note)
    r.Body = io.NopCloser(bytes.NewReader(clean))
    r.ContentLength = int64(len(clean))
    c.forward(w, r, "/save")
}

