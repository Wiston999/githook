package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Wiston999/githook/event"
	"github.com/Wiston999/githook/server"
)

func TestParseConfig(t *testing.T) {
	_, _, err := parseConfig("")
	if err == nil {
		t.Error("An error must be returned when configuration file is empty")
	}

	_, _, err = parseConfig("../examples/non_existent_file.yaml")
	if err == nil {
		t.Error("An error must be returned when configuration file does not exist")
	}

	config, cmdLog, err := parseConfig("./examples/bitbucket.org.yaml")

	if err != nil {
		t.Errorf("An error must not be return with proper file: %s", err)
	}
	if config.Address != "0.0.0.0" {
		t.Error("Parsed address must be 0.0.0.0")
	}
	if config.Port != 65000 {
		t.Error("Parsed port must be 65000")
	}
	if len(config.Hooks) != 2 {
		t.Error("Parsed hooks must be 2")
	}
	switch cmdLog.(type) {
	case *server.MemoryCommandLog:
	default:
		t.Error("CommandLog type should be MemoryCommandLog when no command log dir is set")
	}
	condition := config.Hooks["bitbucket.org"].Type != "bitbucket"
	condition = condition || config.Hooks["bitbucket.org"].Path != "/payload"
	condition = condition || config.Hooks["bitbucket.org"].Timeout != 600
	condition = condition || strings.Join(config.Hooks["bitbucket.org"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 0")
	}
	condition = config.Hooks["github.com"].Type != "github"
	condition = condition || config.Hooks["github.com"].Path != "/github"
	condition = condition || config.Hooks["github.com"].Timeout != 0
	condition = condition || strings.Join(config.Hooks["github.com"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 1")
	}
}

func TestParseYAML(t *testing.T) {
	_, _, err := parseYAML([]byte(`---`))
	if err != nil {
		t.Error("An error must not be return with empty YAML file")
	}

	_, _, err = parseYAML([]byte(`{}[]-:this is an invalid: YAML file`))
	if err == nil {
		t.Error("An error must be returned with invalid YAML file")
	}

	config, cmdLog, err := parseYAML([]byte(`---
address: 127.0.0.1
port: 8080
command_log_dir: './'
hooks:
  test1:
    type: bitbucket
    path: /test1
    cmd: [echo, '{{.Branch}}']
    timeout: 1
  test2:
    type: github
    path: /test2
    timeout: 2
    cmd: [echo, '{{.Branch}}']`))

	if err != nil {
		t.Error("An error must not be return with proper file")
	}
	if config.Address != "127.0.0.1" {
		t.Error("Parsed address must be 127.0.0.1")
	}
	if config.Port != 8080 {
		t.Error("Parsed port must be 8080")
	}
	if len(config.Hooks) != 2 {
		t.Error("Parsed hooks must be 2")
	}
	switch cmdLog.(type) {
	case *server.DiskCommandLog:
	default:
		t.Error("CommandLog should be DiskCommandLog when command_log_dir is present and exists")
	}
	condition := config.Hooks["test1"].Type != "bitbucket"
	condition = condition || config.Hooks["test1"].Path != "/test1"
	condition = condition || config.Hooks["test1"].Timeout != 1
	condition = condition || strings.Join(config.Hooks["test1"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 0")
	}
	condition = config.Hooks["test2"].Type != "github"
	condition = condition || config.Hooks["test2"].Path != "/test2"
	condition = condition || config.Hooks["test2"].Timeout != 2
	condition = condition || strings.Join(config.Hooks["test2"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 1")
	}
}

func TestAddHandlers(t *testing.T) {
	h := http.NewServeMux()

	hooks := make(map[string]event.Hook)
	hooks["test1"] = event.Hook{Type: "github", Path: "/github1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test2"] = event.Hook{Type: "github", Path: "/github2", Cmd: []string{"true"}, Timeout: 500}
	hooks["test3"] = event.Hook{Type: "github", Path: "/github2", Cmd: []string{"true"}, Timeout: 500}
	hooks["test4"] = event.Hook{Type: "bitbucket", Path: "/bitbucket1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test5"] = event.Hook{Type: "gitlab", Path: "/gitlab1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test6"] = event.Hook{Type: "gitlab", Path: "invalid", Cmd: []string{"true"}, Timeout: 500}
	hooks["test7"] = event.Hook{Type: "gitlab", Path: "/hello", Cmd: []string{"true"}, Timeout: 500}
	hooks["test8"] = event.Hook{Type: "invalid", Path: "/invalid1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test9"] = event.Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{}, Timeout: 500}
	hooks["test10"] = event.Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{"true"}, Timeout: 0}
	hooks["test11"] = event.Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{"true"}, Timeout: -10}

	cmdLog := server.NewMemoryCommandLog()
	config := Config{Hooks: hooks}
	hooksHandled := addHandlers(cmdLog, config, h)
	removed := map[string]string{
		"test3":  "Duplicated Path",
		"test6":  "Invalid path (must start with /)",
		"test7":  "/hello is a reserved path",
		"test8":  "Invalid type",
		"test9":  "Invalid Cmd (must be present)",
		"test10": "Timeout must be greater than 0",
		"test11": "Timeout must be greater than 0",
	}

	if len(hooksHandled) != (len(hooks) - len(removed)) {
		t.Errorf("Only %d hooks should have been added, got %d", len(hooks)-len(removed), len(hooksHandled))
	}

	for k, v := range removed {
		if _, found := hooksHandled[k]; found {
			t.Errorf("Case %s should have been removed due to: %s", k, v)
		}
	}
}
