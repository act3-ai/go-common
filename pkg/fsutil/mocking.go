package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"syscall"
	"time"
)

type errorDirEntry struct {
	fs.DirEntry
}

func (e *errorDirEntry) Info() (fs.FileInfo, error) {
	return nil, errors.New("Info error")
}

type errorFile struct {
	fs.File
}

func (ef *errorFile) Stat() (fs.FileInfo, error) {
	return nil, fmt.Errorf("simulated error")
}

type customFS struct {
	fs.FS
	files map[string]customFileInfo
}

func newCustomFS(fs fs.FS, files map[string]customFileInfo) *customFS {
	return &customFS{fs, files}
}

func (cfs *customFS) Open(name string) (fs.File, error) {
	file, err := cfs.FS.Open(name)
	if err != nil {
		return nil, err
	}

	if info, ok := cfs.files[name]; ok {
		return &customFile{file, info}, nil
	}

	return file, nil
}

type customFile struct {
	fs.File
	info customFileInfo
}

func (cf *customFile) Stat() (fs.FileInfo, error) {
	return cf.info.finfo, nil
}

type customFileInfo struct {
	finfo fs.FileInfo
	atime time.Time
}

func newCustomFileInfo(finfo fs.FileInfo, atime time.Time) customFileInfo {
	return customFileInfo{finfo, atime}
}

func (cfi *customFileInfo) Sys() any {
	stat := &syscall.Stat_t{
		Atim: syscall.Timespec{
			Sec:  cfi.atime.Unix(),
			Nsec: int64(cfi.atime.Nanosecond()),
		},
	}
	return stat
}

type errorFS struct {
	fs.FS
	triggerInfoError bool
	triggerRootError bool
}

func newErrorFS(fs fs.FS, triggerInfoError bool, triggerRootError bool) *errorFS {
	return &errorFS{fs, triggerInfoError, triggerRootError}
}

func (efs *errorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(efs.FS, name)
	if err != nil {
		return nil, err
	}

	if efs.triggerInfoError {
		for i, entry := range entries {
			if entry.Name() == "error_info.txt" {
				entries[i] = &errorDirEntry{entry}
				break
			}
		}
	}

	return entries, nil
}

// Open opens the named file. implements fs.FS
func (efs *errorFS) Open(name string) (fs.File, error) { // go-golangci-lint
	if efs.triggerRootError && name == "." {
		return nil, fmt.Errorf("simulated error")
	}

	file, err := efs.FS.Open(name)
	if err != nil {
		return nil, err
	}

	if name == "error.txt" {
		return &errorFile{file}, nil
	}

	return file, nil
}
