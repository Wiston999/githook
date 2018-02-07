package server

import (
	"net/http"
	"testing"
)

func TestListenAndServe(t *testing.T) {
	s := &Server{}
	s.MuxHandler = http.NewServeMux()
	s.CmdLog = NewMemoryCommandLog(10)
	s.WorkerChannels = make(map[string]chan CommandJob)

	err := s.Stop()
	if err != nil {
		t.Errorf("Stop should never fail: %s", err)
	}
}

func TestSetHooks(t *testing.T) {
	s := Server{}
	s.MuxHandler = http.NewServeMux()
	s.Hooks = make(map[string]Hook)
	err := s.setHooks()
	if err == nil {
		t.Errorf("setHooks must error when no hooks are configured")
	}

	hooks := make(map[string]Hook)
	hooks["test1"] = Hook{Type: "github", Path: "/github1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test2"] = Hook{Type: "github", Path: "/github2", Cmd: []string{"true"}, Timeout: 500}
	hooks["test3"] = Hook{Type: "github", Path: "/github2", Cmd: []string{"true"}, Timeout: 500}
	hooks["test4"] = Hook{Type: "bitbucket", Path: "/bitbucket1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test5"] = Hook{Type: "gitlab", Path: "/gitlab1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test6"] = Hook{Type: "gitlab", Path: "invalid", Cmd: []string{"true"}, Timeout: 500}
	hooks["test7"] = Hook{Type: "gitlab", Path: "/admin/hello", Cmd: []string{"true"}, Timeout: 500}
	hooks["test8"] = Hook{Type: "invalid", Path: "/invalid1", Cmd: []string{"true"}, Timeout: 500}
	hooks["test9"] = Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{}, Timeout: 500}
	hooks["test10"] = Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{"true"}, Timeout: 0}
	hooks["test11"] = Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{"true"}, Timeout: -10}
	hooks["test12"] = Hook{Type: "bitbucket", Path: "/invalid2", Cmd: []string{"true"}, Timeout: 10, Concurrency: -10}
	hooks["test13"] = Hook{Type: "bitbucket", Path: "/admin", Cmd: []string{"true"}, Timeout: 500}

	s.Hooks = hooks
	s.WorkerChannels = make(map[string]chan CommandJob)
	err = s.setHooks()

	if err != nil {
		t.Errorf("Unknown error: %s", err)
	}

	removed := map[string]string{
		"test3":  "Duplicated Path",
		"test6":  "Invalid path (must start with /)",
		"test7":  "/admin/hello is a reserved path",
		"test8":  "Invalid type",
		"test9":  "Invalid Cmd (must be present)",
		"test10": "Timeout must be greater than 0",
		"test11": "Timeout must be greater than 0",
		"test12": "Concurrency must be greater than 0",
		"test13": "/admin is a reserved path",
	}

	hooksHandled := s.HooksHandled
	if len(hooksHandled) != (len(hooks) - len(removed)) {
		t.Errorf("Only %d hooks should have been added, got %d", len(hooks)-len(removed), len(hooksHandled))
	}

	if len(hooksHandled) != len(s.WorkerChannels) {
		t.Errorf("Number of hooks handled must match number of worker Channels")
	}
	for k, v := range removed {
		if _, found := hooksHandled[k]; found {
			t.Errorf("Case %s should have been removed due to: %s", k, v)
		}
	}
}

func TestStop(t *testing.T) {
	s := &Server{}
	s.MuxHandler = http.NewServeMux()
	s.CmdLog = NewMemoryCommandLog(10)
	s.WorkerChannels = make(map[string]chan CommandJob)

	err := s.Stop()
	if err != nil {
		t.Errorf("Stop should never fail: %s", err)
	}
}

func TestSetAdminEndpoints(t *testing.T) {
	s := &Server{}
	s.MuxHandler = http.NewServeMux()
	s.CmdLog = NewMemoryCommandLog(10)
	s.WorkerChannels = make(map[string]chan CommandJob)

	err := s.setAdminEndpoints()
	if err != nil {
		t.Errorf("setAdminEndpoints should never fail: %s", err)
	}
}

func TestSetCommandLog(t *testing.T) {
	s := &Server{}
	s.CmdLogDir = ""
	s.setCommandLog()
	switch v := s.CmdLog.(type) {
	case *MemoryCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected MemoryCommandLog", v)
	}

	s.CmdLogDir = "/notfound"
	s.setCommandLog()
	switch v := s.CmdLog.(type) {
	case *MemoryCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected MemoryCommandLog", v)
	}

	s.CmdLogDir = "./"
	s.setCommandLog()
	switch v := s.CmdLog.(type) {
	case *DiskCommandLog:
	default:
		t.Errorf("Command Log type is not the expected, got %#v but expected DiskCommandLog", v)
	}
}
