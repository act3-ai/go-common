package fsutil

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFSUtilAndClose(t *testing.T) {
	fs, err := NewFSUtil("test")
	require.NoError(t, err, "NewFSUtil should not return an error")

	// Check if the directory exists and is a directory
	dirInfo, err := os.Stat(fs.RootDir)
	require.NoError(t, err, "RootDir should exist")
	assert.True(t, dirInfo.IsDir(), "RootDir should be a directory")

	// Create a test file in the root directory
	testFilePath := filepath.Join(fs.RootDir, "test.txt")
	err = ioutil.WriteFile(testFilePath, []byte("test"), 0644)
	require.NoError(t, err, "Creating a test file should not return an error")

	// Close FSUtil and remove the temporary directory
	err = fs.Close()
	require.NoError(t, err, "Close should not return an error")

	// Check if the temporary directory was removed
	_, err = os.Stat(fs.RootDir)
	assert.Error(t, err, "RootDir should not exist after Close")
	assert.True(t, os.IsNotExist(err), "RootDir should be removed")

}

func TestAddFileWithData(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		data   []byte
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "data.txt",
			data: []byte("Hello, world!"),
		},
		{
			name: "Nested relative path",
			path: "nested/data.txt",
			data: []byte("Nested file content"),
		},
		{
			name: "Empty file",
			path: "empty.txt",
			data: []byte{},
		},
		{
			name:   "Invalid absolute path",
			path:   "/invalid/data.txt",
			data:   []byte("Invalid absolute path"),
			errMsg: "path /invalid/data.txt is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddFileWithData(tc.path, tc.data)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileWithData should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileWithData should not return an error")

				// Check if the file exists and has the correct content
				filePath := filepath.Join(fs.RootDir, tc.path)
				fileContent, err := ioutil.ReadFile(filePath)
				require.NoError(t, err, "File should be readable")
				assert.Equal(t, tc.data, fileContent, "File content should match the provided data")
			}
		})
	}
}

func TestAddFileOfSize(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		size   int64
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "random.txt",
			size: 1024,
		},
		{
			name: "Nested relative path",
			path: "nested/random.txt",
			size: 512,
		},
		{
			name: "Empty file",
			path: "empty.txt",
			size: 0,
		},
		{
			name:   "Invalid absolute path",
			path:   "/invalid/random.txt",
			size:   1024,
			errMsg: "path /invalid/random.txt is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddFileOfSize(tc.path, tc.size)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileOfSize should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileOfSize should not return an error")

				// Check if the file exists and has the correct size
				filePath := filepath.Join(fs.RootDir, tc.path)
				fileInfo, err := os.Stat(filePath)
				require.NoError(t, err, "File should be stat-able")
				assert.Equal(t, tc.size, fileInfo.Size(), "File size should match the provided size")
			}
		})
	}
}

func TestAddFileOfSizeDeterministic(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		size   int64
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "deterministic.txt",
			size: 1024,
		},
		{
			name: "Nested relative path",
			path: "nested/deterministic.txt",
			size: 512,
		},
		{
			name: "Empty file",
			path: "empty.txt",
			size: 0,
		},
		{
			name:   "Invalid absolute path",
			path:   "/invalid/deterministic.txt",
			size:   1024,
			errMsg: "path /invalid/deterministic.txt is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddFileOfSizeDeterministic(tc.path, tc.size)
			if tc.errMsg != "" {
				require.Error(t, err, "AddFileOfSizeDeterministic should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddFileOfSizeDeterministic should not return an error")

				// Check if the file exists and has the correct size
				filePath := filepath.Join(fs.RootDir, tc.path)
				fileInfo, err := os.Stat(filePath)
				require.NoError(t, err, "File should be stat-able")
				assert.Equal(t, tc.size, fileInfo.Size(), "File size should match the provided size")

				// Check if the content is deterministic (all zeros)
				fileContent, err := ioutil.ReadFile(filePath)
				require.NoError(t, err, "File should be readable")
				for _, b := range fileContent {
					assert.Equal(t, byte(0), b, "File content should be all zeros")
				}
			}
		})
	}
}

func TestToFS(t *testing.T) {
	fsUtil, err := NewFSUtil("test")
	require.NoError(t, err, "NewFSUtil should not return an error")
	defer fsUtil.Close()

	err = fsUtil.AddFileWithData("testfile.txt", []byte("Test data"))
	require.NoError(t, err, "AddFileWithData should not return an error")

	readOnlyFS, err := fsUtil.ToFS()
	require.NoError(t, err, "ToFS should not return an error")

	data, err := fs.ReadFile(readOnlyFS, "testfile.txt")
	require.NoError(t, err, "ReadFile should not return an error")
	assert.Equal(t, []byte("Test data"), data, "File content should match the provided data")
}

func TestToFSFailEmptyRootDir(t *testing.T) {
	// Create an FSUtil instance with an empty rootDir
	fsUtil := &FSUtil{}

	readOnlyFS, err := fsUtil.ToFS()
	assert.Nil(t, readOnlyFS, "ToFS should return a nil FS when rootDir is empty")
	assert.EqualError(t, err, "rootDir is empty", "ToFS should return an error when rootDir is empty")
}

func TestEqualFilesystem(t *testing.T) {
	testCases := []struct {
		name          string
		fsASetup      func(*FSUtil) error
		fsBSetup      func(*FSUtil) error
		shouldBeEqual bool
	}{
		{
			name: "Both empty",
			fsASetup: func(fsA *FSUtil) error {
				return nil
			},
			fsBSetup: func(fsB *FSUtil) error {
				return nil
			},
			shouldBeEqual: true,
		},
		{
			name: "Same structure and content",
			fsASetup: func(fsA *FSUtil) error {
				if err := fsA.AddFileWithData("file.txt", []byte("test content")); err != nil {
					return err
				}
				if err := fsA.AddFileOfSize("file_random.bin", 256); err != nil {
					return err
				}
				if err := fsA.AddFileOfSizeDeterministic("file_zeros.bin", 256); err != nil {
					return err
				}
				return nil
			},
			fsBSetup: func(fsB *FSUtil) error {
				if err := fsB.AddFileWithData("file.txt", []byte("test content")); err != nil {
					return err
				}
				if err := fsB.AddFileOfSize("file_random.bin", 256); err != nil {
					return err
				}
				if err := fsB.AddFileOfSizeDeterministic("file_zeros.bin", 256); err != nil {
					return err
				}
				return nil
			},
			shouldBeEqual: true,
		},
		{
			name: "Different structure",
			fsASetup: func(fsA *FSUtil) error {
				if err := fsA.AddFileWithData("file.txt", []byte("test content")); err != nil {
					return err
				}
				return nil
			},
			fsBSetup: func(fsB *FSUtil) error {
				if err := fsB.AddFileWithData("subdir/file.txt", []byte("test content")); err != nil {
					return err
				}
				return nil
			},
			shouldBeEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fsA, err := NewFSUtil("test_fsA")
			assert.NoError(t, err, "NewFSUtil should not return an error")
			defer fsA.Close()

			fsB, err := NewFSUtil("test_fsB")
			assert.NoError(t, err, "NewFSUtil should not return an error")
			defer fsB.Close()
			err = tc.fsASetup(fsA)
			assert.NoError(t, err, "fsA setup should not return an error")

			err = tc.fsBSetup(fsB)
			assert.NoError(t, err, "fsB setup should not return an error")

			fsAFS, err := fsA.ToFS()
			assert.NoError(t, err, "fsA ToFS should not return an error")

			fsBFS, err := fsB.ToFS()
			assert.NoError(t, err, "fsB ToFS should not return an error")

			if tc.shouldBeEqual {
				assert.NoError(t, EqualFilesystem(fsAFS, fsBFS), "EqualFilesystem should return an error")
			} else {
				assert.Error(t, EqualFilesystem(fsAFS, fsBFS), "EqualFilesystem should return an error")
			}
		})
	}
}
