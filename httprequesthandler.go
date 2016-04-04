package main

import (
	"net/http"
)

type HttpRequestHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool
}
