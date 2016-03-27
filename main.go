package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	fmt.Println("starting")

	redis_client, err := redis.Dial("tcp", os.Getenv("REDIS"))
	if err != nil {
		panic(err)
	}
	defer redis_client.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/bins", ApiBinIndexHandler(redis_client))
	router.HandleFunc("/api/bins/{binId}", ApiBinHandler(redis_client))
	router.HandleFunc("/{binId}", BinHandler(redis_client))
	router.HandleFunc("/_/{binId}", LogHandler(redis_client))
	router.HandleFunc("/_/{binId}/{param:.*}", LogHandler(redis_client))
	router.HandleFunc("/", HomeHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(os.Getenv("ROOT") + "/static/")))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
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

type Request struct {
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
}

func (request *Request) ISO8601Time() string {
	return request.Time.Format(time.RFC3339)
}

func ParseRequest(r *http.Request) Request {
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

	return Request{
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
	}
}

func ListBins(redis_client redis.Conn) []string {
	bins, err := redis.Strings(redis_client.Do("SMEMBERS", "bins"))
	if err != nil {
		panic(err)
	}
	return bins
}

func ListRequestsFromBin(redis_client redis.Conn, binId string) []Request {
	raw_requests, err := redis.Strings(redis_client.Do("LRANGE", "bins:"+binId, 0, 10))
	if err != nil {
		panic(err)
	}
	var requests = make([]Request, len(raw_requests))
	for i, item := range raw_requests {
		if err = json.Unmarshal([]byte(item), &requests[i]); err != nil {
			panic(err)
		}
	}
	return requests
}

func StoreRequest(redis_client redis.Conn, binId string, request Request) {
	serialised, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	binKey := "bins:" + binId
	if _, err := redis_client.Do("SADD", "bins", binId); err != nil {
		fmt.Println(err)
	}
	if _, err := redis_client.Do("LPUSH", binKey, string(serialised)); err != nil {
		fmt.Println(err)
	}
	if _, err := redis_client.Do("EXPIRE", binKey, 3600*24); err != nil {
		fmt.Println(err)
	}
}

func json_response(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		panic(err)
	}
}

func ApiBinHandler(redis_client redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		bin := mux.Vars(r)["binId"]
		requests := ListRequestsFromBin(redis_client, bin)
		json_response(w, requests)
	}
}

func ApiBinIndexHandler(redis_client redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		bins := ListBins(redis_client)
		json_response(w, bins)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Title string
	}{Title: "Welcome to RequestBin"}

	getTemplate(w, "home").ExecuteTemplate(w, "base", params)
}

func BinHandler(redis_client redis.Conn) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		binId := mux.Vars(r)["binId"]
		requests := ListRequestsFromBin(redis_client, binId)

		params := struct {
			Requests []Request
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

func LogHandler(redis_client redis.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := ParseRequest(r)
		bin := mux.Vars(r)["binId"]
		StoreRequest(redis_client, bin, request)

		splitPath := strings.Split(request.Url, ".")
		extension := splitPath[len(splitPath)-1]

		staticFiles := map[string]bool{
			"jpg": true,
			"png": true,
			"bmp": true,
			"gif": true,
			"css": true,
			"mp3": true,
		}

		if staticFiles[extension] {
			http.Redirect(w, r, "/files/file."+extension, 302)
			return
		}

		dynamicFiles := map[string]bool{
			"odt":     true,
			"svg":     true,
			"m3u":     true,
			"torrent": true,
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

			if extension == "svg" {
				trackerUrl.Path = basePath + ".css"
				w.Header().Set("Content-Type", "image/svg+xml")

				source := os.Getenv("ROOT") + "/documents/file.svg"
				RenderDocumentTemplate(w, source, trackerUrl.String())
				return
			}

			if extension == "m3u" {
				trackerUrl.Path = basePath + ".mp3"
				w.Header().Set("Content-Type", "audio/mpegurl")

				source := os.Getenv("ROOT") + "/documents/file.m3u"
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
