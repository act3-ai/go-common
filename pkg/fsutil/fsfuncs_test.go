package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/djherbis/atime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDirSize tests the DirSize function with different test cases.
func TestDirSize(t *testing.T) {
	t.Run("empty directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-empty-dir")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		fsys := os.DirFS(tmpDir)

		size, err := DirSize(fsys)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("single file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-single-file")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = os.WriteFile(fmt.Sprintf("%s/file.txt", tmpDir), []byte("content"), 0644)
		require.NoError(t, err)

		fsys := os.DirFS(tmpDir)

		size, err := DirSize(fsys)
		assert.NoError(t, err)
		assert.Equal(t, int64(len("content")), size)
	})

	t.Run("nested directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-nested-dir")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = os.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("content1"), 0644)
		require.NoError(t, err)
		err = os.MkdirAll(fmt.Sprintf("%s/subdir", tmpDir), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fmt.Sprintf("%s/subdir/file2.txt", tmpDir), []byte("content2"), 0644)
		require.NoError(t, err)

		fsys := os.DirFS(tmpDir)

		size, err := DirSize(fsys)
		assert.NoError(t, err)
		assert.Equal(t, int64(len("content1")+len("content2")), size)
	})

	t.Run("error on non-existent directory", func(t *testing.T) {
		fsys := os.DirFS("non-existent")

		size, err := DirSize(fsys)
		assert.Error(t, err)
		assert.Equal(t, int64(0), size)
	})
}

// TestGetDirLastUpdate tests the GetDirLastUpdate function with different test cases.
func TestGetDirLastUpdate(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) (fs.FS, func())
		wantErr   bool
	}{
		{
			name: "empty directory",
			setupFunc: func(t *testing.T) (fs.FS, func()) {
				tmpDir, err := os.MkdirTemp("", "test-empty-dir")
				require.NoError(t, err)

				return os.DirFS(tmpDir), func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "single file",
			setupFunc: func(t *testing.T) (fs.FS, func()) {
				tmpDir, err := os.MkdirTemp("", "test-single-file")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file.txt", tmpDir), []byte("content"), 0644)
				require.NoError(t, err)

				return os.DirFS(tmpDir), func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "nested directories",
			setupFunc: func(t *testing.T) (fs.FS, func()) {
				tmpDir, err := os.MkdirTemp("", "test-nested-dir")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("content1"), 0644)
				require.NoError(t, err)
				err = os.MkdirAll(fmt.Sprintf("%s/subdir", tmpDir), 0755)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("%s/subdir/file2.txt", tmpDir), []byte("content2"), 0644)
				require.NoError(t, err)

				return os.DirFS(tmpDir), func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "non-existent directory",
			setupFunc: func(t *testing.T) (fs.FS, func()) {
				return os.DirFS("non-existent"), func() {}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys, cleanup := tc.setupFunc(t)
			defer cleanup()

			lastUpdate, err := GetDirLastUpdate(fsys)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, !lastUpdate.IsZero(), "lastUpdate should not be zero")
			}
		})
	}
}

// TestReadDirSortedByAccessTime tests the ReadDirSortedByAccessTime function with different test cases.
func TestReadDirSortedByAccessTime(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) (fs.FS, string, func())
		wantErr   bool
	}{
		{
			name: "empty directory",
			setupFunc: func(t *testing.T) (fs.FS, string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-empty-dir")
				require.NoError(t, err)

				return os.DirFS(tmpDir), ".", func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "single file",
			setupFunc: func(t *testing.T) (fs.FS, string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-single-file")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file.txt", tmpDir), []byte("content"), 0644)
				require.NoError(t, err)

				return os.DirFS(tmpDir), ".", func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "nested directories",
			setupFunc: func(t *testing.T) (fs.FS, string, func()) {
				tmpDir, err := os.MkdirTemp("", "test-nested-dir")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("content1"), 0644)
				require.NoError(t, err)
				time.Sleep(500 * time.Millisecond) // Add sleep to ensure access time difference
				err = os.MkdirAll(fmt.Sprintf("%s/subdir", tmpDir), 0755)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("%s/subdir/file2.txt", tmpDir), []byte("content2"), 0644)
				require.NoError(t, err)

				return os.DirFS(tmpDir), ".", func() { os.RemoveAll(tmpDir) }
			},
			wantErr: false,
		},
		{
			name: "non-existent directory",
			setupFunc: func(t *testing.T) (fs.FS, string, func()) {
				return os.DirFS("non-existent"), ".", func() {}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys, dir, cleanup := tc.setupFunc(t)
			defer cleanup()

			infos, err := ReadDirSortedByAccessTime(fsys, dir)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(infos) > 1 {
					for i := 1; i < len(infos); i++ {
						assert.True(t, atime.Get(infos[i-1]).Before(atime.Get(infos[i])), "entries should be sorted by access time")
					}
				}
			}
		})
	}
}

func TestGetDirUpdatedPaths(t *testing.T) {
	testCases := []struct {
		name      string
		setupFunc func(t *testing.T) (fs.FS, time.Time, func())
		wantErr   bool
		wantCount int
	}{
		{
			name: "empty directory",
			setupFunc: func(t *testing.T) (fs.FS, time.Time, func()) {
				tmpDir, err := os.MkdirTemp("", "test-empty-dir")
				require.NoError(t, err)

				return os.DirFS(tmpDir), time.Now().Add(-time.Hour), func() { os.RemoveAll(tmpDir) }
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "single file",
			setupFunc: func(t *testing.T) (fs.FS, time.Time, func()) {
				tmpDir, err := os.MkdirTemp("", "test-single-file")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file.txt", tmpDir), []byte("content"), 0644)
				require.NoError(t, err)

				return os.DirFS(tmpDir), time.Now().Add(-time.Hour), func() { os.RemoveAll(tmpDir) }
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "multiple files with varying modification times",
			setupFunc: func(t *testing.T) (fs.FS, time.Time, func()) {
				tmpDir, err := os.MkdirTemp("", "test-multi-file")
				require.NoError(t, err)

				err = os.WriteFile(fmt.Sprintf("%s/file1.txt", tmpDir), []byte("content1"), 0644)
				require.NoError(t, err)
				time.Sleep(500 * time.Millisecond)
				err = os.WriteFile(fmt.Sprintf("%s/file2.txt", tmpDir), []byte("content2"), 0644)
				require.NoError(t, err)
				earliest := time.Now().Add(-(250 * time.Millisecond))

				return os.DirFS(tmpDir), earliest, func() { os.RemoveAll(tmpDir) }
			},
			wantErr:   false,
			wantCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsys, earliest, cleanup := tc.setupFunc(t)
			defer cleanup()

			paths, err := GetDirUpdatedPaths(fsys, earliest)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantCount, len(paths), "unexpected number of updated paths")
			}
		})
	}
}
