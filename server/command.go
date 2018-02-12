package server

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os/exec"
	"text/template"
	"time"

	"github.com/Wiston999/githook/event"
)

// CommandResult stores the result of a command execution
type CommandResult struct {
	Cmd    []string `json:"cmd"`
	Err    error    `json:"err"`
	Stdout []byte   `json:"stdout"`
	Stderr []byte   `json:"stderr"`
}

// TranslateParams translates a list of command parameters (from Hook) based
// on the event received at event.RepoEvent. It uses Go's built-in templating (text/template)
// so all operations on templates can be performed on the command parameters.
// I.e.: `cmd := ["{{.Branch}}", "is", "the", "branch"]` with event.Branch := "develop" will
// be transformed to `cmd := ["develop", "is", "the", "branch"]`
// It returns the translated array of strings and error in case of error
func TranslateParams(cmd []string, event event.RepoEvent) (trCmd []string, err error) {
	for _, arg := range cmd {
		tpl, tmpErr := template.New("cmd-template").Parse(arg)
		if tmpErr != nil {
			err = tmpErr
			return
		}
		buffer := new(bytes.Buffer)

		tmpErr = tpl.Execute(buffer, event)
		if tmpErr != nil {
			err = tmpErr
			return
		}
		trCmd = append(trCmd, buffer.String())
	}
	return
}

// RunCommand executes the hook command on the system, it takes an array of string
// representing the command to be returned, a timeout in seconds and a channel for returning the data.
// It returns an instance of CommandResult
func RunCommand(cmd []string, timeout int) (result CommandResult) {
	result.Cmd = cmd
	if len(cmd) == 0 {
		result.Err = errors.New("Empty command string cannot be run")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	stderr, err := command.StderrPipe()
	if err != nil {
		result.Err = err
		return
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		result.Err = err
		return
	}

	if err := command.Start(); err != nil {
		result.Err = err
		return
	}

	result.Stdout, _ = ioutil.ReadAll(stdout)
	result.Stderr, _ = ioutil.ReadAll(stderr)

	if err := command.Wait(); err != nil {
		result.Err = err
	}
	return
}
