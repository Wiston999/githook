package event

import (
	"net/http"
)

// NewGitlabEvent takes an http.Request object and parses it corresponding
// to Gitlab webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewGitlabEvent(request *http.Request) (event *RepoEvent, err error) {
	return &RepoEvent{}, nil
}
