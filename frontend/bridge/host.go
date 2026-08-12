package bridge

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"unsafe"

	"frontend/shell"

	webview "github.com/webview/webview_go"
)

// Host owns tabs and a DualHost shell (chrome + content surfaces).
type Host struct {
	mu        sync.Mutex
	w         webview.WebView
	tabs      *TabManager
	chromeURL string
	shell     *shell.DualHost
}

func NewHost(chromeURL string) *Host {
	return &Host{
		tabs:      NewTabManager(),
		chromeURL: chromeURL,
		shell:     shell.NewDualHost(chromeURL),
	}
}

func (h *Host) SetWebView(w webview.WebView) {
	h.mu.Lock()
	h.w = w
	h.mu.Unlock()
	h.shell.SetChromeWebView(w)
}

// AttachContentSurface embeds the Windows content WebView2 on the main HWND.
// No-op / false on unsupported platforms.
func (h *Host) AttachContentSurface(dataDir string) bool {
	h.mu.Lock()
	w := h.w
	h.mu.Unlock()
	if w == nil {
		return false
	}
	hwnd := uintptr(w.Window())
	if hwnd == 0 {
		log.Printf("[host] AttachContentSurface: null window handle")
		return false
	}
	return h.shell.AttachContent(hwnd, dataDir)
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
	h.shell.EvalChrome(js)
}

func (h *Host) showChrome() {
	if h.shell.DualActive() {
		// Hide remote content; chrome document stays loaded.
		h.shell.ShowLocalContent("welcome")
		return
	}
	h.shell.LoadChrome()
}

func (h *Host) showContent(url string) {
	h.shell.ShowRemoteContent(url)
}

func (h *Host) Bind(w webview.WebView) {
	w.Bind("hostPing", func() string {
		if h.shell.DualActive() {
			return "pong dual-surface"
		}
		return "pong single-surface"
	})

	w.Bind("hostMinimize", func() {
		log.Printf("[host] minimize requested")
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
		log.Printf("[host] navigate → %s (dual=%v)", url, h.shell.DualActive())
		h.tabs.Navigate(url)
		h.PushTabsToChrome()
		h.showContent(url)
	})

	w.Bind("hostGoBack", func() {
		url, ok := h.tabs.Back()
		if !ok {
			return
		}
		log.Printf("[host] goBack → %s", url)
		h.PushTabsToChrome()
		h.showContent(url)
	})

	w.Bind("hostGoForward", func() {
		url, ok := h.tabs.Forward()
		if !ok {
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
		h.shell.ReloadContent()
	})

	w.Bind("hostShowChrome", func() {
		h.showChrome()
		h.PushTabsToChrome()
	})

	// Silence unused import on non-cgo toolchains that still typecheck unsafe.
	_ = unsafe.Sizeof(0)
}
