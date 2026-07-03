package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"
)

/*
------------------ /api/pdf -------------------------------------------
PSOT {"path": "./papers/name.pdf"} - extractsplain text via ledongthuc/pdf
and formats each page to the zig backend.
*/
func ( c *BackendClient) PDFHandler (w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var req struct{Path string `json:"path"`}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	abs, _ := filepath.Abs(req.Path)
	f, err := os.Open(abs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	stat, _ := f.Stat()

	// ----- github.com/ledongthuc/pdf ---------------------
	doc, err := pdf.NewReader(f, stat.Size())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	results := make([]string, 0, doc.NumPage())
	for i := 1; i <= doc.NumPage(); i++ {
		page := doc.Page(i)
		if page.V.IsNull(){
			continue
		}
		txt, _ := page.GetPlainText(nil)
		results = append(results, fmt.Sprintf("------ page %d -------\n%s", i, txt))
	}

	//Bundle everything into a single Zig save_note call.
	pkt := NoteHiglightEvent{
		PageURL: "file:// " + abs,
		PageTitle: filepath.Base(abs),
		Selection: strings.Join(results, "\n\n"),
		Context: fmt.Sprintf("(%d pages, %.1f KB)", doc.NumPage(), float64(stat.Size())/1024),
		Color: "#f7a41d",
		CreatedAt: time.Now().UnixMilli(),
	}
	out, _ := json.Marshal(pkt)
	r.Body = io.NopCloser(bytes.NewReader(out))
	r.ContentLength = int64(len(out))
	c.forward(w, r, "/save")
}
