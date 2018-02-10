package event

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/tidwall/gjson"
)

// NewBitbucketEvent takes an http.Request object and parses it corresponding
// to Bitbucket webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewBitbucketEvent(request *http.Request) (event *RepoEvent, err error) {
	if request.Body == nil {
		err = errors.New("Unable to parse request.Body == nil")
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	bodyStr := buf.String()
	if err != nil {
		return
	}

	branch := gjson.Get(bodyStr, "push.changes.0.new.name").String()
	commit := gjson.Get(bodyStr, "push.changes.0.new.target.hash").String()
	author := gjson.Get(bodyStr, "push.changes.0.new.target.author.user.username").String()

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
