package event

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBitbucketEventOK(t *testing.T) {
	payload, _ := ioutil.ReadFile("../payloads/bitbucket.org.json")
	request := httptest.NewRequest("POST", "/test", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")

	event, err := NewBitbucketEvent(request)
	if err != nil {
		t.Error("NewBitbucketEvent should not return err != nil")
	}
	if event.Author != "vcabezas" {
		t.Error("event.Author must be vcabezas, got", event.Author)
	}

	if event.Branch != "master" {
		t.Error("event.Branch must be master, got", event.Author)
	}

	if event.Commit != "ffcc6f559d9be8124711e94349c4fe26642d762b" {
		t.Error("event.Commit must be ffcc6f559d9be8124711e94349c4fe26642d762b, got", event.Author)
	}
}

func TestBitbucketEventKO(t *testing.T) {
	request := httptest.NewRequest("POST", "/test", nil)
	request.Header.Set("Content-Type", "application/json")

	_, err := NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail with payload = nil")
	}

	request = httptest.NewRequest("GET", "/test", nil)
	request.Header.Set("Content-Type", "application/json")

	_, err = NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail with payload = nil")
	}

	request = httptest.NewRequest("POST", "/test", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail with payload = \"\"")
	}

	request = httptest.NewRequest("GET", "/test", strings.NewReader(""))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail with payload = \"\"")
	}

	request = httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail with payload = \"{}\"")
	}

	request = httptest.NewRequest("GET", "/test", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")

	_, err = NewBitbucketEvent(request)
	if err == nil {
		t.Error("NewBitbucketEvent should fail request method != POST")
	}

}
