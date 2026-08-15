package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	homeMu  sync.RWMutex
	homeURL = "http://wails.localhost"
)

// SetHomeURL is called once from the asset UI so we can return after browsing.
func (a *App) SetHomeURL(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		return
	}
	if i := strings.Index(url, "?"); i >= 0 {
		url = url[:i]
	}
	if i := strings.Index(url, "#"); i >= 0 {
		url = url[:i]
	}
	url = strings.TrimRight(url, "/")
	homeMu.Lock()
	homeURL = url
	homeMu.Unlock()
}

func (a *App) getHomeURL() string {
	homeMu.RLock()
	defer homeMu.RUnlock()
	return homeURL
}

// Navigate prefers the dual-webview content pane (chrome stays visible).
// Falls back to full-window navigation only if content pane is unavailable.
func (a *App) Navigate(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "about:") {
		url = "https://" + url
	}

	// Dual webview path (Windows)
	if err := a.ContentNavigate(url); err == nil {
		return nil
	} else {
		// log and fall through
		_ = err
	}

	// Fallback: navigate the main webview (loses chrome until GoHome)
	if a.ctx == nil {
		return fmt.Errorf("app not started")
	}
	b, _ := json.Marshal(url)
	runtime.WindowExecJS(a.ctx, fmt.Sprintf("window.location.href = %s;", string(b)))
	return nil
}

// GoHome returns focus to the asset UI and hides the content pane.
func (a *App) GoHome() error {
	_ = a.ContentSetVisible(false)
	if a.ctx == nil {
		return nil
	}
	// Stay on asset UI — do not reload if already home
	runtime.WindowExecJS(a.ctx, `if (window.ConductinoChrome) { /* stay */ }`)
	return nil
}
