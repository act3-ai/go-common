package fsutil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/djherbis/atime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEqualFilesystem(t *testing.T) {
	testCases := []struct {
		name        string
		fsA         fstest.MapFS
		fsB         fstest.MapFS
		opts        ComparisonOpts
		shouldError bool
	}{
		{
			name:        "Empty filesystems",
			fsA:         fstest.MapFS{},
			fsB:         fstest.MapFS{},
			opts:        DefaultComparisonOpts,
			shouldError: false,
		},
		{
			name: "Identical filesystems",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			opts:        DefaultComparisonOpts,
			shouldError: false,
		},
		{
			name: "Different filesystems",
			fsA: fstest.MapFS{
				"fileA.txt": &fstest.MapFile{Data: []byte("File A content")},
			},
			fsB: fstest.MapFS{
				"fileB.txt": &fstest.MapFile{Data: []byte("File B content")},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
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
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched names",
			fsA: fstest.MapFS{
				"file_a.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			fsB: fstest.MapFS{
				"file_b.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched sizes",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("hello world")},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched modes",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: 0600},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: 0644},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched directory status",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("content")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("content"), Mode: fs.ModeDir},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Directory missing in fsB",
			fsA: fstest.MapFS{
				"dir_a": &fstest.MapFile{Mode: fs.ModeDir},
			},
			fsB:         fstest.MapFS{}, // Empty MapFS
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched directory names",
			fsA: fstest.MapFS{
				"dir_a": &fstest.MapFile{Mode: fs.ModeDir},
			},
			fsB: fstest.MapFS{
				"dir_b": &fstest.MapFile{Mode: fs.ModeDir},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched directory modes",
			fsA: fstest.MapFS{
				"dir": &fstest.MapFile{Mode: fs.ModeDir | 0755},
			},
			fsB: fstest.MapFS{
				"dir": &fstest.MapFile{Mode: fs.ModeDir | 0700},
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "Mismatched directory status",
			fsA: fstest.MapFS{
				"dir": &fstest.MapFile{Mode: fs.ModeDir},
			},
			fsB: fstest.MapFS{
				"dir": &fstest.MapFile{Data: []byte("content")}, // Not a directory
			},
			opts:        DefaultComparisonOpts,
			shouldError: true,
		},
		{
			name: "File contents are equal",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			opts:        AllComparisonOpts,
			shouldError: false,
		},
		{
			name: "File contents are not equal",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content A")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content B")},
			},
			opts:        AllComparisonOpts,
			shouldError: true,
		},
	}

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
	testCases := []struct {
		name        string
		fsA         fstest.MapFS
		fsB         fstest.MapFS
		opts        ComparisonOpts
		expectedLen int
	}{
		{
			name:        "Empty filesystems",
			fsA:         fstest.MapFS{},
			fsB:         fstest.MapFS{},
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
			expectedLen: 1,
		},
		{
			name: "Mismatched names",
			fsA: fstest.MapFS{
				"file_a.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			fsB: fstest.MapFS{
				"file_b.txt": &fstest.MapFile{Data: []byte("hello")},
			},
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
			expectedLen: 1,
		},
		{
			name: "Directory missing in fsB",
			fsA: fstest.MapFS{
				"dir_a": &fstest.MapFile{Mode: fs.ModeDir},
			},
			fsB:         fstest.MapFS{}, // Empty MapFS
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
			expectedLen: 1,
		},
		{
			name: "Mismatched directory modes",
			fsA: fstest.MapFS{
				"dir": &fstest.MapFile{Mode: fs.ModeDir | 0755},
			},
			fsB:         fstest.MapFS{"dir": &fstest.MapFile{Mode: fs.ModeDir | 0700}},
			opts:        DefaultComparisonOpts,
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
			opts:        DefaultComparisonOpts,
			expectedLen: 1,
		},
		{
			name: "File contents are equal",
			fsA: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			fsB: fstest.MapFS{
				"file.txt": &fstest.MapFile{Data: []byte("File content")},
			},
			opts:        AllComparisonOpts,
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
			opts:        AllComparisonOpts,
			expectedLen: 1,
		},
	}

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

func TestDirSizeSymlinks(t *testing.T) {
	// Test my test code

	d := t.TempDir()

	// add some files directories
	err := os.Mkdir(filepath.Join(d, "subdir"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(d, "file1"), []byte("the data"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(d, "subdir", "file2"), []byte("in sub dir"), 0666)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Symlink(filepath.Join(d, "file1"), filepath.Join(d, "symlink-to-file1"))
	if err != nil {
		t.Fatal(err)
	}

	err = os.Link(filepath.Join(d, "subdir", "file2"), filepath.Join(d, "hardlink-to-file2"))
	if err != nil {
		t.Fatal(err)
	}

	var size int64
	size, err = DirSize(os.DirFS(d))
	if err != nil {
		t.Fatal(err)
	}

	if size != 8+10 {
		t.Errorf("expected 18 B but got %d B", size)
	}
}

func TestReadDirSortedByAccessTime(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		path        string
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: time.Now(),
					Sys: func() interface{} {
						return &syscall.Stat_t{
							Atim: syscall.Timespec{Sec: time.Now().Unix(), Nsec: int64(time.Now().Nanosecond())},
						}
					}(),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: time.Now(),
					Sys: func() interface{} {
						return &syscall.Stat_t{
							Atim: syscall.Timespec{Sec: time.Now().Add(-time.Hour).Unix(), Nsec: int64(time.Now().Add(-time.Hour).Nanosecond())},
						}
					}(),
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			path:        ".",
			expectedErr: nil,
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: time.Now(),
					Sys: func() interface{} {
						return &syscall.Stat_t{
							Atim: syscall.Timespec{Sec: time.Now().Unix(), Nsec: int64(time.Now().Nanosecond())},
						}
					}(),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: time.Now(),
					Sys: func() interface{} {
						return &syscall.Stat_t{
							Atim: syscall.Timespec{Sec: time.Now().Add(-time.Hour).Unix(), Nsec: int64(time.Now().Add(-time.Hour).Nanosecond())},
						}
					}(),
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return newErrorFS(mapFS, true, false)
			},
			path:        ".",
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			infos, err := ReadDirSortedByAccessTime(fsys, tc.path)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				// Check if the files are sorted by access time
				for i := 0; i < len(infos)-1; i++ {
					assert.True(t, atime.Get(infos[i]).Before(atime.Get(infos[i+1])))
				}
			}
		})
	}
}

func TestGetDirLastUpdate(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now,
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			expectedErr: nil,
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now,
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return newErrorFS(mapFS, true, false)
			},
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			lastUpdate, err := GetDirLastUpdate(fsys)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)

				fileBInfo, err := fsys.Open("fileB.txt")
				require.NoError(t, err)
				fileBStat, err := fileBInfo.Stat()
				require.NoError(t, err)

				assert.Equal(t, fileBStat.ModTime(), lastUpdate)
			}
		})
	}
}

func TestGetDirUpdatedPaths(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		setupFS     func() fs.FS
		earliest    time.Time
		expected    []string
		expectedErr error
	}{
		{
			name: "normal",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-2 * time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				return fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
				}
			},
			earliest: time.Now().Add(-90 * time.Minute),
			expected: []string{"fileB.txt"},
		},
		{
			name: "error_getting_file_info",
			setupFS: func() fs.FS {
				now := time.Now()
				fileA := &fstest.MapFile{
					Data:    []byte("file A content"),
					Mode:    0644,
					ModTime: now.Add(-2 * time.Hour),
				}
				fileB := &fstest.MapFile{
					Data:    []byte("file B content"),
					Mode:    0644,
					ModTime: now.Add(-time.Hour),
				}
				mapFS := fstest.MapFS{
					"fileA.txt": fileA,
					"fileB.txt": fileB,
					"error_info.txt": &fstest.MapFile{
						Data: []byte("error info content"),
						Mode: 0644,
					},
				}
				return newErrorFS(mapFS, true, false)
			},
			earliest:    time.Now().Add(-90 * time.Minute),
			expectedErr: fmt.Errorf("error getting file info: %w", errors.New("Info error")),
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys := tc.setupFS()

			updatedPaths, err := GetDirUpdatedPaths(fsys, tc.earliest)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, updatedPaths)
			}
		})
	}
}
