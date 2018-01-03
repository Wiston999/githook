package event

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitlabEventOK(t *testing.T) {
	payload, _ := ioutil.ReadFile("../payloads/gitlab.com.json")
	request := httptest.NewRequest("POST", "/test", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")

	event, err := NewGitlabEvent(request)
	if err != nil {
		t.Error("NewGitlabEvent should not return err != nil")
	}
	if event.Author != "jsmith" {
		t.Error("event.Author must be jsmith, got", event.Author)
	}

	if event.Branch != "master" {
		t.Error("event.Branch must be master, got", event.Author)
	}

	if event.Commit != "da1560886d4f094c3e6c9ef40349f7d38b5d27d7" {
		t.Error("event.Commit must be da1560886d4f094c3e6c9ef40349f7d38b5d27d7, got", event.Author)
	}
}

func TestGitlabEventKO(t *testing.T) {
	request := httptest.NewRequest("POST", "/test", nil)
	request.Header.Set("Content-Type", "application/json")

	_, err := NewGitlabEvent(request)
	if err == nil {
		t.Error("NewGitlabEvent should fail with payload = nil")
	}

	request = httptest.NewRequest("GET", "/test", nil)
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGitlabEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail with payload = nil")
	}

	request = httptest.NewRequest("POST", "/test", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGitlabEvent(request)
	if err == nil {
		t.Error("NewGitlabEvent should fail with payload = \"\"")
	}

	request = httptest.NewRequest("GET", "/test", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGitlabEvent(request)
	if err == nil {
		t.Error("NewGitlabEvent should fail with payload = \"\"")
	}

	request = httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGitlabEvent(request)
	if err == nil {
		t.Error("NewGitlabEvent should fail with payload = \"{}\"")
	}

	request = httptest.NewRequest("GET", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewGithubEvent(request)
	if err == nil {
		t.Error("NewGithubEvent should fail request method != POST")
	}
}
