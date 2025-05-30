package log

***REMOVED***
***REMOVED***
	"log/slog"
***REMOVED***
	"path"
	"path/filepath"
	"strings"

	"go.les-is.online/imagine/utils"
***REMOVED***

const (
	LogFileExt = "log"
***REMOVED***

var (
	FileDateTimeDefaultFormatCarbon = "dmY_His"
***REMOVED***

var (
	LogFileFormatDefault = fmt.Sprint(
		utils.AppName,
		"-", strings.ReplaceAll(utils.GetAppVersion(***REMOVED***, ".", "_"***REMOVED***,
		"-", LogFileDate***REMOVED***

	LogDirectoryDefault = func(***REMOVED*** string {
		cwd, err := os.Getwd(***REMOVED***

	***REMOVED***
			panic(err***REMOVED***
	***REMOVED***

		return cwd + "/var/logs"
***REMOVED***(***REMOVED***

	LogFileDefaults = FileLog{
		Directory: LogDirectoryDefault,
		Filename:  LogFileFormatDefault,
***REMOVED***
***REMOVED***

type FileLog struct {
	Directory string
	Filename  string
***REMOVED***

func (fl FileLog***REMOVED*** Open(date string***REMOVED*** (file *os.File, err error***REMOVED*** {
	path := fl.FilePath(***REMOVED***

	err = os.MkdirAll(filepath.Dir(path***REMOVED***, os.ModePerm***REMOVED***

***REMOVED***
		fmt.Println("Error creating directory:", err***REMOVED***
		return file, err
***REMOVED***

	// Using all these flags allows us to append to the file not overwrite the data lmao (important!***REMOVED***
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644***REMOVED***
***REMOVED***

func (fl FileLog***REMOVED*** Write(data []byte***REMOVED*** (n int, err error***REMOVED*** {
	path := fl.FilePath(***REMOVED***

	file, err := fl.Open(path***REMOVED***
***REMOVED***
		fmt.Println("Error opening log file", err***REMOVED***
		return
***REMOVED***

	defer file.Close(***REMOVED***

	return file.Write([]byte(data***REMOVED******REMOVED***
***REMOVED***

func (fl FileLog***REMOVED*** FilePath(***REMOVED*** string {
	return path.Join(fl.Directory, fl.Filename+"."+LogFileExt***REMOVED***
***REMOVED***

func NewFileLogger(opts *ImalogHandlerOptions***REMOVED*** slog.Handler {
	logFormat := opts.Format

	switch logFormat {
	case logFormat.Text:
		return slog.NewTextHandler(opts.Writer, opts.HandlerOptions***REMOVED***
	default:
		return slog.NewJSONHandler(opts.Writer, opts.HandlerOptions***REMOVED***
***REMOVED***
***REMOVED***
