//go:build !windows

package shell

import "log"

// ContentWebView is a stub outside Windows — dual surface is Windows-first.
type ContentWebView struct {
	lastURL string
}

func NewContentWebView() *ContentWebView { return &ContentWebView{} }

func (c *ContentWebView) Embed(parentHWND uintptr, dataDir string) bool {
	log.Printf("[content] Embed not implemented on this OS (Windows-only step 1)")
	return false
}

func (c *ContentWebView) Navigate(url string) {
	c.lastURL = url
	log.Printf("[content] Navigate stub: %s", url)
}

func (c *ContentWebView) Reload() {}
func (c *ContentWebView) Hide()   {}
func (c *ContentWebView) Show()   {}
func (c *ContentWebView) Bounds() Bounds { return Bounds{} }
func (c *ContentWebView) SetBounds(b Bounds) {}
func (c *ContentWebView) Ready() bool { return false }
