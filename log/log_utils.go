package log

***REMOVED***
	"io"
	"log/slog"
***REMOVED***
	"regexp"
	"runtime"

***REMOVED***

	"github.com/dromara/carbon/v2"
***REMOVED***

const (
	ConsoleDateTimeDefaultFormatCarbon = "d-m-Y H:i:s"
***REMOVED***

const (
	ServerLoggerGroupKey = "http-chi-server"
	// NOTE: Only here for reference, not currently used in any logic
	LogFileNameFormat = "{AppName***REMOVED***-{AppVersion***REMOVED***-{DateTime***REMOVED***-{Context***REMOVED***[?].log"
***REMOVED***

var (
	LogFileDate    = carbon.Now(***REMOVED***.Format(FileDateTimeDefaultFormatCarbon***REMOVED***
	DateTimeStdOut = carbon.Now(***REMOVED***.Format(ConsoleDateTimeDefaultFormatCarbon***REMOVED***
***REMOVED***

var (
	LoggerProgramInfoGroup = slog.Group(ServerLoggerGroupKey,
		slog.String("go_version", runtime.Version(***REMOVED******REMOVED***,
		slog.String("environment", utils.Environment***REMOVED***,
		slog.String("os", runtime.GOOS***REMOVED***,
		slog.Int("pid", os.Getpid(***REMOVED******REMOVED***,
	***REMOVED***
***REMOVED***

type LogFormat struct {
	Json any
	Text any
***REMOVED***

type ImalogHandlerOptions struct {
	// Not yet implemented
	ShowSource       bool
	Format           LogFormat
	Writer           io.Writer
	OutputEmptyAttrs bool
	*slog.HandlerOptions
***REMOVED***

func StripAnsi(str string***REMOVED*** string {
	regex := regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~]***REMOVED***`***REMOVED***
	return regex.ReplaceAllString(str, ""***REMOVED***
***REMOVED***
