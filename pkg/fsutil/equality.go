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

// EqualFilesystem checks that the filesystems (excluding hidden files/dirs) are identical.
func EqualFilesystem(fsA, fsB fs.FS, opts ComparisonOpts) error {

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
			return fmt.Errorf("File not found in fsB: %s", path)
		}
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			return err
		}
		if opts.Contents {
			fA, err := fsA.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file in fsA: %w", err)
			}
			fB, err := fsB.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file in fsB: %w", err)
			}
			if err := compareFileContents(fA, fB); err != nil {
				return fmt.Errorf("failed to compare file contents for path %s: %w", path, err)
			}
		}

	}

	for path, infoA := range fsInfoA.dirs {
		infoB, ok := fsInfoB.dirs[path]
		if !ok {
			return fmt.Errorf("Dir not found in fsB: %s", path)
		}
		if err := compareFinfo(path, infoA, infoB, opts); err != nil {
			return err
		}
	}

	return nil
}

// DiffFS returns the differences between two filesystems. (A-B)
func DiffFS(fsA, fsB fs.FS, opts ComparisonOpts) ([]fs.FileInfo, error) {
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
		}

		if opts.Contents {
			fA, err := fsA.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open file in fsA: %w", err)
			}
			fB, err := fsB.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open file in fsB: %w", err)
			}
			if err := compareFileContents(fA, fB); err != nil {
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
	Name     bool // Compare name
	Size     bool // Compare size
	Dir      bool // Compare dir
	Mode     bool // Compare mode
	Contents bool // Compare content
}

var (
	// DefaultComparisonOpts compares only the name, size, dir, and mode of fs.FileInfo
	DefaultComparisonOpts = ComparisonOpts{
		Name:     true,
		Size:     true,
		Dir:      true,
		Mode:     true,
		Contents: false,
	}
	// AllComparisonOpts compares all fields of fs.FileInfo, including file contents
	AllComparisonOpts = ComparisonOpts{
		Name:     true,
		Size:     true,
		Dir:      true,
		Mode:     true,
		Contents: true,
	}
)

func compareFinfo(path string, a, b fs.FileInfo, opts ComparisonOpts) error {
	if opts.Name && a.Name() != b.Name() {
		return fmt.Errorf("Names should be equal for path: %s, a: %s, b: %s", path, a.Name(), b.Name())
	}
	if opts.Dir && a.IsDir() != b.IsDir() {
		return fmt.Errorf("IsDir should be equal for path: %s, a: %v, b: %v", path, a.IsDir(), b.IsDir())
	}
	if opts.Size && a.Size() != b.Size() {
		return fmt.Errorf("Sizes should be equal for path: %s, a: %d, b: %d", path, a.Size(), b.Size())
	}
	if opts.Mode && a.Mode() != b.Mode() {
		return fmt.Errorf("Modes should be equal for path: %s, a: %v, b: %v", path, a.Mode(), b.Mode())
	}
	return nil
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
