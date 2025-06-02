package log

***REMOVED***
***REMOVED***
	"log/slog"
***REMOVED***

	slogmulti "github.com/samber/slog-multi"

***REMOVED***
***REMOVED***

func SetupLogHandlers(***REMOVED*** []slog.Handler {
	shouldAddSource := os.Getenv("LOG_SHOW_RECORD"***REMOVED*** == "true" || true
	isProduction := utils.IsProduction

	logFileJSON := FileLog{
		Directory: LogDirectoryDefault,
		Filename:  fmt.Sprint(LogFileFormatDefault, ".json"***REMOVED***,
***REMOVED***

	consoleHandlerOpts := slog.HandlerOptions{
		AddSource: shouldAddSource,
		Level:     slog.LevelDebug,
***REMOVED***

	fileHandlerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
***REMOVED***

	var consoleLogger slog.Handler

	if isProduction {
		// Production logger with no colour
		consoleLogger = slog.NewTextHandler(os.Stderr, &fileHandlerOpts***REMOVED***
***REMOVED*** else {
		// Setups up colour logger
		consoleLogger = NewColourLogger(&ImalogHandlerOptions{
			HandlerOptions: &consoleHandlerOpts,
	***REMOVED******REMOVED***
***REMOVED***

	return []slog.Handler{
		slog.NewJSONHandler(logFileJSON, &consoleHandlerOpts***REMOVED***,
		consoleLogger***REMOVED***
***REMOVED***

func CreateLogger(handlers []slog.Handler***REMOVED*** *slog.Logger {
	logger := slog.New(slogmulti.Fanout(handlers...***REMOVED******REMOVED***

	return logger
***REMOVED***
