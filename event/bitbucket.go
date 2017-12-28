package event

import (
	"encoding/json"
	"errors"
	"io/ioutil"
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

type BitbucketEvent struct {
	RepoEvent
}

func NewBitbucketEvent(request *http.Request) (event *BitbucketEvent, err error) {
	var payload []byte
	payload, err = ioutil.ReadAll(request.Body)

	if err != nil {
		return
	}
	var parsedPayload bitbucketPayloadType
	var branch, author, commit string
	json.Unmarshal(payload, &parsedPayload)

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
	event = &BitbucketEvent{RepoEvent{Author: author, Branch: branch, Commit: commit}}
	return
}
