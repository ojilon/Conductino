package handlers

import (
	"net/url"
	"strings"
)

// NavigationKind classifies omnibox input.
type NavigationKind int

const (
	NavWebsite NavigationKind = iota
	NavSearch
	NavEmpty
)

func (k NavigationKind) String() string {
	switch k {
	case NavWebsite:
		return "website"
	case NavSearch:
		return "search"
	default:
		return "empty"
	}
}

// SearchEngines maps setting keys to query URL templates (%s = query).
// Kept in sync with frontend/web/js/app.js ENGINES where practical.
var SearchEngines = map[string]string{
	"duckduckgo": "https://duckduckgo.com/?q=%s",
	"google":     "https://www.google.com/search?q=%s",
	"bing":       "https://www.bing.com/search?q=%s",
	"startpage":  "https://www.startpage.com/sp/search?query=%s",
}

// NavigationDecision is a pure classification result (no network I/O).
type NavigationDecision struct {
	Kind  NavigationKind
	Input string
	URL   string
	Query string
}

// DetectNavigation decides whether input is a URL or a search query.
// This replaces the old DetectNavigationHandler HTTP endpoint for the chrome path:
// the JS omnibox already does the same check; Go keeps this for API/CLI/tests.
func DetectNavigation(input, engine string) NavigationDecision {
	input = strings.TrimSpace(input)
	if input == "" {
		return NavigationDecision{Kind: NavEmpty}
	}

	if looksLikeURL(input) {
		candidate := input
		if !strings.Contains(candidate, "://") {
			candidate = "https://" + candidate
		}
		return NavigationDecision{
			Kind:  NavWebsite,
			Input: input,
			URL:   candidate,
		}
	}

	base, ok := SearchEngines[strings.ToLower(engine)]
	if !ok {
		base = SearchEngines["duckduckgo"]
	}
	return NavigationDecision{
		Kind:  NavSearch,
		Input: input,
		Query: input,
		URL:   strings.Replace(base, "%s", url.QueryEscape(input), 1),
	}
}

func looksLikeURL(input string) bool {
	if strings.ContainsAny(input, " \t\n") {
		return false
	}
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") ||
		strings.HasPrefix(input, "about:") || strings.HasPrefix(input, "file:") {
		return true
	}
	candidate := input
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}
	u, err := url.Parse(candidate)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	return strings.Contains(host, ".")
}
