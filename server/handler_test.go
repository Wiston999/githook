package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/wiston999/githook/event"
)

func TestHello(t *testing.T) {
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HelloHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var jsonBody Response
	err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
	if err != nil {
		t.Errorf("Unable to decode JSON body into a Response: %s", err)
	}

	if jsonBody.Status != 200 {
		t.Errorf("Status field in response should have 200, got %v", jsonBody.Status)
	}

	if jsonBody.Msg == "" {
		t.Errorf("Msg field in response should not be empty")
	}
}

func TestJSONRequestMiddleware(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := JSONRequestMiddleware(http.HandlerFunc(HelloHandler))

	handler.ServeHTTP(rr, req)

	if count := strings.Count(buf.String(), "\n"); count != 2 {
		t.Errorf("JSONRequestMiddleware must produce 2 log lines, produced %v", count)
	}

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("JSONRequestMiddleware must produce JSON response with Header Content-Type: application/json")
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var jsonBody Response
	err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
	if err != nil {
		t.Errorf("Unable to decode JSON body into a Response: %s", err)
	}

	if jsonBody.Status != 200 {
		t.Errorf("Status field in response should have 200, got %v", jsonBody.Status)
	}

	if jsonBody.Msg == "" {
		t.Errorf("Msg field in response should not be empty")
	}
}

func TestRepoRequestHandler(t *testing.T) {
	testCases := []string{"github", "bitbucket", "unknown-type"}

	for _, test := range testCases {
		hook := event.Hook{
			Type: test,
			Cmd:  "echo {{branch}}",
			Path: "/payloadtest",
		}

		fmt.Printf("%#v\n", hook)
		req, err := http.NewRequest("POST", "/payloadtest", strings.NewReader(""))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RepoRequestHandler("test", hook))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusInternalServerError)
		}

		var jsonBody Response
		err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
		if err != nil {
			t.Errorf("Unable to decode JSON body into a Response: %s", err)
		}

		if jsonBody.Status != 500 {
			t.Errorf("Status field in response should have 500, got %v", jsonBody.Status)
		}

		if jsonBody.Msg == "" {
			t.Errorf("Msg field in response should not be empty")
		}
	}
}
