package handlers

import (
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	nhelper "Conductino/handlers/navigation_helper"
	"html/template"
)

// results.html lives at ../web/templates/results.html relative to this file's
// package (handlers/), i.e. frontend-ui/web/templates/results.html
var resultsTmpl = template.Must(template.ParseFiles(
	filepath.Join("web", "states_html", "plain_text.html"),
))


type ResultsPage struct {
	Query   string
	Results []nhelper.SearchResult
}

func (c *BackendClient) PlainTextOrchestrator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET required", http.StatusMethodNotAllowed)
		log.Println("(PlainText api) Needs method, what's got:", r.Method)
		return
	}

	// read the ?url=
	plaintexturl := r.URL.Query().Get("url")
	if plaintexturl == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	// validate the url
	if _, err := url.ParseRequestURI(plaintexturl); err != nil {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		log.Println("invalid url:", err)
		return
	}

	// 1. fetch the search engine's results page — same pattern as ProxyHandler
	req, err := http.NewRequest(http.MethodGet, plaintexturl, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		log.Println("Badgateway: ", err)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.Browser.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		log.Println("Failed to get response, error -> ", err)
		return
	}
	defer resp.Body.Close()

	if !nhelper.IsHTML(resp.Header.Get("Content-Type")) {
		http.Error(w, "unexpected content type from search engine", http.StatusBadGateway)
		return
	}

	// 2. parse out result blocks — this is the one piece that's genuinely
	// engine-specific (DuckDuckGo/Google/Bing all use different markup),
	// so it lives in its own helper you write per-engine.
	results, err := nhelper.ExtractSearchResults(resp.Body, resp.Request.URL)
	if err != nil {
		log.Println("extract error:", err)
		http.Error(w, "could not parse results", http.StatusInternalServerError)
		return
	}

	// 3. build the page data and render YOUR template, not their HTML
	page := ResultsPage{
		Query:   r.URL.Query().Get("q"), // pass this through from navigate.go's decision if you want it displayed
		Results: results,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := resultsTmpl.Execute(w, page); err != nil {
		log.Println("template execute error:", err)
	}
}