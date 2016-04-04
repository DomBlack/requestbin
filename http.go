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
	"text/template"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

type UserAgentInfo struct {
	AgentType         string `json:"agent_type"`
	AgentName         string `json:"agent_name"`
	AgentVersion      string `json:"agent_version"`
	OSType            string `json:"os_type"`
	OSName            string `json:"os_name"`
	OSVersionName     string `json:"os_versionName"`
	OSVersionNumber   string `json:"os_versionNumber"`
	LinuxDistribution string `json:"linux_distribution"`
}

type HttpRequest struct {
	Request
	Url           string        `json:"url"`
	FullUrl       string        `json:"full_url"`
	Method        string        `json:"method"`
	Headers       http.Header   `json:"headers"`
	Body          string        `json:"body"`
	Host          string        `json:"host"`
	PostForm      url.Values    `json:"post_form"`
	Form          url.Values    `json:"form"`
	JSON          interface{}   `json:"json"`
	BinId         string        `json:"bin_id"`
	UserAgentInfo UserAgentInfo `json:"user_agent_info"`
}

func (request *HttpRequest) ISO8601Time() string {
	return request.Time.Format(time.RFC3339)
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func GetUserAgentInfo(r *http.Request) (UserAgentInfo, error) {
	var userAgentInfo UserAgentInfo
	userAgent := r.Header["User-Agent"]
	if userAgent != nil {
		userAgent, err := UrlEncoded(userAgent[0])
		if err != nil {
			return userAgentInfo, err
		}

		lookupUrl := fmt.Sprintf("http://www.useragentstring.com/?uas=%s&getJSON=all", userAgent)
		res, err := http.Get(lookupUrl)
		if err != nil {
			return userAgentInfo, err
		}

		defer res.Body.Close()

		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&userAgentInfo)
		if err != nil {
			return userAgentInfo, err
		}
	}
	return userAgentInfo, nil
}

func startLoggingHttpServer(port int, staticRoot string, writers []HttpRequestWriter) {
	fmt.Printf("Starting HTTP logging server on port %d\n", port)

	handlers := []HttpRequestHandler{
		StaticFileHttpRequestHandler{},
		DynamicFileHttpRequestHandler{},
	}

	router := mux.NewRouter().StrictSlash(true)
	router.PathPrefix("/files").Handler(http.FileServer(http.Dir(staticRoot)))
	router.HandleFunc("/", LogHandler(handlers, writers))
	router.HandleFunc("/{binId}", LogHandler(handlers, writers))
	router.HandleFunc("/{binId}/{param:.*}", LogHandler(handlers, writers))
	go http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func startAdminHttpServer(port int, staticRoot string, redisClient redis.Conn) {
	fmt.Printf("Starting HTTP admin server on port %d\n", port)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/bins", ApiBinIndexHandler(redisClient))
	router.HandleFunc("/api/bins/{binId}", ApiBinHandler(redisClient))
	router.HandleFunc("/{binId}", BinHandler(redisClient))
	router.HandleFunc("/", HomeHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(staticRoot)))
	go http.ListenAndServe(fmt.Sprintf(":%d", port), router)
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

	userAgentInfo, _ := GetUserAgentInfo(r)

	return HttpRequest{
		Request{
			Time:       time.Now(),
			RemoteAddr: r.RemoteAddr,
		},
		r.URL.Path,
		fullUrl.String(),
		r.Method,
		r.Header,
		string(body),
		r.Host,
		r.PostForm,
		r.Form,
		json_content,
		binId,
		userAgentInfo,
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

func LogHandler(handlers []HttpRequestHandler, writers []HttpRequestWriter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		binId, ok := mux.Vars(r)["binId"]
		if !ok {
			binId = "_root"
		}
		request := ParseHttpRequest(r, binId)
		for _, writer := range writers {
			err := writer.WriteHttpRequest(request)
			if err != nil {
				fmt.Println(err)
			}
		}

		for _, handler := range handlers {
			if handler.Handle(w, r, &request) {
				return
			}
		}

		json_response(w, request)

	}
}
