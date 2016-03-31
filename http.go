package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

type HttpRequest struct {
	Url        string      `json:"url"`
	FullUrl    string      `json:"full_url"`
	Method     string      `json:"method"`
	Time       time.Time   `json:"time"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	Host       string      `json:"host"`
	RemoteAddr string      `json:"remote_addr"`
	PostForm   url.Values  `json:"post_form"`
	Form       url.Values  `json:"form"`
	JSON       interface{} `json:"json"`
	BinId      string      `json:"bin_id"`
}

func (request *HttpRequest) ISO8601Time() string {
	return request.Time.Format(time.RFC3339)
}

func startHTTPServer(root string, port string, redisClient redis.Conn, writers ...HttpRequestWriter) {
	fmt.Println("Starting HTTP server on port " + port)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/bins", ApiBinIndexHandler(redisClient))
	router.HandleFunc("/api/bins/{binId}", ApiBinHandler(redisClient))
	router.HandleFunc("/{binId}", BinHandler(redisClient))
	router.HandleFunc("/_/{binId}", LogHandler(writers...))
	router.HandleFunc("/_/{binId}/{param:.*}", LogHandler(writers...))
	router.HandleFunc("/", HomeHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(root + "/static/")))
	go http.ListenAndServe(":"+port, router)
}

func ParseHttpRequest(r *http.Request, binId string) HttpRequest {
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

	fullUrl := r.URL
	fullUrl.Host = r.Host
	fullUrl.Scheme = "http" // TODO find a solution
	fullUrl.RawQuery = r.URL.Query().Encode()

	return HttpRequest{
		Url:        r.URL.Path,
		FullUrl:    fullUrl.String(),
		Method:     r.Method,
		Host:       r.Host,
		Time:       time.Now(),
		Headers:    r.Header,
		Body:       string(body),
		RemoteAddr: r.RemoteAddr,
		PostForm:   r.PostForm,
		Form:       r.Form,
		JSON:       json_content,
		BinId:      binId,
	}
}

func json_response(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		panic(err)
	}
}

func ApiBinHandler(redisClient redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		bin := mux.Vars(r)["binId"]
		requests := ListRequestsFromBin(redisClient, bin)
		json_response(w, requests)
	}
}

func ApiBinIndexHandler(redisClient redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		bins := ListBins(redisClient)
		json_response(w, bins)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Title string
	}{Title: "Welcome to RequestBin"}

	getTemplate(w, "home").ExecuteTemplate(w, "base", params)
}

func BinHandler(redisClient redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		binId := mux.Vars(r)["binId"]
		requests := ListRequestsFromBin(redisClient, binId)

		params := struct {
			Requests []HttpRequest
			Title    string
		}{
			Requests: requests,
			Title:    "Bin #" + binId,
		}

		err := getTemplate(w, "bin").ExecuteTemplate(w, "base", params)
		if err != nil {
			panic(err)
		}
	}
}

func RenderDocumentTemplate(writer io.Writer, source string, trackerUrl string) {
	t := template.Must(template.New("template").ParseFiles(source))
	params := struct{ Url string }{Url: trackerUrl}
	if err := t.ExecuteTemplate(writer, "document", params); err != nil {
		fmt.Println(err)
	}
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	funcMap := template.FuncMap{
		"lookup_addr": func(addr string) []string {
			names, err := net.LookupAddr(addr)
			if err != nil {
				return []string{addr}
			}
			return names
		},
	}
	templates := template.New("template").Funcs(funcMap)
	templates_folder := os.Getenv("ROOT") + "/templates/"
	_, err := templates.ParseFiles(templates_folder+"base.html", templates_folder+tmpl+".html")
	if err != nil {
		fmt.Println(err)
	}

	return template.Must(templates, err)
}

func LogHandler(writers ...HttpRequestWriter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		binId := mux.Vars(r)["binId"]
		request := ParseHttpRequest(r, binId)
		for _, writer := range writers {
			err := writer.WriteHttpRequest(request)
			if err != nil {
				fmt.Println(err)
			}
		}

		splitPath := strings.Split(request.Url, ".")
		extension := splitPath[len(splitPath)-1]

		staticFiles := map[string]bool{
			"jpg": true,
			"png": true,
			"bmp": true,
			"gif": true,
			"css": true,
			"mp3": true,
			"xml": true,
		}

		if staticFiles[extension] {
			http.Redirect(w, r, "/files/file."+extension, 302)
			return
		}

		dynamicFiles := map[string]bool{
			"odt":     true,
			"svg":     true,
			"m3u":     true,
			"xspf":    true,
			"asx":     true,
			"pls":     true,
			"torrent": true,
			"jspdf":   true,
		}

		if dynamicFiles[extension] {

			basePath := strings.Join(splitPath[:len(splitPath)-1], ".")
			trackerUrl := *r.URL
			trackerUrl.Host = r.Host
			trackerUrl.Scheme = "http" // TODO find a solution

			if extension == "odt" {
				trackerUrl.Path = basePath + ".jpg"
				filename := filepath.Base(r.URL.Path)
				w.Header().Set("Content-Type", "application/vnd.oasis.opendocument.text")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

				source := os.Getenv("ROOT") + "/documents/odt"

				if err := generateODT(w, source, trackerUrl.String()); err != nil {
					fmt.Println(err)
				}
				return
			}

			if extension == "jspdf" {
				trackerUrl.Path = basePath + ".jpg"

				// length of the JS code in pdf
				length := 67 + len(trackerUrl.String())

				filename := filepath.Base(basePath + ".pdf")
				w.Header().Set("Content-Type", "application/pdf")
				w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

				source := os.Getenv("ROOT") + "/documents/file.jspdf"
				t := template.Must(template.New("template").ParseFiles(source))
				params := struct {
					Url    string
					Length int
				}{Url: trackerUrl.String(), Length: length}

				if err := t.ExecuteTemplate(w, "document", params); err != nil {
					fmt.Println(err)
				}

				return
			}

			if extension == "svg" {
				trackerUrl.Path = basePath + ".css"
				w.Header().Set("Content-Type", "image/svg+xml")

				source := os.Getenv("ROOT") + "/documents/file.svg"
				RenderDocumentTemplate(w, source, trackerUrl.String())
				return
			}

			playListMimeType := map[string]string{
				"pls":  "audio/x-scpls",
				"m3u":  "audio/mpegurl",
				"asx":  "video/x-ms-asf",
				"xspf": "application/xspf+xml",
			}[extension]
			if playListMimeType != "" {
				trackerUrl.Path = basePath + ".mp3"
				w.Header().Set("Content-Type", playListMimeType)

				source := os.Getenv("ROOT") + "/documents/file." + extension
				RenderDocumentTemplate(w, source, trackerUrl.String())
				return
			}

			if extension == "torrent" {
				trackerUrl.Path = basePath + ".mp3"
				w.Header().Set("Content-Type", "application/x-bittorrent")

				generateTorrent(w, trackerUrl.String())
				return
			}
		}

		json_response(w, request)

	}
}
