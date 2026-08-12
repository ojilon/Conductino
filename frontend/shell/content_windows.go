//go:build windows

package shell

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/go-webview2/pkg/edge"
)

// DefaultChromeTopPx is the reserved height for title + tabs + toolbar.
// Later this can be measured from the chrome document; for now a fixed band.
const DefaultChromeTopPx int32 = 128

// ContentWebView is a native WebView2 content surface (not the chrome document).
type ContentWebView struct {
	mu       sync.Mutex
	chromium *edge.Chromium
	bounds   Bounds
	ready    bool
	hidden   bool
	lastURL  string
}

// NewContentWebView prepares a content controller (call Embed before use).
func NewContentWebView() *ContentWebView {
	return &ContentWebView{hidden: true}
}

// Embed attaches a WebView2 controller to parent HWND (main window).
// Uses a separate user-data folder from the chrome webview.
// Must run on the UI thread; blocks until the controller is ready.
func (c *ContentWebView) Embed(parentHWND uintptr, dataDir string) bool {
	if parentHWND == 0 {
		log.Printf("[content] Embed: null HWND")
		return false
	}
	if dataDir == "" {
		dataDir = filepath.Join(os.TempDir(), "conductino-content-wv2")
	}
	_ = os.MkdirAll(dataDir, 0o755)

	cr := edge.NewChromium()
	cr.DataPath = dataDir
	cr.Debug = false
	// Note: do not call SetErrorCallback — it is absent on some go-webview2
	// tags (e.g. v1.0.2). Default handler logs; avoid relying on os.Exit paths
	// by treating Embed failure as dual-surface unavailable.

	log.Printf("[content] embedding WebView2 on HWND=%v data=%s", parentHWND, dataDir)
	ok := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[content] Embed panic: %v", r)
				ok = false
			}
		}()
		ok = cr.Embed(parentHWND)
	}()
	if !ok {
		log.Printf("[content] Embed failed")
		return false
	}

	// Reserve top band for chrome HTML (tabs / omnibox).
	cr.SetPadding(edge.Rect{Top: DefaultChromeTopPx})
	if err := cr.Hide(); err != nil {
		log.Printf("[content] Hide: %v", err)
	}

	c.mu.Lock()
	c.chromium = cr
	c.ready = true
	c.hidden = true
	c.mu.Unlock()

	log.Printf("[content] WebView2 ready (chrome top pad=%dpx)", DefaultChromeTopPx)
	return true
}

func (c *ContentWebView) Navigate(url string) {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	c.lastURL = url
	c.mu.Unlock()
	if !ready || cr == nil {
		log.Printf("[content] Navigate skipped (not ready): %s", url)
		return
	}
	log.Printf("[content] Navigate → %s", url)
	cr.Navigate(url)
	if err := cr.Show(); err != nil {
		log.Printf("[content] Show: %v", err)
	}
	c.mu.Lock()
	c.hidden = false
	c.mu.Unlock()
}

func (c *ContentWebView) Reload() {
	c.mu.Lock()
	url := c.lastURL
	cr := c.chromium
	ready := c.ready
	c.mu.Unlock()
	if !ready || cr == nil || url == "" {
		return
	}
	cr.Navigate(url)
}

func (c *ContentWebView) Hide() {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	c.mu.Unlock()
	if !ready || cr == nil {
		return
	}
	_ = cr.Hide()
	c.mu.Lock()
	c.hidden = true
	c.mu.Unlock()
}

func (c *ContentWebView) Show() {
	c.mu.Lock()
	cr := c.chromium
	ready := c.ready
	c.mu.Unlock()
	if !ready || cr == nil {
		return
	}
	_ = cr.Show()
	c.mu.Lock()
	c.hidden = false
	c.mu.Unlock()
}

func (c *ContentWebView) Bounds() Bounds {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.bounds
}

func (c *ContentWebView) SetBounds(b Bounds) {
	c.mu.Lock()
	c.bounds = b
	cr := c.chromium
	ready := c.ready
	c.mu.Unlock()
	if !ready || cr == nil {
		return
	}
	if b.Y > 0 {
		cr.SetPadding(edge.Rect{Top: int32(b.Y)})
	} else {
		cr.SetPadding(edge.Rect{Top: DefaultChromeTopPx})
	}
	cr.Resize()
}

func (c *ContentWebView) Ready() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ready
}
