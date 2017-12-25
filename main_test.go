package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	_, err := parseConfig("")
	if err == nil {
		t.Error("An error must be returned when configuration file is empty")
	}
}

func TestParseYAML(t *testing.T) {
	_, err := parseYAML([]byte(`---`))
	if err != nil {
		t.Error("An error must not be return with empty YAML file")
	}

	config, err := parseYAML([]byte(`---
address: 127.0.0.1
port: 8080
hooks:
  test1:
    type: bitbucket
    path: /test1
    cmd: 'echo {{branch}}'
  test2:
    type: github
    path: /test2
    cmd: 'echo {{branch}}'`))

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
	if config.Hooks["test1"].Type != "bitbucket" || config.Hooks["test1"].Path != "/test1" || config.Hooks["test1"].Cmd != "echo {{branch}}" {
		t.Error("Error parsing hook 0")
	}
	if config.Hooks["test2"].Type != "github" || config.Hooks["test2"].Path != "/test2" || config.Hooks["test2"].Cmd != "echo {{branch}}" {
		t.Error("Error parsing hook 1")
	}
}
