package log

import (
	"fmt"
	"log/slog"
	"os"

	"path"
	"path/filepath"
	"strings"

	libos "imagine/common/os"
	"imagine/utils"
)

const (
	LogFileExt = "log"
)

var (
	FileDateTimeDefaultFormatCarbon = "dmY_His"
)

var (
	LogFileFormatDefault = fmt.Sprint(
		utils.AppName,
		"-", strings.ReplaceAll(utils.GetAppVersion(), ".", "_"),
		"-", LogFileDate,
	)

	LogDirectoryDefault = func() string {
		return path.Join(libos.CurrentWorkingDirectory, "var", "logs")
	}()

	LogFileDefaults = FileLog{
		Directory: LogDirectoryDefault,
		Filename:  LogFileFormatDefault,
	}
)

type FileLog struct {
	Directory string
	Filename  string
}

func (fl FileLog) Open() (file *os.File, err error) {
	path := fl.FilePath()

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return file, err
	}

	// Using all these flags allows us to append to the file not overwrite the data lmao (important!)
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func (fl FileLog) Write(data []byte) (n int, err error) {
	file, err := fl.Open()

	if err != nil {
		fmt.Println("Error opening log file", err)
		return
	}

	defer file.Close()

	return file.Write(data)
}

func (fl FileLog) FilePath() string {
	return libos.StandardisePaths(path.Join(fl.Directory, fl.Filename+"."+LogFileExt))
}

func NewFileLogger(opts *ImalogHandlerOptions) slog.Handler {
	logFormat := opts.Format

	switch logFormat {
	case logFormat.Text:
		return slog.NewTextHandler(opts.Writer, opts.HandlerOptions)
	default:
		return slog.NewJSONHandler(opts.Writer, opts.HandlerOptions)
	}
}
