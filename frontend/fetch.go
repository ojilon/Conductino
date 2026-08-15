package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	reScriptStyle = regexp.MustCompile(`(?is)<(script|style|noscript|svg)[^>]*>.*?</(script|style|noscript|svg)>`)
	reTags        = regexp.MustCompile(`(?s)<[^>]+>`)
	reSpace       = regexp.MustCompile(`[ \t\x0b\f\r]+`)
	reBlankLines  = regexp.MustCompile(`\n{3,}`)
)

// FetchPageText downloads a URL and returns a rough plain-text extract.
// Used when sites block iframe embedding or cross-origin selection.
func (a *App) FetchPageText(url string) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("empty url")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("only http(s) supported")
	}

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ConductinoStudyBrowser/0.4 (local research tool)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.8")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d", res.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, 2<<20)) // 2 MiB cap
	if err != nil {
		return "", err
	}
	ct := res.Header.Get("Content-Type")
	if strings.Contains(ct, "text/plain") {
		return strings.TrimSpace(string(body)), nil
	}
	return htmlToText(string(body)), nil
}

func htmlToText(html string) string {
	s := reScriptStyle.ReplaceAllString(html, " ")
	// Block breaks
	s = regexp.MustCompile(`(?i)</(p|div|h[1-6]|li|tr|br|section|article)>`).ReplaceAllString(s, "\n")
	s = reTags.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = reSpace.ReplaceAllString(s, " ")
	s = reBlankLines.ReplaceAllString(s, "\n\n")
	// Drop control chars except newline
	var b strings.Builder
	for _, r := range s {
		if r == '\n' || unicode.IsPrint(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
