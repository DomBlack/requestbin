package main

import (
	"net/http"
	"strings"
)

type StaticFileHttpRequestHandler struct {
}

func (fh StaticFileHttpRequestHandler) CanHandle(request *HttpRequest) bool {
	staticFiles := map[string]bool{
		"avi":  true,
		"bmp":  true,
		"css":  true,
		"flv":  true,
		"gif":  true,
		"jpg":  true,
		"mp3":  true,
		"mp4":  true,
		"png":  true,
		"txt":  true,
		"webm": true,
		"wmv":  true,
		"xml":  true,
	}

	splitPath := strings.Split(request.Url, ".")
	extension := splitPath[len(splitPath)-1]
	return staticFiles[extension]
}

func (fh StaticFileHttpRequestHandler) Handle(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool {
	if !fh.CanHandle(request) {
		return false
	}
	splitPath := strings.Split(request.Url, ".")
	extension := splitPath[len(splitPath)-1]

	http.Redirect(w, r, "/files/file."+extension, 302)
	return true
}
