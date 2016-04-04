package main

import (
	"net/http"
)

type HttpRequestHandler interface {
	CanHandle(request *HttpRequest) bool
	Handle(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool
}
