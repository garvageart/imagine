package libos

import (
	"fmt"
	"os"
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
	return strings.ReplaceAll(path, "\\", "/")
}

