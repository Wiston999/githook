package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Wiston999/githook/event"
	"github.com/Wiston999/githook/server"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Hooks map[string]event.Hook
}

// parseHooks parses a YAML configuration file given its filename
// It returns a map of [string]event.Hook structure and error in case of errors
func parseHooks(configFile string) (hooks map[string]event.Hook, err error) {
	filename, err := filepath.Abs(configFile)
	if err != nil {
		return
	}
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return
	}
	hooks = config.Hooks
	return
}

// addHandlers configures hook handlers into an http.ServeMux handler given a map of hooks
// It returns a map containing the hooks added as key
func addHandlers(cmdLog server.CommandLog, hooks map[string]event.Hook, h *http.ServeMux) (hooksHandled map[string]int) {
	hooksHandled = make(map[string]int)
	for k, v := range hooks {
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

func setupLogLevel(logLevel string) {
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warn("Unknown log level ", logLevel, " falling back to debug")
		parsedLogLevel = log.DebugLevel
	}
	log.SetLevel(parsedLogLevel)
}

func setupCommandLog(commandLogDir string) (cmdLog server.CommandLog) {
	cmdLog = server.NewMemoryCommandLog()
	defer func() {
		switch cmdLog.(type) {
		case *server.MemoryCommandLog:
			log.Warn("CommandLogDir setting not found or invalid, using in memory command log")
		case *server.DiskCommandLog:
			log.Info("Commands will be logged to", commandLogDir)
		}
	}()
	if commandLogDir == "" {
		return
	}

	absLogDir, absErr := filepath.Abs(commandLogDir)
	if absErr != nil {
		return
	}
	fileMode, statErr := os.Stat(absLogDir)
	if statErr == nil && fileMode.IsDir() {
		cmdLog = server.NewDiskCommandLog(commandLogDir)
	}
	return
}

var opts struct {
	ConfigFile string `short:"c" long:"config" description:"Configuration file location"`
	Addr       string `long:"address" default:"0.0.0.0" description:"Server listening(bind) address"`
	Port       int    `short:"p" long:"port" default:"65000" description:"Server listening port"`
	LogDir     string `long:"command_log_dir" description:"CommandLogDir to store requests' results leave empty to use in-memory storage"`
	LogLevel   string `long:"loglvl" default:"warn" value-name:"choices" choice:"err" choice:"warning" choice:"warn" choice:"info" choice:"debug" description:"Log facility level"`
}

func main() {
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	setupLogLevel(opts.LogLevel)
	commandLog := setupCommandLog(opts.LogDir)
	hooks, err := parseHooks(opts.ConfigFile)
	log.WithFields(log.Fields{"hooks": hooks}).Debug("Hooks parsed from configuration file")
	if err != nil {
		log.Fatal(err)
	}

	h := http.NewServeMux()
	h.HandleFunc("/hello", server.JSONRequestMiddleware(server.HelloHandler))
	h.HandleFunc("/admin/cmdlog", server.JSONRequestMiddleware(server.CommandLogRESTHandler(commandLog)))
	hooksHandled := addHandlers(commandLog, hooks, h)

	if len(hooksHandled) == 0 {
		log.Fatal("No hooks will be handled, I'm useless")
	}

	log.WithFields(log.Fields{"hooks": hooksHandled}).Info("Added ", len(hooksHandled), " hooks")
	log.WithFields(log.Fields{"addr": opts.Addr, "port": opts.Port}).Debug("Starting web server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", opts.Addr, opts.Port), h))
}
