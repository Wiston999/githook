package main

import (
	"testing"
)

func TestParseConfig(t *testing.T) {
	_, err := parseConfig("")
	if err == nil {
		t.Error("An error must be returned when configuration file is empty")
	}

	_, err = parseConfig("../examples/non_existent_file.yaml")
	if err == nil {
		t.Error("An error must be returned when configuration file does not exist")
	}

	config, err := parseConfig("./examples/bitbucket.org.yaml")

	if err != nil {
		t.Error("An error must not be return with proper file")
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
	if config.Hooks["bitbucket.org"].Type != "bitbucket" || config.Hooks["bitbucket.org"].Path != "/payload" || config.Hooks["bitbucket.org"].Cmd != "echo {{branch}}" {
		t.Error("Error parsing hook 0")
	}
	if config.Hooks["github.com"].Type != "github" || config.Hooks["github.com"].Path != "/github" || config.Hooks["github.com"].Cmd != "echo {{branch}}" {
		t.Error("Error parsing hook 1")
	}
}

func TestParseYAML(t *testing.T) {
	_, err := parseYAML([]byte(`---`))
	if err != nil {
		t.Error("An error must not be return with empty YAML file")
	}

	_, err = parseYAML([]byte(`{}[]-:this is an invalid: YAML file`))
	if err == nil {
		t.Error("An error must be returned with invalid YAML file")
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
