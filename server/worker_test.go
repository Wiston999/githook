package server

import (
	"testing"
)

func TestCommandWorker(t *testing.T) {
	testCases := []struct {
		cmd            []string
		timeout        int
		expectedStdout string
		expectedStderr string
		expectedErr    bool
	}{
		{
			[]string{},
			10,
			"",
			"",
			true,
		},
		{
			[]string{"echo", "-n", "HELLO FROM ECHO STDOUT"},
			10,
			"^HELLO FROM ECHO STDOUT$",
			"",
			false,
		},
		{
			[]string{"echo", "HELLO FROM ECHO STDOUT"},
			10,
			"^HELLO FROM ECHO STDOUT\n$",
			"",
			false,
		},
		{
			[]string{"sleep", "1"},
			10,
			"",
			"",
			true,
		},
		{
			[]string{"ifthiscommandexistsiwillfail"},
			10,
			"",
			"",
			true,
		},
		{
			[]string{"logger", "-s", "HELLO FROM LOGGER STDERR"},
			10,
			"",
			"^.*?HELLO FROM LOGGER STDERR\n$",
			false,
		},
	}

	workChannel := make(chan CommandJob, 100)
	for i, test := range testCases {
		workChannel <- CommandJob{Cmd: test.cmd, ID: string(i), Timeout: test.timeout}
	}

	cmdLog := NewMemoryCommandLog(100)
	resultChannel := make(chan int)
	go func(resChan chan int) {
		resultChannel <- CommandWorker("CommandWorkerTest", workChannel, cmdLog)
	}(resultChannel)

	close(workChannel)

	works := <-resultChannel
	if works != len(testCases) {
		t.Errorf(
			"CommandWorker should have return the number of processed jobs, got %v expected %v",
			works,
			len(testCases),
		)
	}
}
