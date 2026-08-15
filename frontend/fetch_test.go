package main

import "testing"

func TestHtmlToTextSmoke(t *testing.T) {
	html := `<html><head><script>evil()</script><style>.x{}</style></head>` +
		`<body><h1>Leaf</h1><p>A leaf is a plant organ.</p></body></html>`
	out := htmlToText(html)
	if out == "" {
		t.Fatal("expected non-empty text")
	}
	if containsFold(out, "evil") {
		t.Fatalf("script body should be stripped, got: %q", out)
	}
	if !containsFold(out, "leaf") {
		t.Fatalf("expected article text, got: %q", out)
	}
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		stringIndexFold(s, sub) >= 0)
}

func stringIndexFold(s, sub string) int {
	// tiny case-insensitive search without importing strings to keep file obvious
	ls, lsub := []rune(s), []rune(sub)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		ok := true
		for j := 0; j < len(lsub); j++ {
			a, b := ls[i+j], lsub[j]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}
