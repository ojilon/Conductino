package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"frontend/bridge"
	"frontend/handlers"
	"frontend/pathutil"

	webview "github.com/webview/webview_go"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cfgPath, err := pathutil.FindFile(cwd, "config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	webDir := resolveWebDir(cwd)
	log.Printf("[frontend] serving chrome from %s", webDir)

	// C++ core (no-op until cgo links libconductino_core)
	dataDir := filepath.Join(cwd, "conductino-data")
	if err := bridge.NativeInit(dataDir); err != nil {
		log.Printf("[frontend] native init: %v", err)
	}
	defer bridge.NativeShutdown()

	api := handlers.NewAPI()
	go startServer(cfg.IPC.FrontendListen, webDir, api)

	chromeURL := fmt.Sprintf("http://%s/", cfg.IPC.FrontendListen)
	host := bridge.NewHost(chromeURL)

	w := webview.New(cfg.Window.Debug)
	defer w.Destroy()
	host.SetWebView(w)

	w.SetTitle(cfg.Window.Title)
	w.SetSize(cfg.Window.Width, cfg.Window.Height, webview.HintNone)

	host.Bind(w)

	log.Printf("[frontend] chrome at %s (see docs/SHELL.md — dual surface next)", chromeURL)
	w.Navigate(chromeURL)
	w.Run()
}

func startServer(addr, webDir string, api *handlers.API) {
	mux := http.NewServeMux()
	api.Register(mux)
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	log.Printf("[frontend] listening on http://%s/ (API + static)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
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
	return filepath.Join(cwd, "web")
}
