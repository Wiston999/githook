# githook
> Execute arbitrary shell commands triggered by GIT webhooks

[![Build Status](https://travis-ci.org/Wiston999/githook.svg?branch=master)](https://travis-ci.org/Wiston999/githook?branch=master)
[![Coverage Status](https://coveralls.io/repos/github/Wiston999/githook/badge.svg)](https://coveralls.io/github/Wiston999/githook)


## Installation

Go to [releases](https://github.com/Wiston999/githook/releases) page and download the latest release for your platform

## Usage example

Right now, everything is setup using a configuration file instead of command argument flags:

```sh
$ githook -h
Usage of githook:
  -config string
    	Configuration file
```

Configuration file syntax is as follows:

```yaml
---
  address: (bind address for HTTP interface, default 0.0.0.0)
  port: (listening port for HTTP interface, default 65000)
  hooks:
    [hook name]
      type: {github, bitbucket, gitlab}
      path: (HTTP path where this hook will be triggered, i.e.: /webhook-payload)
      timeout: (Timeout in seconds before the command execution is treated as failed, required)
      cmd: [Array of strings, the command will be executed using https://golang.org/pkg/os/exec/#Command]
```

Configuration file example:

```yaml
---
  address: 127.0.0.1 # Bind to localhost (using a reverse proxy such as nginx)
  port: 8080
  hooks:
    github_custom_command:
      type: github # Webhook received from a GitHub repository
      path: /github-custom-command
      timeout: 300 # Wait for 300 seconds
      cmd: 
      - bash
      - /path/to/my/custom/script.sh
      - '--branch'
      - '{{.Branch}}'
      - '--author'
    bitbucket_mkdir:
      type: bitbucket # Webhook received from a Bitbucket repository
      path: /bitbucket-mkdir
      timeout: 30 # Wait for 30 seconds
      cmd: [mkdir, '-p', '{{.Author}}/{{.Branch}}/{{.Commit}}'] # Create a folder structure based on commit author, branch and hash
```
 Configuration file can be placed everywhere and be readable by the githook binary. Commands are executed with the same user and group as the githook binary runs.

#### A note on cmd syntax

* Each element of the cmd array must be [golang template](https://golang.org/pkg/text/template/) compliant. Current supported interpolation variables are:
  * Branch
  * Commit
  * Author
* Using array syntax over a single string was decided due to:
  * There is no chance to shell-injection attacks as each element in the list (unless first one) is treated as an argument and so, special shell characters like `;})$&` are treated as simple strings and has not special meaning.
  * Implements a common interface for \*NIX and non-\*NIX systems. This implies an easier implementation as the user is responsible to properly define the command.
* This decision has some caveats like:
  * Due to previous point, there is no way to redirect `cmd` output.
  * There is no way to build complex commands using shell pipelines.
  
## Development setup

If you wish to develop, you will need to have [Go](https://golang.org/) installed and setup in your system. Once Go is setup, clone the forked repository at `$GOPATH/src/github.com/Wiston999/githook`. This will avoid issues with subpackages.

## Contributing

1. Fork it (<https://github.com/Wiston999/githook/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

## Release History

* 0.1.0
  * First release
