package fsutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// DeepEqualFilesystem checks that the filesystems (excluding hidden files/dirs) are identical.
// Equality based on comparion opts.
// It also checks that matching files have identical content.
func DeepEqualFilesystem(fsA, fsB fs.FS, opts ComparisonOpts) error {
	return equalFilesystem(fsA, fsB, opts, true)
}

// EqualFilesystem checks that the filesystems (excluding hidden files/dirs) are identical.
// Equality based on comparion opts.
func EqualFilesystem(fsA, fsB fs.FS, opts ComparisonOpts) error {
	return equalFilesystem(fsA, fsB, opts, false)
}

// equalFilesystem checks that the filesystems (excluding hidden files/dirs) are identical.
func equalFilesystem(fsA, fsB fs.FS, opts ComparisonOpts, deep bool) (err error) {
	fsInfoA, err := getFSInfo(fsA)
	if err != nil {
		return fmt.Errorf("failed to get fsInfo for fsA: %w", err)
	}
	fsInfoB, err := getFSInfo(fsB)
	if err != nil {
		return fmt.Errorf("failed to get fsInfo for fsB: %w", err)
	}

	for path, infoA := range fsInfoA.files {
		infoB, ok := fsInfoB.files[path]
		if !ok {
			return fmt.Errorf("file not found in fsB: %s", path)
		}
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			return err
		}
		if deep {
			if err := openAndCompare(fsA, fsB, path); err != nil {
				return fmt.Errorf("failed to compare file contents for path %s: %w", path, err)
			}
		}

	}

	for path, infoA := range fsInfoA.dirs {
		infoB, ok := fsInfoB.dirs[path]
		if !ok {
			return fmt.Errorf("dir not found in fsB: %s", path)
		}
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			return err
		}
	}

	return nil
}

// DeepDiffFS returns the differences between two filesystems. (A-B)
// File contents are also compared
func DeepDiffFS(fsA, fsB fs.FS, opts ComparisonOpts) ([]fs.FileInfo, error) {
	return diffFS(fsA, fsB, opts, true)
}

// DiffFS returns the differences between two filesystems. (A-B)
func DiffFS(fsA, fsB fs.FS, opts ComparisonOpts) ([]fs.FileInfo, error) {
	return diffFS(fsA, fsB, opts, false)
}

// diffFS returns the differences between two filesystems. (A-B)
// differences are determined by opts.
// if deep is true, then the contents of files are also compared.
func diffFS(fsA, fsB fs.FS, opts ComparisonOpts, deep bool) ([]fs.FileInfo, error) {
	fsInfoA, err := getFSInfo(fsA)
	if err != nil {
		return nil, fmt.Errorf("failed to get fsInfo for fsA: %w", err)
	}
	fsInfoB, err := getFSInfo(fsB)
	if err != nil {
		return nil, fmt.Errorf("failed to get fsInfo for fsB: %w", err)
	}

	var diffs []fs.FileInfo

	for path, infoA := range fsInfoA.files {
		infoB, ok := fsInfoB.files[path]
		// if fileA not in fsB, add to diffs
		if !ok {
			diffs = append(diffs, infoA)
			continue
		}
		// if fileA in fsB but not equal, add to diffs
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			diffs = append(diffs, infoA)
			// if no differences in file info, and deep, compare file contents
			// no need to compare contents if there are differences in file info
		} else if deep {
			if err := openAndCompare(fsA, fsB, path); err != nil {
				diffs = append(diffs, infoA)
			}
		}
	}

	for path, infoA := range fsInfoA.dirs {
		infoB, ok := fsInfoB.dirs[path]
		if !ok {
			diffs = append(diffs, infoA)
			continue
		}
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			diffs = append(diffs, infoA)
		}
	}

	return diffs, nil
}

type fsInfo struct {
	files map[string]os.FileInfo
	dirs  map[string]os.FileInfo
}

func getFSInfo(fsys fs.FS) (*fsInfo, error) {
	fsI := &fsInfo{
		files: make(map[string]os.FileInfo),
		dirs:  make(map[string]os.FileInfo),
	}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		fileInfo, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", path, err)
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			fsI.dirs[path] = fileInfo
		} else {
			if strings.HasPrefix(d.Name(), ".") {
				return nil
			}
			fsI.files[path] = fileInfo
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk filesystem: %w", err)
	}
	return fsI, nil
}

// ComparisonOpts stores options for comparing fs.FileInfo equality
type ComparisonOpts struct {
	Name bool // Compare name
	Size bool // Compare size
	Mode bool // Compare mode
}

var (
	// DefaultComparisonOpts compares only the name, size, dir, and mode of fs.FileInfo
	DefaultComparisonOpts = ComparisonOpts{
		Name: true,
		Size: true,
		Mode: true,
	}
)

func compareFinfo(path string, a, b fs.FileInfo, opts ComparisonOpts) error {
	if opts.Name && a.Name() != b.Name() {
		return fmt.Errorf("names should be equal for path: %s, a: %s, b: %s", path, a.Name(), b.Name())
	}
	if a.IsDir() != b.IsDir() {
		return fmt.Errorf("IsDir should be equal for path: %s, a: %v, b: %v", path, a.IsDir(), b.IsDir())
	}
	if opts.Size && a.Size() != b.Size() {
		return fmt.Errorf("sizes should be equal for path: %s, a: %d, b: %d", path, a.Size(), b.Size())
	}
	if opts.Mode && a.Mode() != b.Mode() {
		return fmt.Errorf("modes should be equal for path: %s, a: %v, b: %v", path, a.Mode(), b.Mode())
	}
	return nil
}

// openAndCompare opens two files and compares their contents.
func openAndCompare(a fs.FS, b fs.FS, path string) error {
	fA, err := a.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file in fsA: %w", err)
	}
	defer func() {
		closeErr := fA.Close()
		if err == nil {
			err = closeErr
		}
	}()
	fB, err := b.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file in fsB: %w", err)
	}
	defer func() {
		closeErr := fB.Close()
		if err == nil {
			err = closeErr
		}
	}()
	return compareFileContents(fA, fB)
}

// compareFileContents compares the contents of two files.
func compareFileContents(a, b fs.File) error {
	bufA := make([]byte, 1024)
	bufB := make([]byte, 1024)

	for {
		nA, err := a.Read(bufA)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read from fileA: %w", err)
		}
		nB, err := b.Read(bufB)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read from fileB: %w", err)
		}
		if nA != nB {
			return fmt.Errorf("file sizes should be equal: %d != %d", nA, nB)
		}
		if !bytes.Equal(bufA[:nA], bufB[:nB]) {
			return fmt.Errorf("file contents should be equal: %s != %s", string(bufA[:nA]), string(bufB[:nB]))
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	return nil
}
