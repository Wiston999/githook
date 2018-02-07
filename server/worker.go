package server

import (
	log "github.com/sirupsen/logrus"
)

// CommandJob encodes a request to execute a command
type CommandJob struct {
	Cmd      []string
	ID       string
	Timeout  int
	Response chan CommandResult
}

func CommandWorker(id string, jobs <-chan CommandJob, cmdLog CommandLog) (executed int) {
	for job := range jobs {
		log.WithFields(log.Fields{
			"worker": id,
			"jobId":  job.ID,
			"cmd":    job.Cmd,
		}).Info("Executing command")
		cmdResult := RunCommand(job.Cmd, job.Timeout)
		log.Debug("Execution of ", job.Cmd, " finished ", cmdResult)
		if cmdResult.Err != nil {
			log.WithFields(log.Fields{
				"worker": id,
				"jobId":  job.ID,
				"err":    cmdResult.Err,
				"stderr": cmdResult.Stderr,
			}).Warn("Command finished unsuccessfully")
		} else {
			log.WithFields(log.Fields{
				"worker": id,
				"jobId":  job.ID,
				"err":    cmdResult.Err,
			}).Info("Command finished successfully")
		}
		cmdLog.AppendResult(cmdResult)
		executed++
		if job.Response != nil {
			job.Response <- cmdResult
		}
	}
	return
}
