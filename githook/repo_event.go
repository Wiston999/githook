package githook

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type RepoEvent struct {
	Author string
	Branch string
	Commit string
}

type GithubEvent RepoEvent

func NewGithubEvent(request *http.Request) GithubEvent {
	var payload []byte
	switch request.Header.Get("Content-Type") {
	case "application/json":
		payload, _ = ioutil.ReadAll(request.Body)
	case "application/x-www-form-urlencoded":
		request.ParseForm()
		payload = []byte(request.PostFormValue("payload"))
	}

	var parsedPayload, tmpPayload map[string]*json.RawMessage
	var ref, branch, author, commit string
	json.Unmarshal(payload, &parsedPayload)
	json.Unmarshal(*parsedPayload["ref"], &ref)
	sl := strings.Split(ref, "/")
	branch = sl[len(sl)-1]
	json.Unmarshal(*parsedPayload["head_commit"], &tmpPayload)
	json.Unmarshal(*tmpPayload["id"], &commit)
	json.Unmarshal(*parsedPayload["head_commit"], &tmpPayload)
	json.Unmarshal(*tmpPayload["author"], &tmpPayload)
	json.Unmarshal(*tmpPayload["username"], &author)

	return GithubEvent{Author: author, Branch: branch, Commit: commit}
}

type BitbucketEvent RepoEvent

func NewBitbucketEvent(request *http.Request) BitbucketEvent {
	var payload []byte
	payload, _ = ioutil.ReadAll(request.Body)

	var parsedPayload, tmpPayload1, tmpPayload2 map[string]*json.RawMessage
	var ref, branch, author, commit string
	json.Unmarshal(payload, &parsedPayload)
	json.Unmarshal(*tmpPayload1["push"], &tmpPayload2)
	json.Unmarshal(*tmpPayload1["actor"], &tmpPayload1)
	json.Unmarshal(*tmpPayload1["username"], &author)
	json.Unmarshal(*tmpPayload2["changes"], &tmpPayload2)
	json.Unmarshal(*tmpPayload2[0], &tmpPayload2)
	json.Unmarshal(*tmpPayload2["new"], &tmpPayload2)
	json.Unmarshal(*tmpPayload2["name"], &branch)
	json.Unmarshal(*tmpPayload2["target"], &tmpPayload2)
	json.Unmarshal(*tmpPayload2["hash"], &commit)

	return BitbucketEvent{Author: author, Branch: branch, Commit: commit}
}

type GitlabEvent RepoEvent
