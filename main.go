package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const port = ":8080"
const docroot = "./docroot"
const redirect = "redirect"
const proxy = "proxy"

func main() {
	// create server
	server := new(server)

	// register handlers
	handlers := make(map[string]http.Handler)
	// redirect handler
	handlers[redirect] = http.HandlerFunc(redirectHandlerFunc)
	// proxy handler
	handlers[proxy] = http.HandlerFunc(proxyHandlerFunc)
	// file system handler
	fs := new(customFileSystem)
	handlers[docroot] = http.FileServer(fs)
	server.handlers = handlers

	// start server
	log.Printf("server start at port = %s\n", port)
	log.Fatal(http.ListenAndServe(port, server))
}

type server struct {
	handlers map[string]http.Handler
}

func (server *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	log.Printf("request path = %s\n", requestPath)

	// check leading slash
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}

	// get handler
	var handler http.Handler
	if strings.HasPrefix(requestPath, "/redirect/") {
		handler = server.handlers[redirect]
	} else if strings.HasPrefix(requestPath, "/proxy/") {
		handler = server.handlers[proxy]
	} else {
		handler = server.handlers[docroot]
	}

	// execute http handle
	handler.ServeHTTP(w, r)
}

type customFileSystem struct{}

func (fs *customFileSystem) Open(name string) (http.File, error) {

	log.Printf("request file name = %s\n", name)

	// check forbidden file
	if name == "/forbidden.txt" {
		return nil, os.ErrPermission
	}

	// create file path
	filePath := docroot + name
	log.Printf("file path = %s\n", filePath)

	// open file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func redirectHandlerFunc(w http.ResponseWriter, r *http.Request) {
	url := "https://go.dev"
	statusCode := 302
	log.Printf("redirect %d %s\n", statusCode, url)
	http.Redirect(w, r, url, statusCode)
}

func proxyHandlerFunc(w http.ResponseWriter, r *http.Request) {
	const proxyUrl = "http://localhost:8081"
	log.Printf("proxy to %s\n", proxyUrl)
	resp, err := http.Get(proxyUrl)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Printf("http proxy error. detail = %v\n", err)
		return
	}
	w.WriteHeader(resp.StatusCode)
	for k := range resp.Header {
		w.Header().Add(k, resp.Header.Get(k))
	}
	io.Copy(w, resp.Body)
}
