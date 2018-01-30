package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Wiston999/githook/server"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config stores the hooks configuration for the process
type Config struct {
	Hooks map[string]server.Hook
}

// parseHooks parses a YAML configuration file given its filename
// It returns a map of [string]server.Hook structure and error in case of errors
func parseHooks(configFile string) (hooks map[string]server.Hook, err error) {
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

func setupLogLevel(logLevel string) {
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warn("Unknown log level ", logLevel, " falling back to debug")
		parsedLogLevel = log.DebugLevel
	}
	log.SetLevel(parsedLogLevel)
}

var opts struct {
	ConfigFile string `short:"c" long:"config" description:"Configuration file location"`
	Addr       string `long:"address" default:"0.0.0.0" description:"Server listening(bind) address"`
	Port       int    `short:"p" long:"port" default:"65000" description:"Server listening port"`
	LogDir     string `long:"command_log_dir" description:"CommandLogDir to store requests' results leave empty to use in-memory storage"`
	LogLevel   string `long:"loglvl" default:"warn" value-name:"choices" choice:"err" choice:"warning" choice:"warn" choice:"info" choice:"debug" description:"Log facility level"`
	TLSCert    string `long:"tlscert" description:"Certificate file for TLS support"`
	TLSKey     string `long:"tlskey" description:"Key file for TLS support, TLS is tried if both tlscert and tlskey are provided"`
}

func main() {
	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	setupLogLevel(opts.LogLevel)
	hooks, err := parseHooks(opts.ConfigFile)
	log.WithFields(log.Fields{"hooks": hooks}).Debug("Hooks parsed from configuration file")
	if err != nil {
		log.Fatal(err)
	}

	server := server.Server{
		Server:    &http.Server{Addr: fmt.Sprintf("%s:%d", opts.Addr, opts.Port)},
		TLSCert:   opts.TLSCert,
		TLSKey:    opts.TLSKey,
		CmdLogDir: opts.LogDir,
		Hooks:     hooks,
	}
	log.WithFields(log.Fields{"addr": opts.Addr, "port": opts.Port}).Debug("Starting web server")
	log.Fatal(server.ListenAndServe())
}
