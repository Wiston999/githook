package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wiston999/githook/event"

	"github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
)

var workerChannels map[string]chan CommandJob

func init() {
	workerChannels = make(map[string]chan CommandJob)
}

// JSONRequestMiddleware implements an http.HandlerFunc middleware that sets
// the HTTP Content-Type header and prints a log line when the request is received and completed
func JSONRequestMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, _ := uuid.NewV4()
		requestIDStr := requestID.String()
		log.WithFields(log.Fields{"reqId": requestID, "url": r.URL.Path, "method": r.Method}).Info("Received request")
		defer log.WithFields(log.Fields{"reqID": requestID}).Info("Request completed")

		w.Header().Set("Content-Type", "application/json")
		ctx := context.WithValue(r.Context(), "requestID", requestIDStr)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HelloHandler implements a basic HTTP Handler that returns a HelloWorld-like response
// for testing or debugging purposes
func HelloHandler(w http.ResponseWriter, req *http.Request) {
	response := Response{Status: 200, Msg: "Hello from githook listener"}
	json.NewEncoder(w).Encode(response)
}

// CommandLogRESTHandler Returns the command log via REST request
func CommandLogRESTHandler(cmdLog CommandLog) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response := Response{Status: 200, Msg: "success"}
		urlQuery := r.URL.Query()
		count := urlQuery.Get("count")

		var countInt int
		var err error
		if count == "" {
			countInt = -1
		} else {
			countInt, err = strconv.Atoi(count)
		}

		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("%s", err)
		}

		results, err := cmdLog.GetResults(countInt)
		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("%s", err)
		} else {
			response.Body = results
		}
		json.NewEncoder(w).Encode(response)
	}
}

// RepoRequestHandler setups an http.HandlerFunc using Hook information
// This function makes the hard work of setting up a listener hook on the HTTP Server
// based on an Hook structure
func RepoRequestHandler(cmdLog CommandLog, hookName string, hookInfo Hook) func(http.ResponseWriter, *http.Request) {
	workerChannels[hookName] = make(chan CommandJob, 1000)
	for i := 0; i < hookInfo.Concurrency; i++ {
		go CommandWorker(hookName, workerChannels[hookName], cmdLog)
	}
	log.WithFields(log.Fields{
		"count": hookInfo.Concurrency,
		"hook":  hookName,
	}).Info("Started command workers")

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value("requestID").(string)
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

		log.Debug("Repository event parsed: ", repoEvent)
		cmd, err := TranslateParams(hookInfo.Cmd, *repoEvent)
		if err != nil {
			response.Status, response.Msg = 500, fmt.Sprintf("Unable to translate hook command template (%s): %s", hookName, err)
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(response)
			return
		}

		cmdJob := CommandJob{Cmd: cmd, ID: requestID, Timeout: hookInfo.Timeout}
		if sync {
			cmdJob.Response = make(chan CommandResult, 1)
		}
		workerChannels[hookName] <- cmdJob
		response.Status, response.Msg, response.Body = 200, "Command sent to execute", strings.Join(cmd, " ")
		if sync {
			log.WithFields(log.Fields{
				"cmd":       cmdJob.Cmd,
				"queue_len": len(workerChannels[hookName]),
				"reqId":     requestID,
			}).Info("Waiting for command to complete before returning")

			result := <-cmdJob.Response
			response.Body = result
		}
		json.NewEncoder(w).Encode(response)
	}
}
