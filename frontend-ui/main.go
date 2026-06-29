/*
The frontend entry point for the Browser

Some libraries used
github.com/webview/webview_go   →  Native WebView2 (Edge) window on Win11.
                                      Loads ./web/index.html as the UI shell.
gopkg.in/yaml.v3                →  Parses config.yaml.
net/http (stdlib)               →  Runs the local IPC router on :8080
                                      and serves the ./web/ folder.

ARCHITECTURE
  ┌──────────────────────┐   HTTP    ┌──────────────────────┐   C-FFI    ┌─────────┐
  │  WebView2 (web/*)    │ ────────► │  Go IPC Router       │ ─────────► │ Zig +   │
  │  ── app.js fetch()   │  :8080    │  (this file)         │   :8081    │ SQLite  │
  └──────────────────────┘  JSON     └──────────────────────┘  JSON      └─────────┘

 WHY A LOCAL HTTP API INSTEAD OF webview.Bind()?
   Bind() couples the JS context tightly to the Go runtime. By using a
   loopback REST API we preserve **Local Context Isolation** — the web
   content cannot reach backend logic except through a well-typed JSON wire
   protocol. The same API can later be reused by CLI tools or browser
   extensions without changing the backend.


*/

package main

import (
	"Conductino/handlers"
	"fmt"
	"log"
	"net/http"
	"os"

	webview "github.com/webview/webview_go" //native webview2 wrapper
	"gopkg.in/yaml.v3"                      //yaml config parser

	"Conductino/pathutil"
)

//config(mirrors the config.yaml)

type Config struct {
	Window struct{
		Title string `yaml:"title" `
		Width int    `yaml:"width" `
		Height int   `yaml:"height"`
		Debug bool    `yaml:"debug"`
	}`yaml:"window"`
	IPC struct {
		FrontendListen string `yaml:"frontend_listen"`
		BackendURL string `yaml:"backend_url"`
	}`yaml:"ipc"`
	Storage struct {
		DatabasePath string `yaml:"database_path"`
	}`yaml:"storage"`
	Archive struct {
		OutputDir string `yaml:"output_dir"`
		MaxBytes int `yaml:"max_bytes"`
	}`yaml:"archive"`
}

func loadConfig(path string) (*Config, error){
	raw, err := os.ReadFile(path)
	if err != nil {
		return  nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	// gopkg.in/yaml.v3 - strict unmarshal of the YAML tree into the struct.
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return  nil, fmt.Errorf("parser config: %w", err)
	}
	return  &cfg, nil
}

// IPC Router (Go <-> JS/ Go <-> Zig)
/*
Start IPC server wires the HTTP routes that the WebView2 page (app.js) will
call via fetch(). Each route is a thin adapter that re-marshals the request 
and forwards it to the Zig backend at cfg.IPC>BackendURL.
*/
func startIPCServer(cfg *Config){
	mux := http.NewServeMux()

	//static file server - serves ./web/index.html, style.css, app.js.
	//WebView2 will navigate to http://127.0.0.1:8080/ui/.
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("./web"))))

	//Hand off the JSON API endpoints to the handlers package.
	api := handlers.NewBackendClient(cfg.IPC.BackendURL)
	mux.HandleFunc("/api/save_note", api.SaveNoteHandler)
	mux.HandleFunc("/api/search", api.SearchHandler)
	mux.HandleFunc("/api/archive", api.ArchiveHandler)//uses golang.org/x/net/html
	mux.HandleFunc("/api/pdf", api.PDFHandler) // uses github.com/ledongthuc/pdf
	mux.HandleFunc("/api/proxy", api.ProxyHandler)

	log.Printf("[Go IPC] listening on %s", cfg.IPC.FrontendListen)
	if err := http.ListenAndServe(cfg.IPC.FrontendListen, mux); err != nil {
		log.Fatalf("IPC server crashed: %v", err)
	}
}

// the main ------------------------------------------
func main() {
	/*
	os.Getwd() returns wherever the shell is when you run the program, not where main.go lives. So if you cd somewhere else and run the binary, it anchors from there.
    For development this is fine — you always run from the project root. If you later package the binary for distribution, switch the anchor to:
	
	exe, _ := os.Executable()
    anchor := filepath.Dir(exe)

    And put config.yaml next to the binary. The FindFile function works the same either way — you just change what you pass in.
	*/


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

	//boot the local APU in a goroutine so the webview2 main thread is free.
	go startIPCServer(cfg)

	/*
	Initialize WebView2
	debug=true enables the embedded DevTools.
	*/
	w := webview.New(cfg.Window.Debug)
	defer w.Destroy()

	w.SetTitle(cfg.Window.Title)
	w.SetSize(cfg.Window.Width, cfg.Window.Height, webview.HintNone)

	/*
	SECURITY: navigate to localhost, NOT to file:// URL. Which means
	the page lives in a real HTTP origin and same-origin policy applies
	to every fetch() call inside app.js. Local context isolation somewhat archieved.
	*/
	w.Navigate(fmt.Sprintf("http://%s/ui/", cfg.IPC.FrontendListen))

	/*
	w.Bind() is intentionally NOT used for data - see the architecture
	comment at the top. Only a tiny diagnostic helper is exposed.
	*/
	w.Bind("hostPing", func() string { return " pong from Go host"})

	//Blocks until the window is closed
	w.Run()
}