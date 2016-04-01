package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type UsersMap map[string]string

func parsePasswdFile(filename string) (UsersMap, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res := UsersMap{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed line, no colon: %s", line)
		}
		res[parts[0]] = parts[1]
	}

	return res, nil
}

type AuthenticatedProxyHandler struct {
	Proxy *httputil.ReverseProxy
	Users *UsersMap
}

func Write401(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"requestbin\"")
	http.Error(w, "authorization failed", http.StatusUnauthorized)
}

func (p AuthenticatedProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		Write401(w)
		return
	}
	hashedPass := fmt.Sprintf("%x", md5.Sum([]byte(pass)))
	if (*p.Users)[user] == hashedPass {
		fmt.Println(r.Host)
		p.Proxy.ServeHTTP(w, r)
	} else {
		Write401(w)
		return
	}

}

func startKibanaProxy(port int, kibanaHost string, authFilename string) {
	fmt.Printf("Starting Kibana proxy on port %d\n", port)

	reverseProxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   kibanaHost,
	})
	users, err := parsePasswdFile(authFilename)
	if err != nil {
		panic(err)
	}
	proxy := AuthenticatedProxyHandler{Proxy: reverseProxy, Users: &users}
	go http.ListenAndServe(fmt.Sprintf(":%d", port), proxy)
}
