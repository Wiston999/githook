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

// CommandLog is the interface that must be implemented by command loggers
type CommandLog interface {
	// AppendResult appends a CommandResult to the underlying CommandLog storage
	AppendResult(result CommandResult) (deleted int, err error)
	// GetResults returns at most the latest n CommandResult stored in the underlying storage
	// sorted from latest to older, if n < 0 it returns all the CommandResult stored
	GetResults(n int) (results []CommandResult, err error)
	// RotateResults rotates the older CommandResult stored in the underlying storage so
	// only MaxCommands CommandResult are left in the underlying storage, it returns the number rotated results
	// i.e.: the deleted ones
	RotateResults() (deleted int, err error)
	// Counts counts the number of CommandResult stored in the underlying storage
	Count() (count int, err error)
}

// MemoryCommandLog implements the CommandLog interface storing the results in memory
type MemoryCommandLog struct {
	MaxCommands int
	CommandLog  []CommandResult
}

// NewMemoryCommandLog creates and object of type MemoryCommandLog
func NewMemoryCommandLog(rotate int) *MemoryCommandLog {
	return &MemoryCommandLog{MaxCommands: rotate}
}

// AppendResult of MemoryCommandLog
func (m *MemoryCommandLog) AppendResult(result CommandResult) (deleted int, err error) {
	m.CommandLog = append(m.CommandLog, result)
	if m.MaxCommands > 0 {
		return m.RotateResults()
	}
	return 0, nil
}

// GetResults of MemoryCommandLog
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

// RotateResults of MemoryCommandLog
func (m *MemoryCommandLog) RotateResults() (deleted int, err error) {
	n := m.MaxCommands
	if n < 0 {
		return n, errors.New("Rotate value must be greater than 0")
	} else if len(m.CommandLog) < n {
		// Nothing to rotate
		return 0, nil
	}
	n = len(m.CommandLog) - n
	var head []CommandResult
	head, m.CommandLog = m.CommandLog[:n], m.CommandLog[n:]
	return len(head), nil
}

// Count of MemoryCommandLog
func (m *MemoryCommandLog) Count() (count int, err error) {
	return len(m.CommandLog), nil
}

// DiskCommandLog implements the CommandLog interface storing the results in disk
type DiskCommandLog struct {
	Location    string
	MaxCommands int
}

// NewDiskCommandLog creates and object of type DiskCommandLog
func NewDiskCommandLog(location string, rotate int) *DiskCommandLog {
	return &DiskCommandLog{Location: location, MaxCommands: rotate}
}

// AppendResult of DiskCommandLog
func (d *DiskCommandLog) AppendResult(result CommandResult) (deleted int, err error) {

	fileName, err := filepath.Abs(filepath.Join(d.Location, fmt.Sprintf("%d", time.Now().UnixNano())))
	if err != nil {
		return
	}

	f, err := os.Create(fileName)
	if err != nil {
		return
	}

	err = json.NewEncoder(f).Encode(result)

	if err == nil && d.MaxCommands > 0 {
		return d.RotateResults()
	}
	return 0, err
}

// GetResults of DiskCommandLog
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

// RotateResults of DiskCommandLog
func (d *DiskCommandLog) RotateResults() (deleted int, err error) {
	n := d.MaxCommands
	if n < 0 {
		return n, errors.New("Rotate value must be greater than 0")
	} else if c, _ := d.Count(); c < n {
		// Nothing to rotate
		return 0, nil
	}
	c, err := d.Count()
	if err != nil {
		return
	}

	n = c - n

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
		deleted++
	}
	return
}

// Count of DiskCommandLog
func (d *DiskCommandLog) Count() (count int, err error) {
	files, err := filepath.Glob(filepath.Join(d.Location, "*"))
	if err != nil {
		return
	}

	return len(files), nil
}
