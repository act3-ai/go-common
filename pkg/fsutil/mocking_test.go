package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
)

type errorDirEntry struct { //nolint: unused
	fs.DirEntry
}

func (e *errorDirEntry) Info() (fs.FileInfo, error) { //nolint: unused
	return nil, errors.New("Info error")
}

type errorFile struct { //nolint: unused
	fs.File
}

func (ef *errorFile) Stat() (fs.FileInfo, error) { //nolint: unused
	return nil, fmt.Errorf("simulated error")
}

type errorFS struct { //nolint: unused
	fs.FS
	triggerInfoError bool
	triggerRootError bool
}

func newErrorFS(fs fs.FS, triggerInfoError bool, triggerRootError bool) *errorFS { //nolint: unused
	return &errorFS{fs, triggerInfoError, triggerRootError}
}

func (efs *errorFS) ReadDir(name string) ([]fs.DirEntry, error) { //nolint: unused
	entries, err := fs.ReadDir(efs.FS, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir: %w", err)
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
func (efs *errorFS) Open(name string) (fs.File, error) { //nolint: unused
	if efs.triggerRootError && name == "." {
		return nil, fmt.Errorf("simulated error")
	}

	file, err := efs.FS.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	if name == "error.txt" {
		return &errorFile{file}, nil
	}

	return file, nil
}
