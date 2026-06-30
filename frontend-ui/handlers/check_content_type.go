package handlers

import "strings"

func IsHTML(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html")
}