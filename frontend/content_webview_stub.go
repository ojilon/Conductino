//go:build !windows

package main

import "fmt"

func (a *App) ContentEnsure() error { return fmt.Errorf("content WebView2 is Windows-only") }
func (a *App) ContentNavigate(url string) error {
	return a.Navigate(url) // fallback: full-window navigate
}
func (a *App) ContentSetVisible(show bool) error  { return nil }
func (a *App) ContentResize() error               { return nil }
func (a *App) ContentSetChromeHeight(px int) error { return nil }
func (a *App) ContentEval(js string) error        { return nil }
func (a *App) ContentLastURL() string             { return "" }
func (a *App) ContentCopySelection() error        { return nil }
