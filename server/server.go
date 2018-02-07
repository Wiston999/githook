package server

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Server an http.Server all the needed information for starting and running the http server
type Server struct {
	*http.Server
	TLSCert           string
	TLSKey            string
	CmdLogDir         string
	CmdLogLimit       int
	WorkerChannelSize int
	Hooks             map[string]Hook
	MuxHandler        *http.ServeMux
	HooksHandled      map[string]int
	WorkerChannels    map[string]chan CommandJob
	CmdLog            CommandLog
}

// ListenAndServe set ups everything needed for the server to run and
// calls underlying http.Server ListenAndServer depending on
// Server is set up to use TLS or not
func (s *Server) ListenAndServe() (err error) {
	if s.MuxHandler == nil {
		s.MuxHandler = http.NewServeMux()
	}
	if err = s.setCommandLog(); err != nil {
		return
	}
	if s.WorkerChannels == nil {
		s.WorkerChannels = make(map[string]chan CommandJob)
	}
	if err = s.setHooks(); err != nil {
		return
	}
	if err = s.setAdminEndpoints(); err != nil {
		return
	}
	s.Server.Handler = s.MuxHandler
	if s.TLSCert != "" && s.TLSKey != "" {
		return s.Server.ListenAndServeTLS(s.TLSCert, s.TLSKey)
	}
	return s.Server.ListenAndServe()
}

// Stop tries to gracefully stop the http.Server finishing all pending tasks
// and closing underlying channels
func (s *Server) Stop() (err error) {
	return
}

// setCommandLog sets and configures the internal CommandLog
func (s *Server) setCommandLog() (err error) {
	s.CmdLog = NewMemoryCommandLog(s.CmdLogLimit)
	defer func() {
		switch s.CmdLog.(type) {
		case *MemoryCommandLog:
			log.Warn("CommandLogDir setting not found or invalid, using in memory command log")
		case *DiskCommandLog:
			log.Info("Commands will be logged to", s.CmdLogDir)
		}
	}()

	if s.CmdLogDir == "" {
		return
	}

	absLogDir, absErr := filepath.Abs(s.CmdLogDir)
	if absErr != nil {
		return
	}
	fileMode, statErr := os.Stat(absLogDir)
	if statErr == nil && fileMode.IsDir() {
		s.CmdLog = NewDiskCommandLog(s.CmdLogDir, s.CmdLogLimit)
	}
	return
}

func (s *Server) setAdminEndpoints() (err error) {
	s.MuxHandler.HandleFunc("/admin/hello", JSONRequestMiddleware(HelloHandler))
	s.MuxHandler.HandleFunc("/admin/cmdlog", JSONRequestMiddleware(CommandLogRESTHandler(s.CmdLog)))
	return
}

// setHooks configures hook handlers into an http.ServeMux handler given a map of hooks
func (s *Server) setHooks() (err error) {
	s.HooksHandled = make(map[string]int)
	for k, v := range s.Hooks {
		log.WithFields(log.Fields{
			"name": k,
			"hook": v,
		}).Info("Read hook")
		if _, exists := s.HooksHandled[v.Path]; exists {
			log.WithFields(log.Fields{"hook": k}).Warn("Path ", v.Path, " already defined, ignoring...")
			continue
		}
		if v.Type != "bitbucket" && v.Type != "github" && v.Type != "gitlab" {
			log.WithFields(log.Fields{"hook": k}).Warn("Unknown repository type, it must be one of: bitbucket, github or gitlab")
			continue
		}
		if !strings.HasPrefix(v.Path, "/") || strings.HasPrefix(v.Path, "/admin") {
			log.WithFields(log.Fields{"hook": k}).Warn("Path must start with / and must not start with /admin")
			continue
		}
		if v.Timeout <= 0 {
			log.WithFields(log.Fields{"hook": k}).Warn("Timeout must be greater than 0, got ", v.Timeout)
			continue
		}
		if len(v.Cmd) == 0 {
			log.WithFields(log.Fields{"hook": k}).Warn("Cmd must be defined")
			continue
		}
		if v.Concurrency < 0 {
			log.WithFields(log.Fields{"hook": k}).Warn("Concurrency level must be a value greater than 0")
			continue
		} else if v.Concurrency == 0 {
			log.WithFields(log.Fields{"hook": k}).Warn("Concurrency level of 0 found, falling back to default 1")
			v.Concurrency = 1
		}

		s.WorkerChannels[k] = make(chan CommandJob, s.WorkerChannelSize)
		s.MuxHandler.HandleFunc(v.Path, JSONRequestMiddleware(RepoRequestHandler(s.CmdLog, s.WorkerChannels[k], k, v)))
		for i := 0; i < v.Concurrency; i++ {
			go CommandWorker(k, s.WorkerChannels[k], s.CmdLog)
		}
		log.WithFields(log.Fields{
			"count": v.Concurrency,
			"hook":  k,
		}).Info("Started command workers")

		s.HooksHandled[v.Path] = 1
	}

	log.WithFields(log.Fields{"hooks": s.HooksHandled}).Debug("Hooks parsed from configuration file")
	if len(s.HooksHandled) == 0 {
		err = errors.New("No hooks parsed")
	}
	return
}
