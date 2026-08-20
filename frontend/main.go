package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:web
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Conductino Study Browser",
		Width:            1280,
		Height:           900,
		MinWidth:         720,
		MinHeight:        480,
		Frameless:        false, // OS title bar: minimize / maximize / close
		DisableResize:    false, // user can resize the app window
		Fullscreen:       false,
		StartHidden:      false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 20, B: 25, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:   false,
			Theme:                windows.SystemDefault,
		},
	})
	if err != nil {
		log.Fatal("Wails exit: ", err)
	}
}
