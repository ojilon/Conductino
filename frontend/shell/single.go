package shell

import (
	"log"
	"strings"
	"sync"

	webview "github.com/webview/webview_go"
)

// SingleSurfaceHost drives the chrome webview.
// When Locked is true (dual surface active), Navigate to non-chrome
// URLs is refused so the shell document stays at chromeURL forever.
type SingleSurfaceHost struct {
	mu        sync.Mutex
	w         webview.WebView
	chromeURL string
	locked    bool // true once dual content surface is active
}

func NewSingleSurfaceHost(chromeURL string) *SingleSurfaceHost {
	return &SingleSurfaceHost{chromeURL: chromeURL}
}

func (h *SingleSurfaceHost) SetWebView(w webview.WebView) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.w = w
}

// LockChrome prevents any Navigate away from the chrome URL.
func (h *SingleSurfaceHost) LockChrome() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.locked = true
	log.Printf("[shell] chrome LOCKED to %s", h.chromeURL)
}

func (h *SingleSurfaceHost) Locked() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.locked
}

func (h *SingleSurfaceHost) Chrome() ChromeSurface   { return h }
func (h *SingleSurfaceHost) Content() ContentSurface { return h }

func (h *SingleSurfaceHost) with(fn func(webview.WebView)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.w != nil {
		fn(h.w)
	}
}

func (h *SingleSurfaceHost) isChromeURL(url string) bool {
	h.mu.Lock()
	base := h.chromeURL
	h.mu.Unlock()
	if url == "" || base == "" {
		return false
	}
	// Accept chrome origin with optional trailing path/query.
	u := strings.TrimRight(url, "/")
	b := strings.TrimRight(base, "/")
	return u == b || strings.HasPrefix(url, b+"/") || strings.HasPrefix(url, b+"?")
}

func (h *SingleSurfaceHost) navigate(url string) {
	h.mu.Lock()
	locked := h.locked
	chrome := h.chromeURL
	h.mu.Unlock()
	if locked && !h.isChromeURL(url) {
		log.Printf("[shell] BLOCKED chrome Navigate to %s (locked to %s)", url, chrome)
		return
	}
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Navigate(url) })
	})
}

// LoadChrome loads the shell once. When locked, only Eval-friendly no-op if already there.
func (h *SingleSurfaceHost) LoadChrome(chromeURL string) {
	if chromeURL != "" {
		h.mu.Lock()
		h.chromeURL = chromeURL
		h.mu.Unlock()
	}
	h.mu.Lock()
	locked := h.locked
	url := h.chromeURL
	h.mu.Unlock()
	if locked {
		// Do not reload — that would wipe JS tab state. Shell stays put.
		log.Printf("[shell] LoadChrome skipped (locked, already on %s)", url)
		return
	}
	log.Printf("[shell] LoadChrome %s", url)
	h.navigate(url)
}

func (h *SingleSurfaceHost) Eval(js string) {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Eval(js) })
	})
}

func (h *SingleSurfaceHost) Bounds() Bounds { return Bounds{} }

func (h *SingleSurfaceHost) SetBounds(b Bounds) {}

// Navigate is only used as ContentSurface fallback when dual is inactive.
func (h *SingleSurfaceHost) Navigate(url string) {
	h.mu.Lock()
	locked := h.locked
	h.mu.Unlock()
	if locked {
		log.Printf("[shell] content Navigate via chrome surface BLOCKED (dual should own content): %s", url)
		return
	}
	log.Printf("[shell] content Navigate %s (single-surface fallback)", url)
	h.navigate(url)
}

func (h *SingleSurfaceHost) Reload() {
	h.mu.Lock()
	locked := h.locked
	h.mu.Unlock()
	if locked {
		// Reload would refresh chrome shell only — rarely wanted.
		log.Printf("[shell] chrome Reload skipped while locked")
		return
	}
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Eval(`location.reload()`) })
	})
}

func (h *SingleSurfaceHost) ShowLocalContent(path string) {
	log.Printf("[shell] ShowLocalContent %s", path)
	h.LoadChrome(h.chromeURL)
}

func (h *SingleSurfaceHost) ShowRemoteContent(url string) {
	h.Navigate(url)
}
