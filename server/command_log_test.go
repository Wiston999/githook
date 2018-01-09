package server

import (
	"bytes"
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
	memoryCmdLog := NewMemoryCommandLog()
	diskCmdLog := NewDiskCommandLog(tmpDir)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		success, err := memoryCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
		}
		success, err = diskCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
		}
	}

	if len(memoryCmdLog.CommandLog) != testRounds {
		t.Errorf("[MemoryCommandLog] After %d AppendResult, log should contain 10 entries, got %d", testRounds, len(memoryCmdLog.CommandLog))
	}

	diskLogFiles, _ := filepath.Glob(filepath.Join(tmpDir, "*"))
	if len(diskLogFiles) != testRounds {
		t.Errorf("[DiskCommandLog] After %d AppendResult, log should contain 10 entries, got %d", testRounds, len(memoryCmdLog.CommandLog))
	}

}

func TestGetResults(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)

	testRounds := 10
	memoryCmdLog := NewMemoryCommandLog()
	diskCmdLog := NewDiskCommandLog(tmpDir)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		success, err := memoryCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
		}
		success, err = diskCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
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
		expected: testRounds,
		err:      true,
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
	}}

	for j, tCase := range testCases {
		tmpDir, _ := ioutil.TempDir("", "")
		memoryCmdLog := NewMemoryCommandLog()
		diskCmdLog := NewDiskCommandLog(tmpDir)

		for i := 0; i < testRounds; i++ {
			cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
			success, err := memoryCmdLog.AppendResult(cmdResult)
			if success != true || err != nil {
				t.Errorf("#%02d. [MemoryCommandLog] AppendResult should not fail, got success == %v and err == %v", j, success, err)
			}
			success, err = diskCmdLog.AppendResult(cmdResult)
			if success != true || err != nil {
				t.Errorf("#%02d. [DiskCommandLog] AppendResult should not fail, got success == %v and err == %v", j, success, err)
			}
		}

		rotateReturn, err := memoryCmdLog.RotateResults(tCase.rotate)
		if tCase.err && err == nil {
			t.Errorf("Expected error but not returned")
		} else if !tCase.err && err != nil {
			t.Errorf("Unable to rotate memoryCmdLog: %v", err)
		} else {

			memoryElements, err := memoryCmdLog.GetResults(-1)

			if err != nil {
				t.Errorf("Unable to get memoryCmdLog: %v", err)
			}

			if rotateReturn != tCase.rotate {
				t.Errorf(
					"#%02d. [MemoryCommandLog] Expected rotate returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.rotate,
					rotateReturn,
				)
			}

			if len(memoryElements) != tCase.expected {
				t.Errorf(
					"#%02d. [MemoryCommandLog] Expected length returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.expected,
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

		rotateReturn, err = diskCmdLog.RotateResults(tCase.rotate)
		if tCase.err && err == nil {
			t.Errorf("Expected error but not returned")
		} else if !tCase.err && err != nil {
			t.Errorf("Unable to rotate diskCmdLog: %v", err)
		} else {

			diskFiles, err := diskCmdLog.GetResults(-1)
			if err != nil {
				t.Errorf("Unable to get diskCmdLog: %v", err)
			}

			if rotateReturn != tCase.rotate {
				t.Errorf(
					"#%02d. [DiskCommandLog] Expected rotate returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.rotate,
					rotateReturn,
				)
			}

			if len(diskFiles) != tCase.expected {
				t.Errorf(
					"#%02d. [DiskCommandLog] Expected lenght returned with %d parameter to be %d, got %d",
					j,
					tCase.rotate,
					tCase.expected,
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
	memoryCmdLog := NewMemoryCommandLog()
	diskCmdLog := NewDiskCommandLog(tmpDir)

	for i := 0; i < testRounds; i++ {
		cmdResult := CommandResult{Stdout: []byte(strconv.Itoa(i)), Stderr: []byte(strconv.Itoa(i))}
		success, err := memoryCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[MemoryCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
		}
		success, err = diskCmdLog.AppendResult(cmdResult)
		if success != true || err != nil {
			t.Errorf("[DiskCommandLog] AppendResult should not fail, got success == %v and err == %v", success, err)
		}
	}

	if c, err := memoryCmdLog.Count(); c != testRounds || err != nil {
		t.Errorf("[MemoryCommandLog] Count should return %d and error should be nil, got %d %v", testRounds, c, err)
	}

	if c, err := diskCmdLog.Count(); c != testRounds || err != nil {
		t.Errorf("[DiskCommandLog] Count should return %d and error should be nil, got %d %v", testRounds, c, err)
	}
}
