package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Wiston999/githook/event"
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
	bbPayload, err := ioutil.ReadFile("../payloads/bitbucket.org.json")
	if err != nil {
		t.Fatal(err)
	}
	bbPayloadStr := string(bbPayload)

	ghPayload, err := ioutil.ReadFile("../payloads/github.com.json")
	if err != nil {
		t.Fatal(err)
	}
	ghPayloadStr := string(ghPayload)
	testCases := []struct {
		Query   string
		Type    string
		Payload io.Reader
		Sync    bool
		Err     bool
	}{
		{"bitbucket", "bitbucket", strings.NewReader(""), false, true},
		{"bitbucket", "bitbucket", nil, false, true},
		{"bitbucket", "bitbucket", strings.NewReader(ghPayloadStr), false, true},
		{"bitbucket", "bitbucket", strings.NewReader(bbPayloadStr), false, false},
		{"bitbucket?sync", "bitbucket", strings.NewReader(bbPayloadStr), true, false},
		{"github", "github", strings.NewReader(""), false, true},
		{"github", "github", nil, false, true},
		{"github", "github", strings.NewReader(bbPayloadStr), false, true},
		{"github", "github", strings.NewReader(ghPayloadStr), false, false},
		{"github?sync", "github", strings.NewReader(ghPayloadStr), true, false},
		{"unknown-type", "unknown", strings.NewReader(""), false, true},
	}

	for i, test := range testCases {
		hook := event.Hook{
			Type:    test.Type,
			Cmd:     []string{"echo", "{{.Branch}}"},
			Path:    "/payloadtest",
			Timeout: 10,
		}

		req, err := http.NewRequest("POST", test.Query, test.Payload)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RepoRequestHandler("test", hook))

		handler.ServeHTTP(rr, req)

		var jsonBody Response
		err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
		if err != nil {
			t.Errorf("%02d. Unable to decode JSON body into a Response: %s", i, err)
		}

		if test.Sync {
			if !strings.Contains(jsonBody.Msg, "with result") {
				t.Errorf("%02d. Msg field in response should contain with result string when sync execution, got %s", i, jsonBody.Msg)
			}
		}

		if jsonBody.Msg == "" {
			t.Errorf("%02d. Msg field in response should not be empty", i)
		}
		if test.Err {
			if status := rr.Code; status != http.StatusInternalServerError {
				t.Errorf("%02d. Handler returned wrong status code: got %v want %v",
					i,
					status, http.StatusInternalServerError)
			}
			if jsonBody.Status != 500 {
				t.Errorf("%02d. Status field in response should have 500, got %v", i, jsonBody.Status)
			}

		} else {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("%02d. Handler returned wrong status code: got %v want %v",
					i,
					status, http.StatusOK)
			}
			if jsonBody.Status != 200 {
				t.Errorf("%02d. Status field in response should have 200, got %v", i, jsonBody.Status)
			}
		}
	}
}
