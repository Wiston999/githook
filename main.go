package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wiston999/githook/event"
	"github.com/Wiston999/githook/server"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var configFile = flag.String("config", "", "Configuration file")

type Config struct {
	// Address where the HTTP server will be bind
	Address string
	// Port where the HTTP server will be listening
	Port          int
	CommandLogDir string `yaml:"command_log_dir"`
	Hooks         map[string]event.Hook
}

// parseConfig parses a YAML configuration file given its filename
// It returns a Config structure and error in case of errors
func parseConfig(configFile string) (config Config, cmdLog server.CommandLog, err error) {
	filename, err := filepath.Abs(configFile)
	if err != nil {
		return
	}
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return
	}
	return parseYAML(yamlFile)
}

// parseYAML parses a YAML configuration file given its string representation
// It returns a Config structure and error in case of errors
func parseYAML(yamlFile []byte) (config Config, cmdLog server.CommandLog, err error) {
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return
	}

	if config.Address == "" {
		config.Address = "0.0.0.0"
	}
	if config.Port == 0 {
		config.Port = 65000
	}
	fileMode, statErr := os.Stat(config.CommandLogDir)
	if statErr == nil && fileMode.IsDir() {
		cmdLog = server.NewDiskCommandLog(config.CommandLogDir)
	} else {
		log.Warn("command_log_dir setting not found or invalid, using in memory command log")
		cmdLog = server.NewMemoryCommandLog()
	}
	return
}

// addHandlers configures hook handlers into an http.ServeMux handler given a Config structure
// It returns a map containing the hooks added as key
func addHandlers(cmdLog server.CommandLog, config Config, h *http.ServeMux) (hooksHandled map[string]int) {
	hooksHandled = make(map[string]int)
	for k, v := range config.Hooks {
		log.WithFields(log.Fields{
			"name": k,
			"hook": v,
		}).Info("Read hook")
		if _, exists := hooksHandled[v.Path]; exists {
			log.Warn("Path ", v.Path, " already defined, ignoring...")
			continue
		}
		if v.Type != "bitbucket" && v.Type != "github" && v.Type != "gitlab" {
			log.Warn("Unknown repository type, it must be one of: bitbucket, github or gitlab")
			continue
		}
		if !strings.HasPrefix(v.Path, "/") || v.Path == "/hello" {
			log.Warn("Path must start with / and be different of /hello")
			continue
		}
		if v.Timeout <= 0 {
			log.Warn("Timeout must be greater than 0, got ", v.Timeout)
			continue
		}
		if len(v.Cmd) == 0 {
			log.Warn("Cmd must be defined")
			continue
		}

		h.HandleFunc(v.Path, server.JSONRequestMiddleware(server.RepoRequestHandler(cmdLog, k, v)))
		hooksHandled[v.Path] = 1
	}
	return
}

func main() {
	flag.Parse()

	config, commandLog, err := parseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	h := http.NewServeMux()
	h.HandleFunc("/hello", server.JSONRequestMiddleware(server.HelloHandler))
	h.HandleFunc("/admin/cmdlog", server.JSONRequestMiddleware(server.CommandLogRESTHandler(commandLog)))
	hooksHandled := addHandlers(commandLog, config, h)

	log.WithFields(log.Fields{"hooks": hooksHandled}).Info("Added ", len(hooksHandled), " hooks")
	log.WithFields(log.Fields{"addr": config.Address, "port": config.Port}).Debug("Starting web server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), h))
}
