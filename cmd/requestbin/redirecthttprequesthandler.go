package main

import (
	"fmt"
	"net/http"
	"strings"
)

type RedirectHttpRequestHandler struct{}

func (fh RedirectHttpRequestHandler) Handle(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool {
	splitPath := strings.SplitN(request.Url, "/", 4)

	// /bin/redirect/type
	if len(splitPath) < 4 || splitPath[2] != "redirect" {
		return false
	}

	urlType := splitPath[3]

	if urlType == "self" {
		trackerUrl := *r.URL
		trackerUrl.Host = r.Host
		trackerUrl.Scheme = "http" // TODO find a solution
		trackerUrl.Path = fmt.Sprintf("/%s/postredirect", splitPath[1])
		http.Redirect(w, r, trackerUrl.String(), 302)
		return true
	}

	urls := GetUrls()
	target := urls[urlType]
	if target != "" {
		http.Redirect(w, r, target, 302)
		return true
	}

	return false
}
