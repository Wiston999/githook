package githook

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestGithubEvent(t *testing.T) {
	payload, _ := ioutil.ReadFile("../payloads/github.com.json")
	request := httptest.NewRequest("post", "/test", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")

	event := NewGithubEvent(request)
	if event.Author != "Wiston999" {
		t.Error("event.Author must be wiston999, got", event.Author)
	}

	if event.Branch != "master" {
		t.Error("event.Branch must be master, got", event.Author)
	}

	if event.Commit != "eddf11a4056b1abc8002c005ddc0a20cd5f1038a" {
		t.Error("event.Commit must be eddf11a4056b1abc8002c005ddc0a20cd5f1038a, got", event.Author)
	}

	v := url.Values{}
	v.Add("payload", string(payload))
	request = httptest.NewRequest("post", "/test", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	event = NewGithubEvent(request)
	if event.Author != "Wiston999" {
		t.Error("event.Author must be wiston999, got", event.Author)
	}

	if event.Branch != "master" {
		t.Error("event.Branch must be master, got", event.Author)
	}

	if event.Commit != "eddf11a4056b1abc8002c005ddc0a20cd5f1038a" {
		t.Error("event.Commit must be eddf11a4056b1abc8002c005ddc0a20cd5f1038a, got", event.Author)
	}
}

func TestBitbucketEvent(t *testing.T) {
	payload, _ := os.Open("./payloads/bitbucket.org.json")
	request := httptest.NewRequest("post", "/test", payload)

	event := NewBitbucketEvent(request)
	fmt.Printf("%#v\n", event)
}
