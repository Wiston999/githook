package main

import (
	"strings"
	"testing"

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
	condition = condition || hooks["bitbucket.org"].Concurrency != 1
	condition = condition || strings.Join(hooks["bitbucket.org"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 0")
	}
	condition = hooks["github.com"].Type != "github"
	condition = condition || hooks["github.com"].Path != "/github"
	condition = condition || hooks["github.com"].Timeout != 0
	condition = condition || hooks["github.com"].Concurrency != 0
	condition = condition || strings.Join(hooks["github.com"].Cmd, " ") != "echo {{.Branch}}"
	if condition {
		t.Error("Error parsing hook 1")
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
