// Package fsutil contains utility functions for working with the filesystem.
package fsutil

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// FSUtil contains common utilities for working with a filesystem.
type FSUtil struct {
	rootDir string
}

// NewFSUtil creates a new FSUtil instance. A temporary directory is created with the given prefix.
// The directory needs to be removed by the caller with the built in `Close` method.
// Defer `Close` is not recommended as it ignores any errors while closing the complex filesystem.
func NewFSUtil(prefix string) (*FSUtil, error) {
	tempDir, err := os.MkdirTemp("", prefix+"*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	return &FSUtil{rootDir: tempDir}, nil
}

// Close removes the root directory.
func (f *FSUtil) Close() error {
	if err := os.RemoveAll(f.rootDir); err != nil {
		return fmt.Errorf("failed to remove root dir: %w", err)
	}
	return nil
}

// RootDir returns the root directory.
func (f *FSUtil) RootDir() string {
	return f.rootDir
}

// AddFileWithData creates a file with the given data.
func (f *FSUtil) AddFileWithData(fPath string, data []byte) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// AddFileOfSize creates a file with the given size, filled with random data.
func (f *FSUtil) AddFileOfSize(fPath string, size int64) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, io.LimitReader(rand.Reader, size))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// AddFileOfSizeDeterministic creates a file with the given size, filled with deterministic data.
func (f *FSUtil) AddFileOfSizeDeterministic(fPath string, size int64) error {
	file, err := f.createPathAndFile(fPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, io.LimitReader(zeroReader{}, size))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", file.Name(), err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %s: %w", file.Name(), err)
	}
	return nil
}

// joinRelative joins the given path to the root dir after checking that the path is relative.
func (f *FSUtil) joinRelative(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path %s is absolute. All FSUtil paths are relative", path)
	}
	return filepath.Join(f.rootDir, path), nil
}

// createPathAndFile creates the path and file.
// need to close file after use
func (f *FSUtil) createPathAndFile(path string) (*os.File, error) {
	fPath, err := f.joinRelative(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create dir %s: %w", fPath, err)
	}
	file, err := os.Create(fPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", fPath, err)
	}
	return file, nil
}

// ToFS returns the root directory as a fs.FS.
// the returned fs.FS is read-only and implements fs.StatFS
func (f *FSUtil) ToFS() (fs.FS, error) {
	if f.rootDir == "" {
		return nil, fmt.Errorf("rootDir is empty")
	}
	return os.DirFS(f.rootDir), nil
}
