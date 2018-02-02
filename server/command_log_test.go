package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestAppendResult(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)

	testRounds := 10
	memoryCmdLog := NewMemoryCommandLog(testRounds)
	diskCmdLog := NewDiskCommandLog(tmpDir, testRounds)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		deleted, err := memoryCmdLog.AppendResult(cmdResult)
		if deleted != 0 || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got success == %v and err == %v", deleted, err)
		}
		deleted, err = diskCmdLog.AppendResult(cmdResult)
		if deleted != 0 || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got success == %v and err == %v", deleted, err)
		}
	}

	if len(memoryCmdLog.CommandLog) != testRounds {
		t.Errorf("[MemoryCommandLog] After %d AppendResult, log should contain 10 entries, got %d", testRounds, len(memoryCmdLog.CommandLog))
	}

	diskLogFiles, _ := filepath.Glob(filepath.Join(tmpDir, "*"))
	if len(diskLogFiles) != testRounds {
		t.Errorf("[DiskCommandLog] After %d AppendResult, log should contain 10 entries, got %d", testRounds, len(memoryCmdLog.CommandLog))
	}

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		deleted, err := memoryCmdLog.AppendResult(cmdResult)
		if deleted != 1 || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should rotate element when 'full', deleted: %d", deleted)
		}
		deleted, err = diskCmdLog.AppendResult(cmdResult)
		if deleted != 1 || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should rotate element when 'full', deleted: %d", deleted)
		}
	}
	memoryCmdLog.MaxCommands = 0
	diskCmdLog.MaxCommands = 0
	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		deleted, err := memoryCmdLog.AppendResult(cmdResult)
		if deleted != 0 || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not rotate element when 'MaxCommands == 0', deleted: %d", deleted)
		}
		deleted, err = diskCmdLog.AppendResult(cmdResult)
		if deleted != 0 || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not rotate element when 'MaxCommands == 0', delted: %d", deleted)
		}
	}
}

func TestGetResults(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)

	testRounds := 10
	memoryCmdLog := NewMemoryCommandLog(100)
	diskCmdLog := NewDiskCommandLog(tmpDir, 100)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		_, err := memoryCmdLog.AppendResult(cmdResult)
		if err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got err == %v", err)
		}
		_, err = diskCmdLog.AppendResult(cmdResult)
		if err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got err == %v", err)
		}
	}

	testCases := []struct {
		get, expected int
	}{{
		get:      -1,
		expected: testRounds,
	}, {
		get:      8,
		expected: 8,
	}, {
		get:      2,
		expected: 2,
	}, {
		get:      1,
		expected: 1,
	}, {
		get:      5,
		expected: 5,
	}}

	for _, tCase := range testCases {
		memoryElements, _ := memoryCmdLog.GetResults(tCase.get)

		if len(memoryElements) != tCase.expected {
			t.Errorf(
				"[MemoryCommandLog] Expected length returned with %d parameter to be %d, got %d",
				tCase.get,
				tCase.expected,
				len(memoryElements),
			)
		}

		for i, cmdResult := range memoryElements {
			if bytes.Compare(cmdResult.Stderr, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
				t.Errorf("[MemoryCommandLog] Expected Stderr content of %d element to be %d, got %s", i, testRounds-1-i, cmdResult.Stderr)
			}

			if bytes.Compare(cmdResult.Stdout, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
				t.Errorf("[MemoryCommandLog] Expected Stdout content of %d element to be %d, got %s", i, testRounds-1-i, cmdResult.Stdout)
			}
		}

		diskFiles, err := diskCmdLog.GetResults(tCase.get)
		if err != nil {
			t.Errorf("Unable to get diskCmdLog: %v", err)
		}

		if len(diskFiles) != tCase.expected {
			t.Errorf(
				"[DiskCommandLog] Expected length returned with %d parameter to be %d, got %d",
				tCase.get,
				tCase.expected,
				len(diskFiles),
			)
		}

		for i, cmdResult := range diskFiles {
			if bytes.Compare(cmdResult.Stderr, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
				t.Errorf("[DiskCommandLog] Expected Stderr content of %d element to be %d, got %s", i, testRounds-1-i, cmdResult.Stderr)
			}

			if bytes.Compare(cmdResult.Stdout, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
				t.Errorf("[DiskCommandLog] Expected Stdout content of %d element to be %d, got %s", i, testRounds-1-i, cmdResult.Stdout)
			}
		}
	}
}

func TestRotateResults(t *testing.T) {
	testRounds := 10
	testCases := []struct {
		rotate, expected int
		err              bool
	}{{
		rotate:   -1,
		expected: -1,
		err:      true,
	}, {
		rotate:   0,
		expected: testRounds,
		err:      false,
	}, {
		rotate:   8,
		expected: testRounds - 8,
		err:      false,
	}, {
		rotate:   2,
		expected: testRounds - 2,
		err:      false,
	}, {
		rotate:   1,
		expected: testRounds - 1,
		err:      false,
	}, {
		rotate:   5,
		expected: testRounds - 5,
		err:      false,
	}, {
		rotate:   testRounds,
		expected: 0,
		err:      false,
	}}

	for j, tCase := range testCases {
		tmpDir, _ := ioutil.TempDir("", "")
		memoryCmdLog := NewMemoryCommandLog(1000)
		diskCmdLog := NewDiskCommandLog(tmpDir, 1000)

		for i := 0; i < testRounds; i++ {
			cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
			_, err := memoryCmdLog.AppendResult(cmdResult)
			if err != nil {
				t.Errorf("#%02d. [MemoryCommandLog] AppendResult should not fail, got err == %v", j, err)
			}
			_, err = diskCmdLog.AppendResult(cmdResult)
			if err != nil {
				t.Errorf("#%02d. [DiskCommandLog] AppendResult should not fail, got err == %v", j, err)
			}
		}

		memoryCmdLog.MaxCommands = tCase.rotate
		rotateReturn, err := memoryCmdLog.RotateResults()
		if tCase.err && err == nil {
			t.Errorf("Expected error but not returned")
		} else if !tCase.err && err != nil {
			t.Errorf("Unable to rotate memoryCmdLog: %v", err)
		} else if !tCase.err {

			memoryElements, err := memoryCmdLog.GetResults(-1)

			if err != nil {
				t.Errorf("Unable to get memoryCmdLog: %v", err)
			}
			fmt.Printf("#%02d. %#v\n", j, memoryCmdLog.CommandLog)
			if rotateReturn != tCase.expected {
				t.Errorf(
					"#%02d. [MemoryCommandLog] Expected rotate returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.expected,
					rotateReturn,
				)
			}

			if len(memoryElements) != tCase.rotate {
				t.Errorf(
					"#%02d. [MemoryCommandLog] Expected length returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.rotate,
					len(memoryElements),
				)
			}

			for i, cmdResult := range memoryElements {
				if bytes.Compare(cmdResult.Stderr, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
					t.Errorf("#%02d. [MemoryCommandLog] Expected Stderr content of %d element to be %d, got %s", j, i, testRounds-1-i, cmdResult.Stderr)
				}

				if bytes.Compare(cmdResult.Stdout, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
					t.Errorf("#%02d. [MemoryCommandLog] Expected Stdout content of %d element to be %d, got %s", j, i, testRounds-1-i, cmdResult.Stdout)
				}
			}
		}

		diskCmdLog.MaxCommands = tCase.rotate
		rotateReturn, err = diskCmdLog.RotateResults()
		if tCase.err && err == nil {
			t.Errorf("Expected error but not returned")
		} else if !tCase.err && err != nil {
			t.Errorf("Unable to rotate diskCmdLog: %v", err)
		} else if !tCase.err {

			diskFiles, err := diskCmdLog.GetResults(-1)
			if err != nil {
				t.Errorf("Unable to get diskCmdLog: %v", err)
			}

			if rotateReturn != tCase.expected {
				t.Errorf(
					"#%02d. [DiskCommandLog] Expected rotate returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.expected,
					rotateReturn,
				)
			}

			if len(diskFiles) != tCase.rotate {
				t.Errorf(
					"#%02d. [DiskCommandLog] Expected lenght returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.rotate,
					len(diskFiles),
				)
			}

			for i, cmdResult := range diskFiles {
				if bytes.Compare(cmdResult.Stderr, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
					t.Errorf("#%02d. [DiskCommandLog] Expected Stderr content of %d element to be %d, got %s", j, i, testRounds-1-i, cmdResult.Stderr)
				}

				if bytes.Compare(cmdResult.Stdout, []byte(strconv.Itoa(testRounds-1-i))) != 0 {
					t.Errorf("#%02d. [DiskCommandLog] Expected Stdout content of %d element to be %d, got %s", j, i, testRounds-1-i, cmdResult.Stdout)
				}
			}
		}
		os.RemoveAll(tmpDir)
	}
}

func TestCount(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)

	testRounds := 10
	memoryCmdLog := NewMemoryCommandLog(100)
	diskCmdLog := NewDiskCommandLog(tmpDir, 100)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		_, err := memoryCmdLog.AppendResult(cmdResult)
		if err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got err == %v", err)
		}
		_, err = diskCmdLog.AppendResult(cmdResult)
		if err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got err == %v", err)
		}
	}

	if c, err := memoryCmdLog.Count(); c != testRounds || err != nil {
		t.Errorf("[MemoryCommandLog] Count should return %d and error should be nil, got %d %v", testRounds, c, err)
	}

	if c, err := diskCmdLog.Count(); c != testRounds || err != nil {
		t.Errorf("[DiskCommandLog] Count should return %d and error should be nil, got %d %v", testRounds, c, err)
	}
}
