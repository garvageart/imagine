package images

import (
	"imagine/internal/config"
	"os"
	"strings"
)

var (
	Directory = func() string {
		cfg, err := config.ReadConfig()
		if err != nil {
			panic(err)
		}
		baseDir := cfg.GetString("base_directory")
		if strings.TrimSpace(baseDir) == "" {
			panic("base directory is not set in config")
		}

		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			panic("base directory does not exist")
		}

		dir := cfg.GetString("upload.location")
		if strings.TrimSpace(dir) == "" {
			panic("upload location is not set in config")
		}

		dir = baseDir + "/" + dir

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		return dir
	}()
)
