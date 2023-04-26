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

func TestAddDir(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		errMsg string
	}{
		{
			name: "Valid relative path",
			path: "random",
		},
		{
			name: "Nested relative path",
			path: "nested/random",
		},
		{
			name: "Empty",
			path: "",
		},
		{
			name:   "Invalid absolute path",
			path:   "/invalid/random",
			errMsg: "path /invalid/random is absolute. All FSUtil paths are relative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, err := NewFSUtil("test")
			require.NoError(t, err, "NewFSUtil should not return an error")
			defer fs.Close()

			err = fs.AddDir(tc.path)
			if tc.errMsg != "" {
				require.Error(t, err, "AddDir should return an error for invalid input")
				assert.EqualError(t, err, tc.errMsg)
			} else {
				require.NoError(t, err, "AddDir should not return an error")

				// Check if the file exists and has the correct size
				dirPath := filepath.Join(fs.RootDir, tc.path)
				dirInfo, err := os.Stat(dirPath)
				require.NoError(t, err, "File should be stat-able")
				assert.True(t, dirInfo.IsDir(), "File should be a directory")
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
