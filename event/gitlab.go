package event

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type gitlabPayloadType struct {
	Ref           string
	User_username string
	Checkout_sha  string
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
	commit = parsedPayload.Checkout_sha
	author = parsedPayload.User_username

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
