package event

import (
	"net/http"
)

func NewGitlabEvent(request *http.Request) (event *RepoEvent, err error) {
	return &RepoEvent{}, nil
}
