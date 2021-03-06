package event

import (
	"encoding/json"
	"errors"
	"net/http"
)

type bitbucketPayloadType struct {
	Push push
}

type push struct {
	Changes []change
}

type change struct {
	Old changeStruct
	New changeStruct
}

type changeStruct struct {
	Type   string
	Name   string
	Target target
}

type target struct {
	Hash   string
	Author author
}

type author struct {
	User actor
}

type actor struct {
	Username string
}

// NewBitbucketEvent takes an http.Request object and parses it corresponding
// to Bitbucket webhook syntax into an RepoEvent object.
// It returns a RepoEvent object and an error in case of error
func NewBitbucketEvent(request *http.Request) (event *RepoEvent, err error) {
	if request.Body == nil {
		err = errors.New("Unable to parse request.Body == nil")
		return
	}
	var parsedPayload bitbucketPayloadType
	var branch, author, commit string
	err = json.NewDecoder(request.Body).Decode(&parsedPayload)

	if err != nil {
		return
	}

	if len(parsedPayload.Push.Changes) > 0 {
		branch = parsedPayload.Push.Changes[0].New.Name
		commit = parsedPayload.Push.Changes[0].New.Target.Hash
		author = parsedPayload.Push.Changes[0].New.Target.Author.User.Username
	} else {
		err = errors.New("Changes array should contain at least 1 element, got 0")
		return
	}
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
