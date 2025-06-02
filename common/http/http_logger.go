package http

***REMOVED***
***REMOVED***
	"log/slog"
***REMOVED***
	"strings"

	imalog "imagine/log"

***REMOVED***
***REMOVED***

func setupChiLogHandler(name string***REMOVED*** []slog.Handler {
	httpLogFileDefaults := imalog.LogFileDefaults
	logLevel := func(***REMOVED*** slog.Level {
	***REMOVED***
			return slog.LevelDebug
	***REMOVED***
		return slog.LevelInfo
***REMOVED***(***REMOVED***

	fileHandler := imalog.NewFileLogger(&imalog.ImalogHandlerOptions{
		Writer: imalog.FileLog{
			Directory: httpLogFileDefaults.Directory + "/http",
			Filename:  fmt.Sprintf("%s-http-%s", httpLogFileDefaults.Filename, strings.ReplaceAll(name, "-", "_"***REMOVED******REMOVED***,
***REMOVED***
		ShowSource: true,
		HandlerOptions: &slog.HandlerOptions{
			Level: logLevel,
			// Strip all ANSI codes from the log output set by the
			// go-chi logger middleware. Even if colour logging is disabled is production
			// during development the middleware wraps strings in colour
			ReplaceAttr: func(groups []string, a slog.Attr***REMOVED*** slog.Attr {
				if a.Key == slog.MessageKey {
					a.Value = slog.StringValue(imalog.StripAnsi(a.Value.String(***REMOVED******REMOVED******REMOVED***
			***REMOVED***

				return a
	***REMOVED***
***REMOVED***
***REMOVED******REMOVED***

	consoleHandler := imalog.NewColourLogger(&imalog.ImalogHandlerOptions{
		HandlerOptions: &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: false,
***REMOVED***
		Writer:           os.Stderr,
		OutputEmptyAttrs: true,
***REMOVED******REMOVED***

	return []slog.Handler{
		fileHandler,
		consoleHandler,
***REMOVED***
***REMOVED***

func SetupChiLogger(name string***REMOVED*** *slog.Logger {
	handlers := setupChiLogHandler(name***REMOVED***

	logger := imalog.CreateLogger(handlers***REMOVED***
	return logger
***REMOVED***
