package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type DynamicFileHttpRequestHandler struct{}

func (fh DynamicFileHttpRequestHandler) CanHandle(request *HttpRequest) bool {
	splitPath := strings.Split(request.Url, ".")
	extension := splitPath[len(splitPath)-1]

	dynamicFiles := map[string]bool{
		"odt":     true,
		"svg":     true,
		"m3u":     true,
		"xspf":    true,
		"asx":     true,
		"pls":     true,
		"torrent": true,
		"jspdf":   true,
		"xml":     true,
		"dtd":     true,
	}

	return dynamicFiles[extension]
}

func RenderDocumentTemplate(writer io.Writer, source string, trackerUrl string, targetUrl string) {
	t := template.Must(template.New("template").ParseFiles(source))
	params := struct {
		Url    string
		Target string
	}{Url: trackerUrl, Target: targetUrl}
	if err := t.ExecuteTemplate(writer, "document", params); err != nil {
		fmt.Println(err)
	}
}

func GenerateUrlWithExtension(r *http.Request, extension string) url.URL {
	trackerUrl := *r.URL
	trackerUrl.Host = r.Host
	trackerUrl.Scheme = "http" // TODO find a solution

	splitPath := strings.Split(r.URL.Path, ".")
	basePath := strings.Join(splitPath[:len(splitPath)-1], ".")
	trackerUrl.Path = basePath
	if extension != "" {
		trackerUrl.Path += "." + extension
	}
	return trackerUrl
}

func HandleXML(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool {

	splitPath := strings.Split(request.Url, "/")
	if len(splitPath) < 3 {
		return false
	}

	filename := splitPath[len(splitPath)-1]

	source := os.Getenv("ROOT") + "/documents/xml/" + filename
	if _, err := os.Stat(source); err != nil {
		return false
	}

	// X/bin/<path>/filename
	targetUrl := GenerateUrlWithExtension(r, "txt")
	target := targetUrl.String()

	if len(splitPath) >= 4 {
		//  /bin/<path>/filename where path might contain /
		urlType := strings.Join(splitPath[2:len(splitPath)-1], "/")
		newTarget := GetUrls()[urlType]

		if newTarget != "" {
			target = newTarget
		}
	}

	trackerUrl := GenerateUrlWithExtension(r, "dtd")

	RenderDocumentTemplate(w, source, trackerUrl.String(), target)
	return true
}

func (fh DynamicFileHttpRequestHandler) Handle(w http.ResponseWriter, r *http.Request, request *HttpRequest) bool {
	if !fh.CanHandle(request) {
		return false
	}
	splitPath := strings.Split(request.Url, ".")
	extension := splitPath[len(splitPath)-1]

	basePath := strings.Join(splitPath[:len(splitPath)-1], ".")
	trackerUrl := *r.URL
	trackerUrl.Host = r.Host
	trackerUrl.Scheme = "http" // TODO find a solution

	if extension == "xml" || extension == "dtd" {
		return HandleXML(w, r, request)
	}

	if extension == "odt" {
		trackerUrl.Path = basePath + ".jpg"
		filename := filepath.Base(r.URL.Path)
		w.Header().Set("Content-Type", "application/vnd.oasis.opendocument.text")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		source := os.Getenv("ROOT") + "/documents/odt"

		if err := generateODT(w, source, trackerUrl.String()); err != nil {
			fmt.Println(err)
		}
		return true
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

		return true
	}

	if extension == "svg" {
		trackerUrl.Path = basePath + ".css"
		w.Header().Set("Content-Type", "image/svg+xml")

		source := os.Getenv("ROOT") + "/documents/file.svg"
		RenderDocumentTemplate(w, source, trackerUrl.String(), "")
		return true
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
		RenderDocumentTemplate(w, source, trackerUrl.String(), "")
		return true
	}

	if extension == "torrent" {
		trackerUrl.Path = basePath + ".mp3"
		w.Header().Set("Content-Type", "application/x-bittorrent")

		generateTorrent(w, trackerUrl.String())
		return true
	}
	return false
}
