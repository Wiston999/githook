package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Wiston999/githook/event"
	"github.com/Wiston999/githook/server"

	log "github.com/sirupsen/logrus"
)

func TestParseHooks(t *testing.T) {
	_, err := parseHooks("")
	if err == nil {
		t.Error("An error must be returned when configuration file is empty")
	}

	_, err = parseHooks("../examples/non_existent_file.yaml")
	if err == nil {
		t.Error("An error must be returned when configuration file does not exist")
	}

	hooks, err := parseHooks("./examples/bitbucket.org.yaml")

	if err != nil {
		t.Errorf("An error must not be return with proper file: %s", err)
	}
	condition := hooks["bitbucket.org"].Type != "bitbucket"
	condition = condition || hooks["bitbucket.org"].Path != "/payload"
	condition = condition || hooks["bitbucket.org"].Timeout != 600
	condition = condition || strings.Join(hooks["bitbucket.org"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 0")
	}
	condition = hooks["github.com"].Type != "github"
	condition = condition || hooks["github.com"].Path != "/github"
	condition = condition || hooks["github.com"].Timeout != 0
	condition = condition || strings.Join(hooks["github.com"].Cmd, " ") != "echo {{.Branch}}"
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
	hooksHandled := addHandlers(cmdLog, hooks, h)
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

func TestSetupLogLevel(t *testing.T) {
	testCases := []struct {
		level    string
		expected log.Level
	}{
		{"error", log.ErrorLevel},
		{"warn", log.WarnLevel},
		{"warning", log.WarnLevel},
		{"info", log.InfoLevel},
		{"debug", log.DebugLevel},
		{"unknown", log.DebugLevel},
	}

	for i, test := range testCases {
		setupLogLevel(test.level)
		level := log.GetLevel()
		if level != test.expected {
			t.Errorf("%02d. Log level is not the expected, got %v but expected %v", i, level, test.expected)
		}
	}
}

func TestSetupCommandLog(t *testing.T) {
	cmdLog := setupCommandLog("")
	switch v := cmdLog.(type) {
	case *server.MemoryCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected MemoryCommandLog", v)
	}

	cmdLog = setupCommandLog("/notfound")
	switch v := cmdLog.(type) {
	case *server.MemoryCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected MemoryCommandLog", v)
	}
	cmdLog = setupCommandLog("./")
	switch v := cmdLog.(type) {
	case *server.DiskCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected DiskCommandLog", v)
	}
}

func TestSetupWebServer(t *testing.T) {
	cmdLog := server.NewMemoryCommandLog()
	hooks := make(map[string]event.Hook)
	hooks["test1"] = event.Hook{Type: "github", Path: "/github1", Cmd: []string{"true"}, Timeout: 500}

	server, err := setupWebServer("127.0.0.1", 10000, cmdLog, hooks)
	if err != nil {
		t.Errorf("setupWebServer should not fail with proper args: %s", err)
	}
	if server.Addr != "127.0.0.1:10000" {
		t.Errorf("Server should have been setup to listen on 127.0.0.1:10000")
	}

	hooks = make(map[string]event.Hook)

	_, err = setupWebServer("127.0.0.1", 10000, cmdLog, hooks)
	if err == nil {
		t.Errorf("setupWebServer should fail with no hooks to serve")
	}
}
