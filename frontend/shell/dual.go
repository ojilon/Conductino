package shell

import (
	"log"
	"sync"

	webview "github.com/webview/webview_go"
)

// DualHost keeps chrome on the primary webview (local shell only) and
// routes remote pages to a separate ContentWebView surface.
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

// AttachContent embeds the content WebView2 and locks the chrome surface.
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

// ShowLocalContent hides the remote content surface and shows a chrome panel.
// Never reloads the chrome document when dual is active.
func (h *DualHost) ShowLocalContent(path string) {
	log.Printf("[shell] local content %s (dual=%v)", path, h.DualActive())
	if h.DualActive() {
		h.content.Hide()
		// Drive panel visibility inside the already-loaded chrome document.
		switch := path
		if panel == "" {
			panel = "welcome"
		}
		js := `(function(){ if (window.ConductinoChrome && window.ConductinoChrome.showWelcome) {` +
			` window.ConductinoChrome.showWelcome(); } ` +
			`var w=document.getElementById('welcome'); if(w){ w.classList.add('active'); w.removeAttribute('hidden'); }` +
			`})()`
		if panel == "settings" {
			js = `(function(){ var s=document.getElementById('settings-panel');` +
				`['welcome','settings-panel','stub-panel'].forEach(function(id){` +
				`var n=document.getElementById(id); if(!n)return;` +
				`var on=id==='settings-panel'; n.classList.toggle('active',on);` +
				`if(on)n.removeAttribute('hidden'); else n.setAttribute('hidden','');});})()`
		}
		h.chrome.Eval(js)
		return
	}
	h.chrome.LoadChrome(h.chromeURL)
}

// ShowRemoteContent navigates only the content surface when dual is active.
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
	// Initial load only; no-op when locked.
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
