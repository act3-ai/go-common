//go:build linux || darwin

package fsutil

import (
	"io/fs"
	"syscall"
)

// getInode returns the inode for a file.
func getInode(fi fs.FileInfo) (uint64, error) {
	return fi.Sys().(*syscall.Stat_t).Ino, nil
}
