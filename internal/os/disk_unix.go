//go:build !windows

package os

import (
	"syscall"
)

// GetTotalDiskSpace returns the total disk space in bytes for the given path.
func GetTotalDiskSpace(path string) (uint64, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return 0, err
	}
	// Calculate total disk space in bytes
	// f_bsize is the optimal transfer block size
	// f_blocks is the total number of blocks
	total := fs.Blocks * uint64(fs.Bsize)
	return total, nil
}