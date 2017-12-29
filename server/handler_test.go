package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	testCases := []struct {
		Query string
		Sync  bool
	}{
		{"github", false},
		{"bitbucket", false},
		{"bitbucket", true},
		{"unknown-type", false},
	}

	for i, test := range testCases {
		hook := event.Hook{
			Type: test.Query,
			Cmd:  []string{"echo", "{{.Branch}}"},
			Path: "/payloadtest",
		}

		fmt.Printf("%#v\n", hook)
		var req *http.Request
		var err error
		if test.Sync {
			payload, _ := ioutil.ReadFile("../payloads/bitbucket.org.json")
			req, err = http.NewRequest("POST", "/payloadtest?sync", strings.NewReader(string(payload)))
		} else {
			req, err = http.NewRequest("POST", "/payloadtest", strings.NewReader(""))
		}
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RepoRequestHandler("test", hook))

		handler.ServeHTTP(rr, req)

		var jsonBody Response
		err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
		if test.Sync {
			if !strings.Contains(jsonBody.Msg, "with result") {
				t.Errorf("%02d. Msg field in response should contain with result string when sync execution, got %s", i, jsonBody.Msg)
			}
		} else {
			if status := rr.Code; status != http.StatusInternalServerError {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusInternalServerError)
			}

			if err != nil {
				t.Errorf("%02d. Unable to decode JSON body into a Response: %s", i, err)
			}

			if jsonBody.Status != 500 {
				t.Errorf("%02d. Status field in response should have 500, got %v", i, jsonBody.Status)
			}

			if jsonBody.Msg == "" {
				t.Errorf("%02d. Msg field in response should not be empty", i)
			}
		}
	}
}
