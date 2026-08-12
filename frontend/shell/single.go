package shell

import (
	"log"
	"sync"

	webview "github.com/webview/webview_go"
)

// SingleSurfaceHost is the interim adapter: one webview plays both roles.
// Remote Navigate still replaces chrome — documented limitation.
// Replace with DualWebViewHost when content WebView2 is ready.
type SingleSurfaceHost struct {
	mu        sync.Mutex
	w         webview.WebView
	chromeURL string
}

func NewSingleSurfaceHost(chromeURL string) *SingleSurfaceHost {
	return &SingleSurfaceHost{chromeURL: chromeURL}
}

func (h *SingleSurfaceHost) SetWebView(w webview.WebView) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.w = w
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

func (h *SingleSurfaceHost) navigate(url string) {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Navigate(url) })
	})
}

// ChromeSurface
func (h *SingleSurfaceHost) LoadChrome(chromeURL string) {
	if chromeURL != "" {
		h.chromeURL = chromeURL
	}
	log.Printf("[shell] LoadChrome %s (single-surface)", h.chromeURL)
	h.navigate(h.chromeURL)
}

func (h *SingleSurfaceHost) Eval(js string) {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Eval(js) })
	})
}

func (h *SingleSurfaceHost) Bounds() Bounds { return Bounds{} }

func (h *SingleSurfaceHost) SetBounds(b Bounds) {}

// ContentSurface
func (h *SingleSurfaceHost) Navigate(url string) {
	log.Printf("[shell] content Navigate %s (WARNING: replaces chrome on single-surface)", url)
	h.navigate(url)
}

func (h *SingleSurfaceHost) Reload() {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() { wv.Eval(`location.reload()`) })
	})
}

func (h *SingleSurfaceHost) ShowLocalContent(path string) {
	// Until dual-surface: local panels live inside chrome document.
	log.Printf("[shell] ShowLocalContent %s (chrome document panels)", path)
	h.LoadChrome(h.chromeURL)
}

func (h *SingleSurfaceHost) ShowRemoteContent(url string) {
	h.Navigate(url)
}
