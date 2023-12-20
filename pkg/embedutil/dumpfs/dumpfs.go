// Package dumpfs is a simple utility for writing the contents of an fs.FS to a directory, performing path transformations and content conversions in the process.
package dumpfs

import (
	"fmt"
	"io/fs"

	"git.act3-ace.com/ace/go-common/pkg/fsutil"
)

// Options for dumping an fs.FS
type Options struct {
	// PathFunc is called on each file path to
	// modify the file name or location based on the desired output format
	PathFunc func(path string) (string, error)

	// ContentFunc is called on the contents of each file
	// to modify the contents based on the desired output format
	ContentFunc func(data []byte) ([]byte, error)
}

// DumpFS dumps the contents of an fs.FS into another fs.FS
func DumpFS(sourceFS fs.FS, outputFS *fsutil.FSUtil, opts *Options) ([]string, error) {
	// Map of paths output to outputFS to the unmodified path from sourceFS
	usedPaths := map[string]string{}

	// Store all used paths for indexing later
	paths := []string{}

	err := fs.WalkDir(sourceFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, AddFileWithData handles directory creation
		if d.IsDir() {
			return nil
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

		return outputFS.AddFileWithData(outPath, outContent)
	})
	if err != nil {
		return paths, fmt.Errorf("failed to walk fs: %w", err)
	}

	return paths, nil
}
