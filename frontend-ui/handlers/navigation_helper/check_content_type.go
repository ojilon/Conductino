package navigation_helper

import "strings"

func IsHTML(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/html")
}