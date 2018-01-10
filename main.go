package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Wiston999/githook/event"
	"github.com/Wiston999/githook/server"

	"gopkg.in/yaml.v2"
)

var configFile = flag.String("config", "", "Configuration file")

type Config struct {
	// Address where the HTTP server will be bind
	Address string
	// Port where the HTTP server will be listening
	Port  int
	Hooks map[string]event.Hook
}

// parseConfig parses a YAML configuration file given its filename
// It returns a Config structure and error in case of errors
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

// parseYAML parses a YAML configuration file given its string representation
// It returns a Config structure and error in case of errors
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

// addHandlers configures hook handlers into an http.ServeMux handler given a Config structure
// It returns a map containing the hooks added as key
func addHandlers(cmdLog server.CommandLog, config Config, h *http.ServeMux) (hooksHandled map[string]int) {
	hooksHandled = make(map[string]int)
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
		if v.Timeout <= 0 {
			log.Printf("[WARN] Timeout must be greater than 0, got %v", v.Timeout)
			continue
		}
		if len(v.Cmd) == 0 {
			log.Printf("[WARN] Cmd must be defined")
			continue
		}

		h.HandleFunc(v.Path, server.JSONRequestMiddleware(server.RepoRequestHandler(cmdLog, k, v)))
		hooksHandled[v.Path] = 1
	}
	return
}

func main() {
	flag.Parse()

	config, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	h := http.NewServeMux()
	commandLog := server.NewMemoryCommandLog()
	h.HandleFunc("/hello", server.JSONRequestMiddleware(server.HelloHandler))
	h.HandleFunc("/admin/cmdlog", server.JSONRequestMiddleware(server.CommandLogRESTHandler(&commandLog)))
	hooksHandled := addHandlers(&commandLog, config, h)

	log.Printf("Added %d hooks (%v):", len(hooksHandled), hooksHandled)
	log.Printf("Starting web server at %s:%d\n", config.Address, config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), h))
}
