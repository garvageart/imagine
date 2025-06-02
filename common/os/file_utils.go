package libos

import (
	"os"
	"path/filepath"
)

type File struct {
	Path string
}

type OsPerm struct {
	DirPerm  os.FileMode
	FilePerm os.FileMode
}

func (fl File) Open(date string) (file *os.File, err error) { // date string might be unused if Path is absolute and complete
	path := fl.Path

	// Create directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(path), os.ModePerm) // os.ModePerm (0777) is often too permissive for production
	if err != nil {
		return file, err
	}

	// Using all these flags allows us to append to the file not overwrite the data lmao (important!
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (fl File) Write(data []byte) (n int, err error) {
	file, err := fl.Open("") // Assuming date is not strictly needed here or should be passed
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(data)
}

func (fl File) Read() ([]byte, error) {
	return os.ReadFile(fl.Path)
}
