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
	"io"
	"net/http"
	"time"

    nhelper "Conductino/handlers/navigation_helper"
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
	Browser *nhelper.Browser
}

func NewBackendClient(baseURL string) *BackendClient {
	return &BackendClient{
		baseURL: baseURL,
		http: &http.Client{Timeout: 5 * time.Second},
		Browser: nhelper.NewBrowser(),
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