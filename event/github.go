package event

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
)

// NewGithubEvent takes an http.Request object and parses it corresponding
// to Github webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewGithubEvent(request *http.Request) (event *RepoEvent, err error) {
	var payload string
	if request.Body == nil {
		err = errors.New("Unable to parse request.Body == nil")
		return
	}
	switch request.Header.Get("Content-Type") {
	case "application/json":
		buf := new(bytes.Buffer)
		buf.ReadFrom(request.Body)
		payload = buf.String()
	case "application/x-www-form-urlencoded":
		err = request.ParseForm()
		payload = request.PostFormValue("payload")
	}

	if err != nil {
		return nil, err
	}

	sl := strings.Split(gjson.Get(payload, "ref").String(), "/")
	branch := sl[len(sl)-1]
	commit := gjson.Get(payload, "head_commit.id").String()
	author := gjson.Get(payload, "head_commit.author.username").String()

	fmt.Printf("Author: %s\n", author)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Printf("Commit: %s\n", commit)
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
