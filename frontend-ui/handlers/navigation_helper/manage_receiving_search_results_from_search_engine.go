package navigation_helper

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type SearchResult struct {
	Title   string
	Snippet string
	URL     string // pre-rewritten to /api/proxy?url=...
}

// ExtractSearchResults walks a DuckDuckGo HTML results page and pulls out
// each result's title, link, and snippet.
//
// DuckDuckGo's html.duckduckgo.com markup (as of writing) looks roughly like:
//
//	<div class="result results_links results_links_deep web-result">
//	    <a class="result__a" href="...">Title text</a>
//	    <a class="result__snippet" href="...">Snippet text</a>
//	</div>
//
// This WILL break if DuckDuckGo changes their class names — that's normal
// for scraping and not a bug in your code; you'd just update the class
// strings below.
func ExtractSearchResults(body io.Reader, baseURL *url.URL) ([]SearchResult, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	var current *SearchResult // the result we're currently filling in

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			class := getAttr(node, "class")

			switch {
			case node.Data == "div" && strings.Contains(class, "result__body"):
				// start of a new result block
				results = append(results, SearchResult{})
				current = &results[len(results)-1]

			case node.Data == "a" && strings.Contains(class, "result__a"):
				if current != nil {
					href := getAttr(node, "href")
					current.URL = resolveAndRewrite(href, baseURL)
					current.Title = textContent(node)
				}

			case node.Data == "a" && strings.Contains(class, "result__snippet"):
				if current != nil {
					current.Snippet = textContent(node)
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	// drop any partially-built results that never got a title/URL
	// (DuckDuckGo sometimes has ad blocks or empty containers)
	clean := results[:0]
	for _, r := range results {
		if r.Title != "" && r.URL != "" {
			clean = append(clean, r)
		}
	}

	return clean, nil
}

// getAttr returns the value of a single attribute, or "" if not present.
// Same job as the loop inside rewriteAttribute in rewrite.go, just read-only.
func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// textContent flattens all text inside a node's subtree into one string.
// e.g. <a>Hello <b>World</b></a> -> "Hello World"
func textContent(node *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(node)
	return strings.TrimSpace(sb.String())
}

// resolveAndRewrite turns a raw href from the results page into an
// absolute URL, then routes it through your proxy — same idea as
// rewriteAttribute in rewrite.go, just returning a string instead of
// mutating a node in place.
func resolveAndRewrite(href string, baseURL *url.URL) string {
	if href == "" {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	absolute := baseURL.ResolveReference(ref)
	return "/api/proxy?url=" + url.QueryEscape(absolute.String())
}