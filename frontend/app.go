package main

import (
	"context"
	"fmt"
)

// App is the Wails-bound application API.
// Methods here are callable from JS as window.go.main.App.<Method>().
// Step 1: health + window only. Tabs/navigate/file come in Step 3.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet is a smoke-test binding so you can verify JS ↔ Go before wiring real features.
func (a *App) Greet(name string) string {
	if name == "" {
		name = "researcher"
	}
	return fmt.Sprintf("Conductino is alive, %s. Wails shell ready.", name)
}

// AppInfo returns basic runtime metadata for the chrome UI.
func (a *App) AppInfo() map[string]string {
	return map[string]string{
		"name":    "Conductino Study Browser",
		"version": "0.2.0-wails",
		"engine":  "wails-v2",
	}
}
