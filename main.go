package main

import (
	"encoding/json"
	"flag"
	"fmt"
	// "html"
	"log"
	"net/http"
)

var addr = flag.String("addr", "0.0.0.0", "Bind address")
var path = flag.String("path", "payload", "Payload path")
var port = flag.Int("port", 65000, "Listening port")

type Response struct {
	Status int
	Msg    string
}

func main() {
	flag.Parse()
	http.Handle("/hello", http.HandlerFunc(hello))
	http.Handle(fmt.Sprintf("/%s", *path), http.HandlerFunc(payload))
	log.Printf("Starting web server at %s:%d\n", *addr, *port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addr, *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %q", req.URL.Path)
	response := Response{Status: 200, Msg: "Hello from git-webhook listener"}
	json.NewEncoder(w).Encode(response)
}

func payload(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %q", req.URL.Path)
}
