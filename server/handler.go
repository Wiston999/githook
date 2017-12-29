package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/wiston999/githook/event"

	"github.com/nu7hatch/gouuid"
)

// HelloHandler implements a basic HTTP Handler that returns a HelloWorld-like response
// for testing or debugging purposes
func HelloHandler(w http.ResponseWriter, req *http.Request) {
	response := Response{Status: 200, Msg: "Hello from githook listener"}
	json.NewEncoder(w).Encode(response)
}

// JSONRequestMiddleware implements an http.HandlerFunc middleware that sets
// the HTTP Content-Type header and prints a log line when the request is received and completed
func JSONRequestMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId, _ := uuid.NewV4()
		log.Printf("[INFO] Received request (%s) '%s' %s", requestId, r.URL.Path, r.Method)
		defer log.Printf("[INFO] Request (%s) completed", requestId)

		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}

// RepoRequestHandler setups an http.HandlerFunc using event.Hook information
// This function makes the hard work of setting up a listener hook on the HTTP Server
// based on an event.Hook structure
func RepoRequestHandler(hookName string, hookInfo event.Hook) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		repoEvent := &event.RepoEvent{}
		var response Response
		var err error
		urlQuery := r.URL.Query()
		_, sync := urlQuery["sync"]

		switch hookInfo.Type {
		case "bitbucket":
			localEvent, localErr := event.NewBitbucketEvent(r)
			if localErr != nil {
				repoEvent, err = nil, localErr
			} else {
				repoEvent, err = localEvent, localErr
			}
		case "github":
			localEvent, localErr := event.NewGithubEvent(r)
			if localErr != nil {
				repoEvent, err = nil, localErr
			} else {
				repoEvent, err = localEvent, localErr
			}
		case "gitlab":
			localEvent, localErr := event.NewGitlabEvent(r)
			if localErr != nil {
				repoEvent, err = nil, localErr
			} else {
				repoEvent, err = localEvent, localErr
			}
		}

		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("Error while parsing event: %s", err)
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(response)
			return
		}

		if repoEvent.Branch == "" {
			response.Status, response.Msg = 500, "Repository type is unknown"
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(response)
			return
		}

		log.Printf("[DEBUG] Repository event parsed: %#v", repoEvent)
		cmd, err := TranslateParams(hookInfo.Cmd, *repoEvent)
		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("Unable to translate hook command template (%s): %s", hookName, err)
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(response)
			return
		}

		channel := make(chan CommandResult)
		go RunCommand(cmd, hookInfo.Timeout, channel)
		if sync {
			result := <-channel
			response.Status = 200
			response.Msg = fmt.Sprintf(
				"Command '%s' sent to execute with result {Err: '%v', stdout: '%s', stderr: '%s'}",
				strings.Join(cmd, " "),
				result.Err,
				result.Stdout,
				result.Stderr,
			)
			json.NewEncoder(w).Encode(response)

		} else {
			response.Status, response.Msg = 200, fmt.Sprintf("Command '%s' sent to execute", strings.Join(cmd, " "))
			json.NewEncoder(w).Encode(response)
		}
	}
}
