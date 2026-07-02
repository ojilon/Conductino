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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"   //local-PDF text edtractor
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
	Browser *Browser
}

func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL: baseURL,
		http: &http.Client{Timeout: 5 * time.Second},
		Browser: NewBrowser(),
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

	//validate URL
	if _, err := url.ParseRequestURI(targetURL); err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}

	/*proposed future refactor
	func NewHTTPClient() *http.Client {

		return &http.Client{

			CheckRedirect: func(req *http.Request, via []*http.Request) error {

				log.Println("Redirect ->", req.URL.String())

				return nil
			},
		}
	}

	then client becomes
	client := NewHTTPClient()

	to lATER add things like 
	cookies, proxy support, custom User-Agent, TLS settings without
	touching the function ProxyHAndler
	*/

	//create the request
	req, err := http.NewRequest(
		http.MethodGet,
		targetURL,
		nil,
	)	
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Set required custom headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

    //execute request to get response
    resp, err := c.Browser.Do(req)
    if err != nil {
    	http.Error(w, err.Error(), http.StatusBadGateway)
    	return
    }
    defer resp.Body.Close()

	//response debugging added
	log.Println("Final URL     :", resp.Request.URL.String())
	log.Println("Status        :", resp.Status)
	log.Println("Content-Type  :", resp.Header.Get("Content-Type"))

	// 3. Remove headers that stop embedding in Webview
	resp.Header.Del("X-Frame-Options")
	// TODO:
    // Instead of deleting CSP completely,
    // rewrite it where possible.
    // Removing it permanently reduces security.
	resp.Header.Del("Content-Security-Policy")


	// Stream body directly
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//handle rewriting links to the resources of downloaded response
	contentType := resp.Header.Get("Content-Type")
	action := DecideResponse(contentType)
	switch action {
	case ActionDisplayHTML:
		baseURL := resp.Request.URL
		body, err = RewriteHTML(body, baseURL)
		if err != nil {
			log.Println(err)
		}

		for k, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}

		w.Header().Set(
		    "Content-Type",
		    "text/html; charset=utf-8",
		)

		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Write error:", err)
			return
		}
	case ActionStream:
		w.Header().Set(
			"Content-Type",
			resp.Header.Get("Content-Type"),
		)
		_, err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Write error:", err)
			return
		}
	case ActionDownload:
		for k, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}

		w.Header().Set(
			"Content-Disposition",
			"attachment",
		)
		_,err = w.Write(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Write error:", err)
			return
		}
	default:
		return
	}
}


/*
================================================================================
FUTURE ARCHITECTURE NOTE - NAVIGATION CONTEXT
================================================================================

Current ProxyHandler() works using several independent local variables:

    targetURL
    contentType
    statusCode
    body
    resp

This is fine while the browser is small.

However, as more browser features are added, these values will need to be shared
between many parts of the program. Instead of passing many separate variables,
create a single structure representing the current page navigation.

Example:

type PageContext struct {
    TargetURL   *url.URL
    FinalURL    *url.URL
    StatusCode  int
    ContentType string
    Body        []byte
}

Future fields may include:

    Method          string
    RequestHeaders  http.Header
    ResponseHeaders http.Header
    Cookies         []*http.Cookie
    Referrer        string
    RedirectChain   []string
    DownloadedAt    time.Time
    IsHTML          bool
    TabID           int
    NavigationID    int

Eventually ProxyHandler() becomes something like:

    ctx := NewPageContext(targetURL)

    DownloadPage(ctx)
    RewriteHTML(ctx)
    RecordHistory(ctx)
    DetectDownloads(ctx)
    UpdateTab(ctx)
    SendResponse(ctx)

Benefits:

- One object represents the entire navigation.
- Easier to pass information between browser components.
- Future features won't require changing many function signatures.
- Makes debugging easier because one structure contains everything about
  the current page.
- Similar design is used in real browsers, where a navigation object
  carries request and response state through the loading pipeline.

================================================================================
Possible future browser pipeline

User enters URL
        │
        ▼
Normalize URL
        │
        ▼
Create PageContext
        │
        ▼
Send HTTP Request
        │
        ▼
Receive Response
        │
        ▼
Detect Content-Type
        │
        ▼
If HTML
    Parse HTML
        │
        ▼
Walk DOM Tree
        │
        ▼
Rewrite Resources
        │
        ▼
Render HTML
Else
    Stream Resource Directly
        │
        ▼
Update Tab State
        │
        ▼
Record History
        │
        ▼
Handle Downloads
        │
        ▼
Send Final Response to WebView

================================================================================
Future browser features that can use PageContext

Navigation
    - Back / Forward
    - Reload
    - Tabs
    - Split View

History
    - Visited pages
    - Search history
    - Most visited
    - Typed URLs

Downloads
    - Download manager
    - Pause / Resume
    - Progress tracking

Security
    - HTTPS information
    - Certificate details
    - Mixed content detection
    - Safe browsing checks

Study Features
    - AI summarization
    - Highlight storage
    - Notes
    - Bookmarks
    - Reading progress
    - Offline page storage

Networking
    - Redirect tracking
    - Cookies
    - Cache
    - Compression
    - Custom headers
    - User-Agent switching

This is NOT necessary yet.
Keep ProxyHandler simple while learning HTTP.
Introduce PageContext only after networking becomes stable and additional
browser features begin sharing the same navigation information.
================================================================================
*/