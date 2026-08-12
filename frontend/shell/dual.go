package shell

import (
	"log"
	"sync"

	webview "github.com/webview/webview_go"
)

// DualHost keeps chrome on the primary webview (local shell only) and
// routes remote pages to a separate ContentWebView surface.
//
// On Windows, ContentWebView is a second WebView2 with top padding so the
// chrome band stays visible. On other OSes Embed fails and remote falls
// back to single-surface behaviour with a log warning.
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

// AttachContent embeds the content WebView2 on the main window HWND.
// Call after the chrome webview window exists (Window() != nil).
func (h *DualHost) AttachContent(parentHWND uintptr, dataDir string) bool {
	ok := h.content.Embed(parentHWND, dataDir)
	h.mu.Lock()
	h.dualOK = ok
	h.mu.Unlock()
	if ok {
		log.Printf("[shell] dual surface active — chrome permanent, content separate")
	} else {
		log.Printf("[shell] dual surface unavailable — falling back to single-surface warnings")
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

func (h *DualHost) ShowLocalContent(path string) {
	log.Printf("[shell] local content %s", path)
	if h.DualActive() {
		h.content.Hide()
	}
	// Chrome document already has welcome/settings panels.
	h.chrome.LoadChrome(h.chromeURL)
}

func (h *DualHost) ShowRemoteContent(url string) {
	if url == "" {
		h.ShowLocalContent("")
		return
	}
	if h.DualActive() {
		// Critical: do NOT navigate the chrome webview.
		h.content.Navigate(url)
		return
	}
	log.Printf("[shell] dual inactive — remote will replace chrome (legacy)")
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
