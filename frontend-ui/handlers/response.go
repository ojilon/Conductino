package handlers

import (
	"strings"
)

type ResponseAction int

const (
	ActionDisplayHTML ResponseAction = iota
	ActionStream
	ActionDownload 
)

func DecideResponse(contentType string) ResponseAction {
	contentType = strings.ToLower(contentType)

	switch {
	case strings.Contains(contentType, "text/html"):
		return ActionDisplayHTML
	case strings.Contains(contentType, "application/xhtml+xml"):
		return  ActionDisplayHTML
	case strings.HasPrefix(contentType, "image/"):
		return ActionStream
	case strings.HasPrefix(contentType, "video/"):
		return ActionStream
	case strings.HasPrefix(contentType, "audio/"):
		return ActionStream
	case strings.Contains(contentType, "application/pdf"):
		return ActionStream
	default:
		return ActionDownload
	}
}