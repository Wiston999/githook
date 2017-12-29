package server

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/wiston999/githook/event"
)

type CommandResult struct {
	Err    error
	Stdout []byte
	Stderr []byte
}

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

func RunCommand(cmd []string, timeout int, ch chan CommandResult) {
	result := CommandResult{}
	if len(cmd) == 0 {
		result.Err = errors.New("Empty command string present")
		ch <- result
		return
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
		stderr, err := command.StderrPipe()
		if err != nil {
			result.Err = err
			ch <- result
			return
		}

		stdout, err := command.StdoutPipe()
		if err != nil {
			result.Err = err
			ch <- result
			return
		}

		if err := command.Start(); err != nil {
			result.Err = err
			ch <- result
			return
		}

		result.Stdout, _ = ioutil.ReadAll(stdout)
		result.Stderr, _ = ioutil.ReadAll(stderr)

		if err := command.Wait(); err != nil {
			result.Err = err
		}
		log.Printf("Command '%s' executed (Err: %v, STDOUT: %s, STDERR: %s)", strings.Join(cmd, " "), result.Err, result.Stdout, result.Stderr)
		ch <- result
	}
}
