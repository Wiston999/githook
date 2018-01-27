package server

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Wiston999/githook/event"
)

func TestTranslateParams(t *testing.T) {
	event := event.RepoEvent{Branch: "my-branch", Author: "my-self", Commit: "0123456789abcdef"}

	testCases := []struct {
		tCase    []string
		expected []string
		err      bool
	}{
		{
			[]string{""},
			[]string{""},
			false,
		},
		{
			[]string{"this", "must", "remain", "unchanged"},
			[]string{"this", "must", "remain", "unchanged"},
			false,
		},
		{
			[]string{"The branch is {{.Branch}}"},
			[]string{"The branch is my-branch"},
			false,
		},
		{
			[]string{"The author is {{.Author}}"},
			[]string{"The author is my-self"},
			false,
		},
		{
			[]string{"The commit is {{.Commit}}"},
			[]string{"The commit is 0123456789abcdef"},
			false,
		},
		{
			[]string{"{{.Branch}}", "{{.Author}}", "{{.Commit}}"},
			[]string{"my-branch", "my-self", "0123456789abcdef"},
			false,
		},
		{
			[]string{"{{.Author}}", "{{.Branch}}", "{{.Commit}}"},
			[]string{"my-self", "my-branch", "0123456789abcdef"},
			false,
		},
		{
			[]string{"{{.Commit}}", "{{.Branch}}", "{{.Author}}"},
			[]string{"0123456789abcdef", "my-branch", "my-self"},
			false,
		},
		{
			[]string{"{{.Unknown}}"},
			[]string{},
			true,
		},
		{
			[]string{"{{.Unknown"},
			[]string{""},
			true,
		},
	}

	for i, test := range testCases {
		got, err := TranslateParams(test.tCase, event)
		if err != nil && !test.err {
			t.Errorf("%02d. TranslateParams should not throw error with %v but got %s", i, test.tCase, err)
		}

		if strings.Join(got, "") != strings.Join(test.expected, "") {
			t.Errorf("%02d. TranslateParams error, expected %v got %v", i, test.expected, got)
		}
	}
}

func TestRunCommand(t *testing.T) {
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

	for i, test := range testCases {
		got := RunCommand(test.cmd, test.timeout)

		if got.Err != nil && !test.expectedErr {
			t.Errorf("%02d. RunCommand should not throw error with %v but got %s", i, test.cmd, got.Err)
		}
		if match, _ := regexp.Match(test.expectedStdout, got.Stdout); !match {
			t.Errorf("%02d. RunCommand STDOUT does not match, expected %v but got %s", i, test.expectedStdout, got.Stdout)
		}
		if match, _ := regexp.Match(test.expectedStderr, got.Stderr); !match {
			t.Errorf("%02d. RunCommand STDERR does not match, expected %v but got %s", i, test.expectedStderr, got.Stderr)
		}
	}
}
