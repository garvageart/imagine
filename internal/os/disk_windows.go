package os

import (
	"path/filepath"

	"golang.org/x/sys/windows"
)

// GetTotalDiskSpace returns the total disk space in bytes for a given path.
// For Windows, it uses windows.GetDiskFreeSpaceEx.
func GetTotalDiskSpace(path string) (uint64, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}

	if len(absPath) == 2 && absPath[1] == ':' {
		absPath += `\`
	}

	var freeBytesAvailableToCaller uint64
	var totalNumberOfBytes uint64
	var totalNumberOfFreeBytes uint64

	pathPtr, err := windows.UTF16PtrFromString(absPath)
	if err != nil {
		return 0, err
	}

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailableToCaller,
		&totalNumberOfBytes,
		&totalNumberOfFreeBytes,
	)
	if err != nil {
		return 0, err
	}

	total := totalNumberOfBytes

	return total, nil
}
