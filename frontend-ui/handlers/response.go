package handlers

import (
	"log"
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
		log.Println("To display html: ", contentType)
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
		log.Println("\n\n-----------------------------------")
		log.Println("Downloading", contentType)
		log.Println("---------------------------------------s")

		return ActionDownload
	}
}