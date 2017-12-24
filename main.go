package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	// "html"
	"log"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

var config = flag.String("config", "", "Configuration file")

type Response struct {
	Status int
	Msg    string
}

type Config struct {
	Address string
	Port    int
	Hooks   map[string]Hook
}

type Hook struct {
	Type string
	Path string
	Cmd  string
}

func main() {
	flag.Parse()
	if *config == "" {
		log.Fatal("Configuration file path is mandatory")
		panic("Configuration file path is mandatory")
	}
	filename, _ := filepath.Abs(*config)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatal("Unable to open configuration file: %s", *config)
		panic(err)
	}

	var configuration Config
	err = yaml.Unmarshal(yamlFile, &configuration)
	if err != nil {
		log.Fatal("Unable to parse configuration file: %s", *config)
		panic(err)
	}
	log.Printf("Read configuration: %#v", configuration)
	addr := configuration.Address
	if addr == "" {
		addr = "0.0.0.0"
	}
	port := configuration.Port
	if port == 0 {
		port = 65000
	}
	http.Handle("/hello", http.HandlerFunc(hello))

	hooksHandled := map[string]int{}
	for k, v := range configuration.Hooks {
		log.Printf("Read hook %s: %#v", k, v)
		if _, exists := hooksHandled[v.Path]; exists {
			log.Printf("[WARN] Path %s already defined, ignoring...", v.Path)
			continue
		}
		if v.Type != "bitbucket" && v.Type != "github" && v.Type != "gitlab" {
			log.Printf("[WARN] Unknown repository type, it must be one of: bitbucket, github or gitlab")
			continue
		}
		if !strings.HasPrefix(v.Path, "/") || v.Path == "/hello" {
			log.Printf("[WARN] Path must start with / and be different of /hello")
			continue
		}
		if v.Cmd == "" {
			log.Printf("[WARN] Cmd must be defined")
			continue
		}

		hooksHandled[v.Path] = 1
	}

	log.Printf("Added %d hooks", len(hooksHandled))
	log.Printf("Starting web server at %s:%d\n", addr, port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %q", req.URL.Path)
	response := Response{Status: 200, Msg: "Hello from githook listener"}
	json.NewEncoder(w).Encode(response)
}

func payload(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %q", req.URL.Path)
}
