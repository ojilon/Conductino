package bridge

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	webview "github.com/webview/webview_go"
)

// Host owns the native webview, tab model, and chrome URL for return-to-shell.
type Host struct {
	mu        sync.Mutex
	w         webview.WebView
	tabs      *TabManager
	chromeURL string
}

func NewHost(chromeURL string) *Host {
	return &Host{
		tabs:      NewTabManager(),
		chromeURL: chromeURL,
	}
}

func (h *Host) SetWebView(w webview.WebView) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.w = w
}

func (h *Host) with(fn func(webview.WebView)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.w != nil {
		fn(h.w)
	}
}

func (h *Host) navigateNative(url string) {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() {
			wv.Navigate(url)
		})
	})
}

func (h *Host) eval(js string) {
	h.with(func(wv webview.WebView) {
		wv.Dispatch(func() {
			wv.Eval(js)
		})
	})
}

// PushTabsToChrome serializes tab snapshot into the chrome JS model.
func (h *Host) PushTabsToChrome() {
	snap := h.tabs.Snapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		log.Printf("[bridge] marshal tabs: %v", err)
		return
	}
	// Only works while chrome document is loaded.
	js := fmt.Sprintf(`(function(){ if (window.ConductinoChrome && window.ConductinoChrome.applyTabSnapshot) { window.ConductinoChrome.applyTabSnapshot(%s); } })()`, string(b))
	h.eval(js)
}

// Bind registers all Go functions exposed to the chrome (and later content scripts).
func (h *Host) Bind(w webview.WebView) {
	w.Bind("hostPing", func() string { return "pong from Go host" })

	// —— Window ——
	w.Bind("hostMinimize", func() {
		log.Printf("[host] minimize requested (wire via native handle later)")
	})
	w.Bind("hostMaximize", func() {
		log.Printf("[host] maximize/restore requested")
	})
	w.Bind("hostClose", func() {
		h.with(func(wv webview.WebView) {
			wv.Dispatch(func() { wv.Destroy() })
		})
	})

	// —— Tabs (Go is source of truth) ——
	w.Bind("hostTabNew", func() int {
		id := h.tabs.NewTab("New Tab", "")
		log.Printf("[host] tab new → %d", id)
		h.PushTabsToChrome()
		// Stay on chrome for empty tab.
		h.navigateNative(h.chromeURL)
		return id
	})

	w.Bind("hostTabClose", func(id int) {
		log.Printf("[host] tab close %d", id)
		h.tabs.CloseTab(id)
		h.PushTabsToChrome()
		active := h.tabs.Active()
		if active == nil || active.URL == "" {
			h.navigateNative(h.chromeURL)
			return
		}
		h.navigateNative(active.URL)
	})

	w.Bind("hostTabActivate", func(id int) {
		if !h.tabs.Activate(id) {
			return
		}
		log.Printf("[host] tab activate %d", id)
		h.PushTabsToChrome()
		active := h.tabs.Active()
		if active == nil || active.URL == "" {
			h.navigateNative(h.chromeURL)
			return
		}
		h.navigateNative(active.URL)
	})

	w.Bind("hostTabList", func() string {
		b, _ := json.Marshal(h.tabs.Snapshot())
		return string(b)
	})

	// —— Navigation (drives native webview) ——
	w.Bind("hostNavigate", func(url string) {
		if url == "" {
			return
		}
		log.Printf("[host] native navigate → %s", url)
		h.tabs.Navigate(url)
		h.PushTabsToChrome()
		h.navigateNative(url)
	})

	w.Bind("hostGoBack", func() {
		url, ok := h.tabs.Back()
		if !ok {
			log.Printf("[host] goBack: nothing")
			return
		}
		log.Printf("[host] goBack → %s", url)
		h.PushTabsToChrome()
		h.navigateNative(url)
	})

	w.Bind("hostGoForward", func() {
		url, ok := h.tabs.Forward()
		if !ok {
			log.Printf("[host] goForward: nothing")
			return
		}
		log.Printf("[host] goForward → %s", url)
		h.PushTabsToChrome()
		h.navigateNative(url)
	})

	w.Bind("hostReload", func() {
		url := h.tabs.CurrentURL()
		if url == "" {
			log.Printf("[host] reload: on chrome shell")
			h.navigateNative(h.chromeURL)
			return
		}
		log.Printf("[host] reload → %s", url)
		h.navigateNative(url)
	})

	w.Bind("hostShowChrome", func() {
		log.Printf("[host] show chrome %s", h.chromeURL)
		h.navigateNative(h.chromeURL)
		// After chrome loads, JS will request tab list; also try push shortly.
		h.PushTabsToChrome()
	})
}
