// This file uses implicit build constraints to exclude it from non-Windows builds.
package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"syscall"
)

// inspired by https://go.dev/src/os/types_windows.go

// getInode returns the inode for a file on Windows.
// this is a workaround for the lack of inode support on Windows.
// the returned inode is a combination of the volume serial number and file index.
func getInode(fi fs.FileInfo) (uint64, error) {
	var inode uint64
	pathp, err := syscall.UTF16PtrFromString(fi.Name())
	if err != nil {
		return inode, fmt.Errorf("failed to get UTF16 pointer from file name: %w", err)
	}
	attrs := uint32(syscall.FILE_FLAG_BACKUP_SEMANTICS)

	// check if the file is a symlink
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		attrs |= syscall.FILE_FLAG_OPEN_REPARSE_POINT
	}

	// create file handle
	h, err := syscall.CreateFile(pathp, 0, 0, nil, syscall.OPEN_EXISTING, attrs, 0)
	if err != nil {
		return inode, fmt.Errorf("failed to create file handle: %w", err)
	}
	defer syscall.CloseHandle(h)
	var i syscall.ByHandleFileInformation
	err = syscall.GetFileInformationByHandle(h, &i)
	if err != nil {
		return inode, fmt.Errorf("failed to get file information by handle: %w", err)
	}
	inode = uint64(i.VolumeSerialNumber)<<32 | uint64(i.FileIndexHigh)<<32 | uint64(i.FileIndexLow)

	return inode, nil
}
