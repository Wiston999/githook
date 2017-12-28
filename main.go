package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/wiston999/githook/event"

	"github.com/nu7hatch/gouuid"
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
	filename, err := filepath.Abs(configFile)
	if err != nil {
		return config, err
	}

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

func logRequest(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId, _ := uuid.NewV4()
		log.Printf("[INFO] Received request (%s) '%s' %s", requestId, r.URL.Path, r.Method)
		defer log.Printf("[INFO] Request (%s) completed", requestId)

		h.ServeHTTP(w, r)
	})
}

func repoRequest(h http.HandlerFunc, hookName string, hookInfo Hook) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var repoEvent *event.RepoEvent
		var response Response
		var err error

		switch hookInfo.Type {
		case "bitbucket":
			localEvent, localErr := event.NewBitbucketEvent(r)
			if localErr != nil {
				repoEvent, err = nil, localErr
			} else {
				repoEvent, err = &localEvent.RepoEvent, localErr
			}
		case "github":
			localEvent, localErr := event.NewGithubEvent(r)
			if localErr != nil {
				repoEvent, err = nil, localErr
			} else {
				repoEvent, err = &localEvent.RepoEvent, localErr
			}
		}

		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("Error while parsing event: %s", err)
			json.NewEncoder(w).Encode(response)
			return
		}

		if repoEvent.Branch == "" {
			response.Status, response.Msg = 500, "Repository type is unknown"
			json.NewEncoder(w).Encode(response)
			return
		}

		log.Printf("[DEBUG] Repository event parsed: %#v", repoEvent)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	response := Response{Status: 200, Msg: "Hello from githook listener"}
	json.NewEncoder(w).Encode(response)
}

func payload(w http.ResponseWriter, req *http.Request) {
	log.Printf("Received request: %q", req.URL.Path)
}

func main() {
	flag.Parse()

	config, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	h := http.NewServeMux()

	h.HandleFunc("/hello", logRequest(hello))

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
		h.HandleFunc(v.Path, logRequest(repoRequest(payload, k, v)))
	}

	log.Printf("Added %d hooks", len(hooksHandled))
	log.Printf("Starting web server at %s:%d\n", config.Address, config.Port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), h)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
