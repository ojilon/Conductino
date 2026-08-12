package main

import (
	"fmt"
	"frontend/pathutil"
	"log"
	"os"

	webview "github.com/webview/webview_go" //native webview2 wrapper
)

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
		go StartIPCServer(cfg)

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
		w.Navigate(fmt.Sprintf("http://%s/", cfg.IPC.FrontendListen))

		/*
		w.Bind() is intentionally NOT used for data
		*/
		w.Bind("hostPing", func() string { return " pong from Go host"})

		//Blocks until the window is closed
		w.Run()
}