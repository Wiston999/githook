package event

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type githubPayloadType struct {
	Ref         string
	Head_commit headCommit
}

type headCommit struct {
	Id     string
	Author commitAuthor
}

type commitAuthor struct {
	Username string
}

type GithubEvent struct {
	RepoEvent
}

func NewGithubEvent(request *http.Request) (event *GithubEvent, err error) {
	var payload []byte
	switch request.Header.Get("Content-Type") {
	case "application/json":
		payload, err = ioutil.ReadAll(request.Body)
	case "application/x-www-form-urlencoded":
		err = request.ParseForm()
		payload = []byte(request.PostFormValue("payload"))
	}

	if err != nil {
		return nil, err
	}

	var parsedPayload githubPayloadType
	var branch, author, commit string
	err = json.Unmarshal(payload, &parsedPayload)
	if err != nil {
		return nil, err
	}

	sl := strings.Split(parsedPayload.Ref, "/")
	branch = sl[len(sl)-1]
	commit = parsedPayload.Head_commit.Id
	author = parsedPayload.Head_commit.Author.Username

	if branch == "" {
		err = errors.New("Unable to parse branch")
	}
	if commit == "" {
		err = errors.New("Unable to parse commit")
	}
	if author == "" {
		err = errors.New("Unable to parse author")
	}
	event = &GithubEvent{RepoEvent{Author: author, Branch: branch, Commit: commit}}
	return
}
