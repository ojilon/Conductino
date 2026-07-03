package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
    xhtml "golang.org/x/net/html" //streaming HTMl% tokenizer
)

/*
-------------------/api/archive----------------------------------------
POST {"url":"..."} -downloads a remote page, tokenizes its HTML with
golang.org/x/net/html, strips scripts/iframes, and stores the clean text
via the Zig backend. Demonstrtes the html tokenizer use case.
*/
func (c *BackendClient) ArchiveHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "POST required", http.StatusMethodNotAllowed)
        return
    }
    var req struct{ URL string `json:"url"`}
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if _, err := url.Parse(req.URL); err != nil {
        http.Error(w, "bad url", http.StatusBadRequest)
        return
    }

    resp, err := http.Get(req.URL)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer resp.Body.Close()

    /*
    ---------------------------------- golang.org/x/net/html ----------------------------
    A streaming tokenizer means we don't have to lead the whole document
    into memory -useful for large web archives.
    */
    tz := xhtml.NewTokenizer(resp.Body)
    var text strings.Builder
    skip := false
    for {
        tt := tz.Next()
        if tt == xhtml.ErrorToken {
            break
        }
        tok := tz.Token()
        switch tok.Data {
        case "script", "style", "iframe":
            skip = tt == xhtml.StartTagToken
        }
        if tt == xhtml.TextToken && !skip {
            text.WriteString(strings.TrimSpace(tok.Data))
            text.WriteByte(' ')
        }
    }

    //Hand the cleaned text to zig as a "note" so it's indexed by FTS5.
    pkt := NoteHiglightEvent{
        PageURL: req.URL,
        PageTitle: "[archived] " + req.URL,
        Selection: text.String(),
        Context: "(full-page archive)",
        Color: "#0f80cc",
        CreatedAt: time.Now().UnixMilli(),
    }
    out, _ := json.Marshal(pkt)
    r.Body = io.NopCloser(bytes.NewReader(out))
    r.ContentLength = int64(len(out))
    c.forward(w, r, "/save")
}
