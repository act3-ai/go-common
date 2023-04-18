package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// EqualFilesystem checks that the filesystems (excluding hidden files/dirs) are identical.
func EqualFilesystem(fsA, fsB fs.FS) error {

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
		if err := compareFinfo(path, infoA, infoB); err != nil {
			return err
		}

	}

	for path, infoA := range fsInfoA.dirs {
		infoB, ok := fsInfoB.dirs[path]
		if !ok {
			return fmt.Errorf("Dir not found in fsB: %s", path)
		}
		if err := compareFinfo(path, infoA, infoB); err != nil {
			return err
		}
	}
	return nil
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

func compareFinfo(path string, a, b fs.FileInfo) error {
	if a.Name() != b.Name() {
		return fmt.Errorf("Names should be equal for path: %s", path)
	}
	if a.Size() != b.Size() {
		return fmt.Errorf("Sizes should be equal for path: %s", path)
	}
	if a.Mode() != b.Mode() {
		return fmt.Errorf("Modes should be equal for path: %s", path)
	}
	if a.IsDir() != b.IsDir() {
		return fmt.Errorf("IsDir should be equal for path: %s", path)
	}
	// TODO: above can all be equal but the content can be different
	// Do we want to check the content?
	return nil
}
