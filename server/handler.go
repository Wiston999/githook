package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/wiston999/githook/event"

	"github.com/nu7hatch/gouuid"
)

func HelloHandler(w http.ResponseWriter, req *http.Request) {
	response := Response{Status: 200, Msg: "Hello from githook listener"}
	json.NewEncoder(w).Encode(response)
}

func JSONRequestMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId, _ := uuid.NewV4()
		log.Printf("[INFO] Received request (%s) '%s' %s", requestId, r.URL.Path, r.Method)
		defer log.Printf("[INFO] Request (%s) completed", requestId)

		w.Header().Set("Content-Type", "application/json")
		h.ServeHTTP(w, r)
	})
}

func RepoRequestHandler(hookName string, hookInfo event.Hook) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		repoEvent := &event.RepoEvent{}
		var response Response
		var err error

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
	}
}
