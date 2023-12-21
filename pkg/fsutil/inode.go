//go:build linux || darwin

package fsutil

import (
	"io/fs"
	"syscall"
)

// GetInode returns the inode for a file.
func GetInode(fi fs.FileInfo) (uint64, error) {
	return fi.Sys().(*syscall.Stat_t).Ino, nil
}
