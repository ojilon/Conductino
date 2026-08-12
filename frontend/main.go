package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"frontend/pathutil"

	webview "github.com/webview/webview_go"
)

// Host holds the webview instance so bound callbacks can control it.
type Host struct {
	mu sync.Mutex
w  webview.WebView
}

func (h *Host) set(w webview.WebView) {
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

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Prefer running from repo root or from frontend/.
	cfgPath, err := pathutil.FindFile(cwd, "config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve web asset directory (frontend/web).
	webDir := resolveWebDir(cwd)
	log.Printf("[frontend] serving chrome from %s", webDir)

	go startStaticServer(cfg.IPC.FrontendListen, webDir)

	host := &Host{}

	w := webview.New(cfg.Window.Debug)
	defer w.Destroy()
	host.set(w)

	w.SetTitle(cfg.Window.Title)
	w.SetSize(cfg.Window.Width, cfg.Window.Height, webview.HintNone)

	// Bind chrome → native actions.
	// Navigation uses the *same* native webview (no iframe for remote pages).
	bindHost(w, host)

	// Load local chrome shell over loopback HTTP (real origin, not file://).
	chromeURL := fmt.Sprintf("http://%s/", cfg.IPC.FrontendListen)
	log.Printf("[frontend] navigating to chrome %s", chromeURL)
	w.Navigate(chromeURL)

	w.Run()
}

func bindHost(w webview.WebView, host *Host) {
	// Diagnostic
	w.Bind("hostPing", func() string { return "pong from Go host" })

	// Window controls (best-effort; webview_go exposes limited window APIs).
	w.Bind("hostMinimize", func() {
		log.Printf("[host] minimize requested (platform support varies)")
		// TODO: platform-specific minimize via w.Window() handle when needed.
	})
	w.Bind("hostMaximize", func() {
		log.Printf("[host] maximize/restore requested (platform support varies)")
	})
	w.Bind("hostClose", func() {
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Destroy()
			})
		})
	})

	// Native navigation — this is the content surface.
	// Note: navigating leaves the chrome HTML; returning to chrome is hostShowChrome.
	// Multi-webview / panel composition is documented in docs/GUI.md.
	w.Bind("hostNavigate", func(url string) {
		log.Printf("[host] native navigate → %s", url)
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Navigate(url)
			})
		})
	})

	w.Bind("hostGoBack", func() {
		log.Printf("[host] goBack (Eval history.back fallback)")
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Eval(`history.back()`)
			})
		})
	})
	w.Bind("hostGoForward", func() {
		log.Printf("[host] goForward")
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Eval(`history.forward()`)
			})
		})
	})
	w.Bind("hostReload", func() {
		log.Printf("[host] reload")
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Eval(`location.reload()`)
			})
		})
	})

	// Return to chrome shell after a remote navigate.
	w.Bind("hostShowChrome", func() {
		cfgPath, err := pathutil.FindFile(mustCwd(), "config.yaml")
		if err != nil {
			log.Printf("[host] showChrome: config not found: %v", err)
			return
		}
		cfg, err := loadConfig(cfgPath)
		if err != nil {
			log.Printf("[host] showChrome: %v", err)
			return
		}
		url := fmt.Sprintf("http://%s/", cfg.IPC.FrontendListen)
		host.with(func(wv webview.WebView) {
			wv.Dispatch(func() {
				wv.Navigate(url)
			})
		})
	})
}

func startStaticServer(addr, webDir string) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	log.Printf("[frontend] static server on http://%s/", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("static server failed: %v", err)
	}
}

func resolveWebDir(cwd string) string {
	candidates := []string{
		filepath.Join(cwd, "web"),
		filepath.Join(cwd, "frontend", "web"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	// Fallback: relative to this source layout
	return filepath.Join(cwd, "web")
}

func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
