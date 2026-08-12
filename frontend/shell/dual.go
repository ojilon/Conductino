package shell

import (
	"log"
	"sync"

	webview "github.com/webview/webview_go"
)

type DualHost struct {
	mu        sync.Mutex
	chromeURL string
	chrome    *SingleSurfaceHost
	content   *ContentWebView
	dualOK    bool
}

func NewDualHost(chromeURL string) *DualHost {
	return &DualHost{
		chromeURL: chromeURL,
		chrome:    NewSingleSurfaceHost(chromeURL),
		content:   NewContentWebView(),
	}
}

func (h *DualHost) SetChromeWebView(w webview.WebView) {
	h.chrome.SetWebView(w)
}

func (h *DualHost) AttachContent(parentHWND uintptr, dataDir string) bool {
	ok := h.content.Embed(parentHWND, dataDir)
	h.mu.Lock()
	h.dualOK = ok
	h.mu.Unlock()
	if ok {
		h.chrome.LockChrome()
		log.Printf("[shell] dual surface active — chrome permanent at %s", h.chromeURL)
	} else {
		log.Printf("[shell] dual surface unavailable — single-surface fallback")
	}
	return ok
}

func (h *DualHost) DualActive() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.dualOK
}

func (h *DualHost) Chrome() ChromeSurface { return h.chrome }

func (h *DualHost) Content() ContentSurface {
	if h.DualActive() {
		return h.content
	}
	return h.chrome
}

// SetSidebarOpen insets the content surface so the chrome sidebar is visible.
func (h *DualHost) SetSidebarOpen(open bool) {
	if !h.DualActive() {
		return
	}
	var inset int32
	if open {
		inset = DefaultSidebarWidthPx
	}
	h.content.SetLeftInset(inset)
}

func (h *DualHost) ShowLocalContent(path string) {
	log.Printf("[shell] local content %s (dual=%v)", path, h.DualActive())
	if h.DualActive() {
		h.content.Hide()
		js := `(function(){ if (window.ConductinoChrome && window.ConductinoChrome.showWelcome) {` +
			` window.ConductinoChrome.showWelcome(); } ` +
			`var w=document.getElementById('welcome'); if(w){ w.classList.add('active'); w.removeAttribute('hidden'); }` +
			`})()`
		h.chrome.Eval(js)
		return
	}
	h.chrome.LoadChrome(h.chromeURL)
}

func (h *DualHost) ShowRemoteContent(url string) {
	if url == "" {
		h.ShowLocalContent("welcome")
		return
	}
	if h.DualActive() {
		log.Printf("[shell] remote → content surface only: %s", url)
		h.content.Navigate(url)
		return
	}
	log.Printf("[shell] dual inactive — remote falls back to single surface")
	h.chrome.ShowRemoteContent(url)
}

func (h *DualHost) LoadChrome() {
	h.chrome.LoadChrome(h.chromeURL)
}

func (h *DualHost) EvalChrome(js string) {
	h.chrome.Eval(js)
}

func (h *DualHost) ReloadContent() {
	if h.DualActive() {
		h.content.Reload()
		return
	}
	h.chrome.Reload()
}
