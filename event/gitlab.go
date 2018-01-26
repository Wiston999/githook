package event

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type gitlabPayloadType struct {
	Ref          string `json:"ref" yaml:"ref"`
	UserUsername string `json:"user_username" yaml:"user_username"`
	CheckoutSha  string `json:"checkout_sha" yaml:"checkout_sha"`
}

// NewGitlabEvent takes an http.Request object and parses it corresponding
// to Gitlab webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewGitlabEvent(request *http.Request) (event *RepoEvent, err error) {
	if request.Body == nil {
		err = errors.New("Unable to parse request.Body == nil")
		return
	}
	var parsedPayload gitlabPayloadType
	var branch, author, commit string
	err = json.NewDecoder(request.Body).Decode(&parsedPayload)

	if err != nil {
		return
	}

	sl := strings.Split(parsedPayload.Ref, "/")
	branch = sl[len(sl)-1]
	commit = parsedPayload.CheckoutSha
	author = parsedPayload.UserUsername

	if branch == "" {
		err = errors.New("Unable to parse branch")
	}
	if commit == "" {
		err = errors.New("Unable to parse commit")
	}
	if author == "" {
		err = errors.New("Unable to parse author")
	}
	event = &RepoEvent{Author: author, Branch: branch, Commit: commit}
	return
}
