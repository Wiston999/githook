package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CommandLog interface {
	// AppendResult appends a CommandResult to the underlying CommandLog storage
	AppendResult(result CommandResult) (success bool, err error)
	// GetResults returns at most the latest n CommandResult stored in the underlying storage
	// sorted from latest to older, if n < 0 it returns all the CommandResult stored
	GetResults(n int) (results []CommandResult, err error)
	// RotateResults rotates the older CommandResult stored in the underlying storage so
	// only n CommandResult are left in the underlying storage, it returns the number rotated results
	// i.e.: the deleted ones
	RotateResults(n int) (deleted int, err error)
	// Counts counts the number of CommandResult stored in the underlying storage
	Count() (count int, err error)
}

type MemoryCommandLog struct {
	CommandLog []CommandResult
}

func NewMemoryCommandLog() MemoryCommandLog {
	cmdLog := MemoryCommandLog{}
	return cmdLog
}

func (m *MemoryCommandLog) AppendResult(result CommandResult) (success bool, err error) {
	m.CommandLog = append(m.CommandLog, result)
	return true, nil
}

func (m *MemoryCommandLog) GetResults(n int) (results []CommandResult, err error) {
	if n < 0 {
		results = make([]CommandResult, len(m.CommandLog))
	} else {
		results = make([]CommandResult, n)
	}

	for i := 0; i < len(results); i = i + 1 {
		results[i] = m.CommandLog[len(m.CommandLog)-1-i]
	}

	return
}

func (m *MemoryCommandLog) RotateResults(n int) (deleted int, err error) {
	if n < 0 {
		return n, errors.New("Rotate value must be greater than 0")
	}
	var head []CommandResult
	head, m.CommandLog = m.CommandLog[:n], m.CommandLog[n:]
	return len(head), nil
}

func (m *MemoryCommandLog) Count() (count int, err error) {
	return len(m.CommandLog), nil
}

type DiskCommandLog struct {
	Location string
}

func NewDiskCommandLog(location string) DiskCommandLog {
	cmdLog := DiskCommandLog{Location: location}
	return cmdLog
}

func (d *DiskCommandLog) AppendResult(result CommandResult) (success bool, err error) {

	fileName, err := filepath.Abs(filepath.Join(d.Location, fmt.Sprintf("%d", time.Now().UnixNano())))
	if err != nil {
		return
	}

	f, err := os.Create(fileName)
	if err != nil {
		return
	}

	err = json.NewEncoder(f).Encode(result)
	return (err == nil), err
}

func (d *DiskCommandLog) GetResults(n int) (results []CommandResult, err error) {
	files, err := filepath.Glob(filepath.Join(d.Location, "*"))
	if err != nil {
		return
	}

	var filesInt []int
	for _, f := range files {
		sf := strings.Split(f, "/")
		sfi, _ := strconv.Atoi(sf[len(sf)-1])
		filesInt = append(filesInt, sfi)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(filesInt)))

	if n < 0 {
		n = len(files)
	}
	for _, fileName := range filesInt[:n] {
		filePath, localErr := filepath.Abs(filepath.Join(d.Location, fmt.Sprintf("%d", fileName)))
		if localErr != nil {
			return results, localErr
		}
		file, localErr := os.Open(filePath)
		if localErr != nil {
			return results, localErr
		}

		var cmdResult CommandResult
		err = json.NewDecoder(file).Decode(&cmdResult)
		if err != nil {
			return
		}
		results = append(results, cmdResult)
	}
	return
}

func (d *DiskCommandLog) RotateResults(n int) (deleted int, err error) {
	if n < 0 {
		return n, errors.New("Rotate value must be greater than 0")
	}
	files, err := filepath.Glob(filepath.Join(d.Location, "*"))
	if err != nil {
		return
	}

	var filesInt []int
	for _, f := range files {
		sf := strings.Split(f, "/")
		sfi, _ := strconv.Atoi(sf[len(sf)-1])
		filesInt = append(filesInt, sfi)
	}
	sort.Ints(filesInt)

	for _, fileName := range filesInt[:n] {
		filePath, localErr := filepath.Abs(filepath.Join(d.Location, fmt.Sprintf("%d", fileName)))
		if localErr != nil {
			return deleted, localErr
		}

		err = os.Remove(filePath)
		if err != nil {
			return
		}
		deleted += 1
	}
	return
}

func (d *DiskCommandLog) Count() (count int, err error) {
	files, err := filepath.Glob(filepath.Join(d.Location, "*"))
	if err != nil {
		return
	}

	return len(files), nil
}
