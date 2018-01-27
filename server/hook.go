package server

// Hook structure holds all the information needed to configure an HTTP endpoint
// and execute the custom command on the system
// Type refers to the repository provider, it can be github, bitbucket or gitlab
// Path refers to the HTTP path where the HTTP server will be listening. I.e.: /mycustompayloadlistener
// Timeout specifies the number of seconds to wait for the custom command to be completed before
// killing it
// Cmd is the custom command to be executed each time an HTTP request is received
// It is a list of strings and each element will be an argument to exec.Command
// this implies that any kind of redirection using a shell won't work and will be treated as
// another parameter to exec function. I.e.:
// `Cmd: ["echo", "I want this in STDERR", "1>&2"]`
// will actually print "I want this in STDERR 1>&2" and won't print `"I want this in STDERR"` to `/dev/stderr`
// While this is less flexible from an UNIX shell perspective, it makes easier to run in different OS
// that don't has the same UNIX shell behaviour.
// It is also safer as remote data from git webhook provider won't be treated as part of a shell command
// but will be treated as part of a shell-command parameter
// Concurrency determines the number of concurrent workers that will be available to run command
// a concurrency level of 1 means that only 1 command can be executed at a time (mutex mode), default is 1
type Hook struct {
	Type        string
	Path        string
	Timeout     int
	Cmd         []string
	Concurrency int
}
