package log

import (
	"fmt"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"imagine/utils" 
)

func SetupLogHandlers() []slog.Handler {
	logShowRecordEnv := os.Getenv("LOG_SHOW_RECORD")
	shouldAddSource := logShowRecordEnv == "true" || logShowRecordEnv == ""
	isProduction := utils.IsProduction

	logFileJSON := FileLog{
		Directory: LogDirectoryDefault,
		Filename:  fmt.Sprintf("%s.json", LogFileFormatDefault),
	}

	consoleHandlerOpts := slog.HandlerOptions{
		AddSource: shouldAddSource,
		Level:     slog.LevelDebug,
	}

	fileHandlerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}

	var consoleLogger slog.Handler
	if isProduction {
		// Production logger with no colour
		consoleLogger = slog.NewTextHandler(os.Stderr, &fileHandlerOpts)
	} else {
		// Setups up colour logger
		consoleLogger = NewColourLogger(&ImalogHandlerOptions{
			HandlerOptions: &consoleHandlerOpts,
			Writer:         os.Stderr, // Explicitly set writer for colour logger
		})
	}

	// Everything below is AI generated when I was trying to fix my git fuck up so idk
	// Open the log file for the JSON handler
	logFileWriter, err := logFileJSON.Open(LogFileDate) // Use LogFileDate or an appropriate date string
	if err != nil {
		// Fallback or panic if file cannot be opened
		slog.Error("Failed to open log file for JSON handler", "path", logFileJSON.FilePath(), "error", err)
		// Potentially return only consoleLogger or panic, depending on desired behavior
		return []slog.Handler{consoleLogger}
	}
	// Note: logFileWriter needs to be closed on application shutdown.

	return []slog.Handler{
		slog.NewJSONHandler(logFileWriter, &fileHandlerOpts), // Use fileHandlerOpts for file logger
		consoleLogger,
	}
}

func CreateLogger(handlers []slog.Handler) *slog.Logger {
	logger := slog.New(slogmulti.Fanout(handlers...))
	return logger
}
