package embedutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// copyOpts stores options for copying an fs.FS
type copyOpts struct {
	// PathFunc is called on each file path to
	// modify the file name or location based on the desired output format
	PathFunc func(path string) (string, error)

	// ContentFunc is called on the contents of each file
	// to modify the contents based on the desired output format
	ContentFunc func(data []byte) ([]byte, error)
}

// copyConvert writes the contents of a directory, performing path transformations and content conversions in the process
func copyConvert(sourceDir, outputDir string, opts *copyOpts) ([]string, error) {
	// Map of paths output to outputFS to the unmodified path from sourceFS
	usedPaths := map[string]string{}

	// Store all used paths for indexing later
	paths := []string{}

	sourceFS := os.DirFS(sourceDir)
	err := fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(outputDir, path), 0o755)
		}

		// Read file content
		content, err := fs.ReadFile(sourceFS, path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Set output path and content
		outPath := path
		outContent := content

		// Modify path and content if requested
		if opts != nil {
			outPath, err = opts.PathFunc(path)
			if err != nil {
				return err
			}

			outContent, err = opts.ContentFunc(content)
			if err != nil {
				return err
			}
		}

		// Check for path collisions before writing
		for usedPath, usedOriginal := range usedPaths {
			if outPath == usedPath {
				return fmt.Errorf("path collision: path %q already used for source document %q", usedPath, usedOriginal)
			}
		}

		paths = append(paths, outPath)

		return os.WriteFile(filepath.Join(outputDir, outPath), outContent, 0o644)
	})
	if err != nil {
		return paths, fmt.Errorf("failed to walk fs: %w", err)
	}

	return paths, nil
}
