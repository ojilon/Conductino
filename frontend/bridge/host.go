package bridge

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"frontend/shell"

	webview "github.com/webview/webview_go"
)

// Host owns the tab model and delegates surfaces to shell.
// Until dual-WebView exists, SingleSurfaceHost is used — remote
// Navigate still replaces chrome (docs/SHELL.md).
type Host struct {
	mu        sync.Mutex
	w         webview.WebView
	tabs      *TabManager
	chromeURL string
	shell     *shell.SingleSurfaceHost
}

func NewHost(chromeURL string) *Host {
	return &Host{
		tabs:      NewTabManager(),
		chromeURL: chromeURL,
		shell:     shell.NewSingleSurfaceHost(chromeURL),
	}
}

func (h *Host) SetWebView(w webview.WebView) {
	h.mu.Lock()
	h.w = w
	h.mu.Unlock()
	h.shell.SetWebView(w)
}

func (h *Host) destroy() {
	h.mu.Lock()
	w := h.w
	h.mu.Unlock()
	if w == nil {
		return
	}
	w.Dispatch(func() { w.Destroy() })
}

func (h *Host) PushTabsToChrome() {
	snap := h.tabs.Snapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		log.Printf("[bridge] marshal tabs: %v", err)
		return
	}
	js := fmt.Sprintf(`(function(){ if (window.ConductinoChrome && window.ConductinoChrome.applyTabSnapshot) { window.ConductinoChrome.applyTabSnapshot(%s); } })()`, string(b))
	h.shell.Eval(js)
}

func (h *Host) showChrome() {
	h.shell.LoadChrome(h.chromeURL)
}

func (h *Host) showContent(url string) {
	if url == "" {
		h.showChrome()
		return
	}
	h.shell.ShowRemoteContent(url)
}

func (h *Host) Bind(w webview.WebView) {
	w.Bind("hostPing", func() string { return "pong from Go host" })

	w.Bind("hostMinimize", func() {
		log.Printf("[host] minimize requested (wire via native handle later)")
	})
	w.Bind("hostMaximize", func() {
		log.Printf("[host] maximize/restore requested")
	})
	w.Bind("hostClose", func() {
		h.destroy()
	})

	w.Bind("hostTabNew", func() int {
		id := h.tabs.NewTab("New Tab", "")
		log.Printf("[host] tab new → %d", id)
		h.PushTabsToChrome()
		h.showChrome()
		return id
	})

	w.Bind("hostTabClose", func(id int) {
		log.Printf("[host] tab close %d", id)
		h.tabs.CloseTab(id)
		h.PushTabsToChrome()
		active := h.tabs.Active()
		if active == nil || active.URL == "" {
			h.showChrome()
			return
		}
		h.showContent(active.URL)
	})

	w.Bind("hostTabActivate", func(id int) {
		if !h.tabs.Activate(id) {
			return
		}
		log.Printf("[host] tab activate %d", id)
		h.PushTabsToChrome()
		active := h.tabs.Active()
		if active == nil || active.URL == "" {
			h.showChrome()
			return
		}
		h.showContent(active.URL)
	})

	w.Bind("hostTabList", func() string {
		b, _ := json.Marshal(h.tabs.Snapshot())
		return string(b)
	})

	w.Bind("hostNavigate", func(url string) {
		if url == "" {
			return
		}
		log.Printf("[host] native navigate → %s", url)
		h.tabs.Navigate(url)
		h.PushTabsToChrome()
		h.showContent(url)
	})

	w.Bind("hostGoBack", func() {
		url, ok := h.tabs.Back()
		if !ok {
			log.Printf("[host] goBack: nothing")
			return
		}
		log.Printf("[host] goBack → %s", url)
		h.PushTabsToChrome()
		h.showContent(url)
	})

	w.Bind("hostGoForward", func() {
		url, ok := h.tabs.Forward()
		if !ok {
			log.Printf("[host] goForward: nothing")
			return
		}
		log.Printf("[host] goForward → %s", url)
		h.PushTabsToChrome()
		h.showContent(url)
	})

	w.Bind("hostReload", func() {
		url := h.tabs.CurrentURL()
		if url == "" {
			h.showChrome()
			return
		}
		log.Printf("[host] reload → %s", url)
		h.shell.Content().Reload()
	})

	w.Bind("hostShowChrome", func() {
		log.Printf("[host] show chrome %s", h.chromeURL)
		h.showChrome()
		h.PushTabsToChrome()
	})
}
