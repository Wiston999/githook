package event

import (
	"bytes"
	"errors"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

// NewGitlabEvent takes an http.Request object and parses it corresponding
// to Gitlab webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewGitlabEvent(request *http.Request) (event *RepoEvent, err error) {
	if request.Body == nil {
		err = errors.New("Unable to parse request.Body == nil")
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	payload := buf.String()

	sl := strings.Split(gjson.Get(payload, "ref").String(), "/")
	branch := sl[len(sl)-1]
	commit := gjson.Get(payload, "checkout_sha").String()
	author := gjson.Get(payload, "user_username").String()

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
