package logger

import (
	"io"
	"log/slog"
	"os"

	"regexp"
	"runtime"

	"github.com/dromara/carbon/v2"

	"imagine/internal/utils"
)

const (
	ConsoleDateTimeDefaultFormatCarbon = "d-m-Y H:i:s"
)

const (
	// idk wtf this is about
	ServerLoggerGroupKey = "http-chi-server"
	// NOTE: Only here for reference, not currently used in any logic
	LogFileNameFormat = "{AppName}-{AppVersion}-{DateTime}-{Context}[?].log"
)

var (
	LogFileDate    = carbon.Now().Format(FileDateTimeDefaultFormatCarbon)
	DateTimeStdOut = carbon.Now().Format(ConsoleDateTimeDefaultFormatCarbon)
)

var (
	LoggerProgramInfoGroup = slog.Group("program_info",
		slog.String("go_version", runtime.Version()),
		slog.String("environment", utils.Environment),
		slog.String("os", runtime.GOOS),
		slog.Int("pid", os.Getpid()),
		slog.String("app_version", utils.GetAppVersion()),
	)
)

type LogFormat struct {
	Json any
	Text any
}

type ImalogHandlerOptions struct {
	// Not yet implemented
	ShowSource       bool
	Format           LogFormat
	Writer           io.Writer
	OutputEmptyAttrs bool
	*slog.HandlerOptions
}

func StripAnsi(str string) string {
	regex := regexp.MustCompile("/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g")
	return regex.ReplaceAllString(str, "")
}
