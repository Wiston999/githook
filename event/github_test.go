package event

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGithubEventOK(t *testing.T) {
	payload, _ := ioutil.ReadFile("../payloads/github.com.json")
	request := httptest.NewRequest("POST", "/test", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")

	event, err := NewGithubEvent(request)
	if err != nil {
		t.Error("NewGithubEvent should not return err != nil")
	}
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
	request = httptest.NewRequest("POST", "/test", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	event, err = NewGithubEvent(request)
	if err != nil {
		t.Error("NewGithubEvent should not return err != nil")
	}
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

func TestGithubEventKO(t *testing.T) {
	request := httptest.NewRequest("POST", "/test", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	_, err := NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with payload = \"\"")
	}

	request = httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with payload = \"{}\"")
	}

	request = httptest.NewRequest("GET", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail request method != POST")
	}

	v := url.Values{}
	v.Add("payload", "")
	request = httptest.NewRequest("POST", "/test", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with payload body = \"\"")
	}

	v.Add("payload", "{}")
	request = httptest.NewRequest("POST", "/test", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with payload body = \"{}\"")
	}

	v.Add("payload", "")
	request = httptest.NewRequest("GET", "/test", strings.NewReader(v.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with request method != POST")
	}
}
