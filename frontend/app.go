package main

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails-bound application API.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	if name == "" {
		name = "researcher"
	}
	return fmt.Sprintf("Conductino is alive, %s. Wails shell ready.", name)
}

func (a *App) AppInfo() map[string]string {
	return map[string]string{
		"name":    "Conductino Study Browser",
		"version": "0.2.0-wails",
		"engine":  "wails-v2",
	}
}

// WindowMinimise — OS minimize (also available via title-bar).
func (a *App) WindowMinimise() {
	if a.ctx == nil {
		return
	}
	runtime.WindowMinimise(a.ctx)
}

// WindowToggleMaximise toggles maximize / restore.
func (a *App) WindowToggleMaximise() {
	if a.ctx == nil {
		return
	}
	runtime.WindowToggleMaximise(a.ctx)
}

// WindowClose quits the application.
func (a *App) WindowClose() {
	if a.ctx == nil {
		return
	}
	runtime.Quit(a.ctx)
}
