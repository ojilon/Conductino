package shell

// Shell is the persistent application frame: chrome always visible,
// content surface swappable (local state vs remote URL).
//
// Today webview_go only gives one surface; Host implementation still
// multiplexes both roles until a dual-WebView backend is plugged in.
// See docs/SHELL.md.

type Bounds struct {
	X, Y, Width, Height int
}

// ContentSurface loads either a remote URL (native) or a local app URL.
// Implementations must not use iframes for remote pages.
type ContentSurface interface {
	Navigate(url string)
	Reload()
	Bounds() Bounds
	SetBounds(b Bounds)
}

// ChromeSurface hosts the permanent UI (tabs, omnibox, sidebar).
// It should only ever load the local chrome origin.
type ChromeSurface interface {
	LoadChrome(chromeURL string)
	Eval(js string)
	Bounds() Bounds
	SetBounds(b Bounds)
}

// Host owns both surfaces and window-level actions.
type Host interface {
	Chrome() ChromeSurface
	Content() ContentSurface
	// ShowLocalContent shows welcome/settings inside the content area
	// without leaving the chrome document (once dual-surface exists).
	ShowLocalContent(path string)
	// ShowRemoteContent navigates the content surface only.
	ShowRemoteContent(url string)
}
