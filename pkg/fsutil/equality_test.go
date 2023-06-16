package fsutil

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

var noContentOpts = ComparisonOpts{Name: true, Size: true, Mode: true, Content: false}

var testCases = []struct {
	name        string
	fsA         fstest.MapFS
	fsB         fstest.MapFS
	opts        ComparisonOpts
	shouldError bool
	expectedLen int
}{
	{
		name:        "Empty filesystems",
		fsA:         fstest.MapFS{},
		fsB:         fstest.MapFS{},
		opts:        noContentOpts,
		shouldError: false,
		expectedLen: 0,
	},
	{
		name: "Identical filesystems",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content")},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content")},
		},
		opts:        noContentOpts,
		shouldError: false,
		expectedLen: 0,
	},
	{
		name: "Different filesystems",
		fsA: fstest.MapFS{
			"fileA.txt": &fstest.MapFile{Data: []byte("File A content")},
		},
		fsB: fstest.MapFS{
			"fileB.txt": &fstest.MapFile{Data: []byte("File B content")},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Filesystem with Info() error",
		fsA: fstest.MapFS{
			"file.txt":       &fstest.MapFile{Data: []byte("File content")},
			"error_info.txt": &fstest.MapFile{Data: []byte("Error content")},
		},
		fsB: fstest.MapFS{
			"file.txt":       &fstest.MapFile{Data: []byte("File content")},
			"error_info.txt": &fstest.MapFile{Data: []byte("Error content")},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 0,
	},
	{
		name: "Mismatched names",
		fsA: fstest.MapFS{
			"file_a.txt": &fstest.MapFile{Data: []byte("hello")},
		},
		fsB: fstest.MapFS{
			"file_b.txt": &fstest.MapFile{Data: []byte("hello")},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched sizes",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("hello")},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("hello world")},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched modes",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: 0600},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: 0644},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched directory status",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("content")},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: fs.ModeDir},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Directory missing in fsB",
		fsA: fstest.MapFS{
			"dir_a": &fstest.MapFile{Mode: fs.ModeDir},
		},
		fsB:         fstest.MapFS{}, // Empty MapFS
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched directory names",
		fsA: fstest.MapFS{
			"dir_a": &fstest.MapFile{Mode: fs.ModeDir},
		},
		fsB: fstest.MapFS{
			"dir_b": &fstest.MapFile{Mode: fs.ModeDir},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched directory modes",
		fsA: fstest.MapFS{
			"dir": &fstest.MapFile{Mode: fs.ModeDir | 0755},
		},
		fsB: fstest.MapFS{
			"dir": &fstest.MapFile{Mode: fs.ModeDir | 0700},
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	{
		name: "Mismatched directory status",
		fsA: fstest.MapFS{
			"dir": &fstest.MapFile{Mode: fs.ModeDir},
		},
		fsB: fstest.MapFS{
			"dir": &fstest.MapFile{Data: []byte("content")}, // Not a directory
		},
		opts:        noContentOpts,
		shouldError: true,
		expectedLen: 1,
	},
	// also compare contents
	{
		name: "File contents are equal",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content")},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content")},
		},
		opts:        DefaultComparisonOpts,
		shouldError: false,
		expectedLen: 0,
	},
	{
		name: "File contents are not equal",
		fsA: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content A")},
		},
		fsB: fstest.MapFS{
			"file.txt": &fstest.MapFile{Data: []byte("File content B")},
		},
		opts:        DefaultComparisonOpts,
		shouldError: true,
		expectedLen: 1,
	},
}

func TestEqualFilesystem(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsA := &errorFS{FS: tc.fsA, triggerInfoError: tc.shouldError}
			fsB := &errorFS{FS: tc.fsB, triggerInfoError: tc.shouldError}
			err := EqualFilesystem(fsA, fsB, tc.opts)

			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDiffFS(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsA := tc.fsA
			fsB := tc.fsB
			diffs, err := DiffFS(fsA, fsB, tc.opts)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(diffs) != tc.expectedLen {
				t.Errorf("Expected %d diffs, got %d", tc.expectedLen, len(diffs))
			}
		})
	}
}
