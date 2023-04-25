package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreatePathForFile creates a directory if one does not exist that can contain the provided file path.
// no file is created, only the directory path is created.
func CreatePathForFile(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0777); err != nil {
		return fmt.Errorf("create file path: %w", err)
	}
	return nil
}
