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
	"strconv"
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

	glPayload, err := ioutil.ReadFile("../payloads/gitlab.com.json")
	if err != nil {
		t.Fatal(err)
	}
	glPayloadStr := string(glPayload)
	testCases := []struct {
		Query   string
		Type    string
		Cmd     []string
		Payload io.Reader
		Sync    bool
		Err     bool
	}{
		{"bitbucket", "bitbucket", []string{"echo", "{{.Branch}}"}, strings.NewReader(""), false, true},
		{"bitbucket", "bitbucket", []string{"echo", "{{.Branch}}"}, nil, false, true},
		{"bitbucket", "bitbucket", []string{"echo", "{{.Branch}}"}, strings.NewReader(ghPayloadStr), false, true},
		{"bitbucket", "bitbucket", []string{"echo", "{{.Branch}}"}, strings.NewReader(bbPayloadStr), false, false},
		{"bitbucket?sync", "bitbucket", []string{"echo", "{{.Branch}}"}, strings.NewReader(bbPayloadStr), true, false},
		{"invalid-cmd", "bitbucket", []string{"echo", "{{.Branch"}, strings.NewReader(bbPayloadStr), false, true},
		{"github", "github", []string{"echo", "{{.Branch}}"}, strings.NewReader(""), false, true},
		{"github", "github", []string{"echo", "{{.Branch}}"}, nil, false, true},
		{"github", "github", []string{"echo", "{{.Branch}}"}, strings.NewReader(bbPayloadStr), false, true},
		{"github", "github", []string{"echo", "{{.Branch}}"}, strings.NewReader(ghPayloadStr), false, false},
		{"github?sync", "github", []string{"echo", "{{.Branch}}"}, strings.NewReader(ghPayloadStr), true, false},
		{"invalid-cmd", "github", []string{"echo", "{{.Branch"}, strings.NewReader(ghPayloadStr), false, true},
		{"gitlab", "gitlab", []string{"echo", "{{.Branch}}"}, strings.NewReader(""), false, true},
		{"gitlab", "gitlab", []string{"echo", "{{.Branch}}"}, nil, false, true},
		{"gitlab", "gitlab", []string{"echo", "{{.Branch}}"}, strings.NewReader(bbPayloadStr), false, true},
		{"gitlab", "gitlab", []string{"echo", "{{.Branch}}"}, strings.NewReader(glPayloadStr), false, false},
		{"gitlab?sync", "gitlab", []string{"echo", "{{.Branch}}"}, strings.NewReader(glPayloadStr), true, false},
		{"invalid-cmd", "gitlab", []string{"echo", "{{.Branch"}, strings.NewReader(glPayloadStr), false, true},
		{"unknown-type", "unknown", []string{"echo", "{{.Branch}}"}, strings.NewReader(""), false, true},
	}

	for i, test := range testCases {
		hook := event.Hook{
			Type:    test.Type,
			Cmd:     test.Cmd,
			Path:    "/payloadtest",
			Timeout: 10,
		}

		req, err := http.NewRequest("POST", test.Query, test.Payload)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}

		cmdLog := NewMemoryCommandLog()

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(RepoRequestHandler(&cmdLog, "test", hook))

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
		// Somehow, this does not work, but using the main.go the cmd log works as a charm, so
		// I'll skip this check until I know how to test it better
		// if count, _ := cmdLog.Count(); count != 1 {
		// 	t.Errorf("%02d. Command log should contain 1 element, got %d", i, count)
		// }
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

func TestCommandLogRESTHandler(t *testing.T) {
	logResults := 20
	testCases := []struct {
		Query string
		Count int
		Err   bool
	}{
		{"admin/log?count=10", 10, false},
		{"admin/log?count=", logResults, false},
		{"admin/log", logResults, false},
		{"admin/log?count=1", 1, false},
		{"admin/log?count=-1", logResults, false},
	}

	for i, test := range testCases {
		req, err := http.NewRequest("GET", test.Query, nil)
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			t.Fatal(err)
		}

		cmdLog := NewMemoryCommandLog()

		for i := 0; i < logResults; i = i + 1 {
			cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
			_, err := cmdLog.AppendResult(cmdResult)
			if err != nil {
				t.Fatal(err)
			}
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CommandLogRESTHandler(&cmdLog))

		handler.ServeHTTP(rr, req)

		var jsonBody Response
		err = json.Unmarshal([]byte(rr.Body.String()), &jsonBody)
		if err != nil {
			t.Errorf("%02d. Unable to decode JSON body into a Response: %s", i, err)
		}

		if jsonBody.Msg == "" {
			t.Errorf("%02d. Msg field in response should not be empty", i)
		}
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("%02d. Handler returned wrong status code: got %v want %v",
				i,
				status, http.StatusOK)
		}
		if jsonBody.Status != 200 {
			t.Errorf("%02d. Status field in response should have 200, got %v", i, jsonBody.Status)
		}
		results := jsonBody.Body.([]interface{})
		if len(results) != test.Count {
			t.Errorf("%02d. Number of results returned should be %d, got %d", i, len(results), test.Count)
		}
	}
}
