// handlers/api.go — Bridges UI actions to the Zig backend over HTTP/JSON.
//
// ─────────────────────────────────────────────────────────────────────────────
// LIBRARIES USED IN THIS FILE
// ─────────────────────────────────────────────────────────────────────────────
//   net/http (stdlib)            →  Talks to the Zig backend (127.0.0.1:8081)
//   encoding/json (stdlib)       →  Wire format for IPC packets
//   golang.org/x/net/html        →  Streaming HTML5 tokenizer used by
//                                   ArchiveHandler when saving a web page
//                                   for offline study.
//   github.com/ledongthuc/pdf    →  Extracts text + page coordinates from
//                                   local PDF files in PDFHandler.
// ─────────────────────────────────────────────────────────────────────────────

package handlers

import(
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf" //local-PDF text edtractor
	xhtml "golang.org/x/net/html" //streaming HTMl% tokenizer
)

/*
-------------------IPC packets -----------------------------
These structs are the wire contract between Go and Zig. The Zig backend
re-parses the JSON into its own `Note` struct in document.zig.
Keeping the contract small & explicit is the reason for not using
protobuf here.
*/
type NoteHiglightEvent struct {
	PageURL string `json:"page_url"`
	PageTitle string `json:"page_title"`
	Selection string `json:"selection"`
	Context string   `json:"context"`
	Color string      `json:"color"`  //"#e94560" etc.
	Coords Coords     `json:"coords"`
	CreatedAt int64    `json:"created_at"` //unix millis
}

type Coords struct {
	StartX int  `json:"start_x"`
	StartY int   `json:"start_y"`
	EndX   int    `json:"end_x"`
	EndY   int    `json:"end_y"`
}

type SearchResults struct {
	Query string   `json:"query"`
	Count int       `json:"count"`
	Results []NoteHiglightEvent `json:"results"`
}

//-------Backend client ------------------------------------

type BackendClient struct {
	baseURL string
	http *http.Client
}

func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL: baseURL,
		http: &http.Client{Timeout: 5 * time.Second},
	}
}

/*
forward is a generic helper: it copies the request body to the Zig backend
and pipes the response back to the WebView2 page unchanges. This keeps the 
Go side dumb - the Zig backedn is the single source of truth for storage.
*/
func (c *BackendClient) forward(w http.ResponseWriter, r *http.Request, path string) {
	target := c.baseURL + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, target, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := c.http.Do(req)
	if err != nil {
		http.Error(w, "backend unreachable: " + err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k,v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}


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


// ---------/api/search -------------------------------------------
//GEt ?query=memory - passes through to the Zig FTS5 endpoint.
func (c *BackendClient) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query().Get("query")
	if strings.TrimSpace(q) == "" {
		http.Error(w , "query param required", http.StatusBadRequest)
		return
	}
	//query is already in r.URL.RawQuesry; forward will pass it along.
	c.forward(w, r, "/search")
}

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


//--------------api/proxy -----------------------
func (c *BackendClient) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", http.StatusMethodNotAllowed)
		return
	}

	// 1. Read ?url=
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// Optional: validate URL
	if _, err := url.ParseRequestURI(targetURL); err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	// 2. Download the page
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 3. Remove headers that stop embedding
	resp.Header.Del("X-Frame-Options")
	resp.Header.Del("Content-Security-Policy")

	// 4. Copy remaining headers
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	// Send same status code
	w.WriteHeader(resp.StatusCode)

	// Stream body directly
	io.Copy(w, resp.Body)
}