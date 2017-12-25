package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var configFile = flag.String("config", "", "Configuration file")

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

func parseConfig(configFile string) (config Config, err error) {
	if configFile == "" {
		return config, errors.New("Configuration file path is mandatory")
	}

	filename, _ := filepath.Abs(configFile)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return config, err
	}
	return parseYAML(yamlFile)
}

func parseYAML(yamlFile []byte) (config Config, err error) {
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	if config.Address == "" {
		config.Address = "0.0.0.0"
	}
	if config.Port == 0 {
		config.Port = 65000
	}
	return config, nil
}

func main() {
	flag.Parse()

	config, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/hello", http.HandlerFunc(hello))

	hooksHandled := map[string]int{}
	for k, v := range config.Hooks {
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
	log.Printf("Starting web server at %s:%d\n", config.Address, config.Port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), nil)
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
