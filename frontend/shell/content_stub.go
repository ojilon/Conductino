//go:build !windows

package shell

import "log"

type ContentWebView struct {
	lastURL string
	leftPad int32
}

func NewContentWebView() *ContentWebView { return &ContentWebView{} }

func (c *ContentWebView) Embed(parentHWND uintptr, dataDir string) bool {
	log.Printf("[content] Embed not implemented on this OS")
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

func (c *ContentWebView) SetLeftInset(px int32) {
	c.leftPad = px
	log.Printf("[content] left inset stub = %d", px)
}
