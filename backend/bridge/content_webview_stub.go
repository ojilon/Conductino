package bridge

import "fmt"

// ContentEnsure initializes the content WebView2 surface.
// Currently Windows-only; on other platforms this is a no-op.
func ContentEnsure() error {
	return fmt.Errorf("content WebView2 is Windows-only")
}

// ContentNavigate delegates to the native navigate function.
func ContentNavigate(url string) error {
	return nil // placeholder - actual implementation in native code
}

// ContentSetVisible sets the content pane visibility.
func ContentSetVisible(show bool) error {
	return nil // placeholder - actual implementation in native code
}

// ContentResize resizes the content pane.
func ContentResize() error {
	return nil // placeholder
}

// ContentSetChromeHeight sets the chrome height for content pane.
func ContentSetChromeHeight(px int) error {
	return nil // placeholder
}

// ContentEval evaluates JavaScript in the content pane.
func ContentEval(js string) error {
	return nil // placeholder
}

// ContentLastURL returns the last URL loaded in the content pane.
func ContentLastURL() string {
	return ""
}

// ContentCopySelection copies the selected text from the content pane.
func ContentCopySelection() error {
	return nil // placeholder
}

// ContentRecreateAsPopup recreates content as a popup window.
// Currently Windows-only.
func ContentRecreateAsPopup() error {
	return fmt.Errorf("content WebView2 is Windows-only")
}