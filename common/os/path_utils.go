package libos

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	CurrentWorkingDirectory = func() string {
		cwd, err := os.Getwd()
		if err != nil {
			panic(fmt.Errorf("error retrieving current working directory: %w", err))
		}

		return StandardisePaths(cwd)
	}()
)

// Microsoft you will pay for your crimes against standards
func StandardisePaths(path string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(path, "/", "\\")
	} else {
		return strings.ReplaceAll(path, "\\", "/")
	}
}
