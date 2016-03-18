package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

var Logger = log.New(os.Stdout, " ", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	fmt.Println("starting")
	redis_client, err := redis.Dial("tcp", "redis:6379")
	if err != nil {
		panic(err)
	}
	defer redis_client.Close()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/bins", ApiBinIndexHandler(redis_client))
	router.HandleFunc("/api/bins/{binId}", ApiBinHandler(redis_client))
	router.HandleFunc("/{binId}", BinHandler(redis_client))
	router.HandleFunc("/_/{binId}", LogHandler(redis_client))
	router.HandleFunc("/", HomeHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/app/static/")))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("/app/templates/base.html", "/app/templates/"+tmpl+".html")
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
	redis_client.Do("SADD", "bins", binId)
	redis_client.Do("LPUSH", binKey, string(serialised))
	redis_client.Do("EXPIRE", binKey, string(3600*24))
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
		}{Requests: requests, Title: "Bin #" + binId}

		getTemplate(w, "bin").ExecuteTemplate(w, "base", params)
	}
}

func LogHandler(redis_client redis.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		request := ParseRequest(r)
		bin := mux.Vars(r)["binId"]

		StoreRequest(redis_client, bin, request)
		json_response(w, request)
	}
}
