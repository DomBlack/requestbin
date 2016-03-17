package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/{binId}/{rest:.*}", Index)
	log.Fatal(http.ListenAndServe(":8080", router))
}

type Request struct {
	Url        string      `json:"url"`
	Method     string      `json:"method"`
	Time       time.Time   `json:"time"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	Host       string      `json:"host"`
	RemoteAddr string      `json:"remote_addr"`
	PostForm   url.Values  `json:"post_form"`
	Form       url.Values  `json:"form"`
	JSON       interface{} `json:"json"`
}

type Requests []Request

func Index(w http.ResponseWriter, r *http.Request) {
	// Read the content
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
	}
	// Restore the io.ReadCloser to its original state
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// Use the content
	body := string(bodyBytes)

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	var json_content interface{}
	if err := json.NewDecoder(r.Body).Decode(&json_content); err != nil {
		// ignore json parsing errors
	}

	response := Request{
		Url:        r.URL.Path,
		Method:     r.Method,
		Host:       r.Host,
		Time:       time.Now(),
		Headers:    r.Header,
		Body:       string(body),
		RemoteAddr: r.RemoteAddr,
		PostForm:   r.PostForm,
		Form:       r.Form,
		JSON:       json_content,
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}
