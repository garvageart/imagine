package log

import (
	"fmt"
	"log/slog"
	"os"

	"imagine/utils"

	slogmulti "github.com/samber/slog-multi"
)

var (
	GlobalLogger    = CreateLogger(SetupDefaultLogHandlers())
	DefaultLogLevel = func() slog.Level {
		logDebugEnvVar := os.Getenv("LOG_DEBUG")
		if !utils.IsProduction ||
			logDebugEnvVar != "" ||
			logDebugEnvVar == "true" {
			return slog.LevelDebug
		}

		return slog.LevelInfo
	}()
)

func SetupDefaultLogHandlers() []slog.Handler {
	logShowRecordEnv := os.Getenv("LOG_SHOW_RECORD")
	shouldAddSource := logShowRecordEnv == "true"
	isProduction := utils.IsProduction

	logFileJSON := FileLog{
		Directory: LogDirectoryDefault,
		Filename:  fmt.Sprintf("%s.json", LogFileFormatDefault),
	}

	consoleHandlerOpts := slog.HandlerOptions{
		AddSource: shouldAddSource,
		Level:     DefaultLogLevel,
	}

	fileHandlerOpts := slog.HandlerOptions{
		AddSource: true,
		Level:     DefaultLogLevel,
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

	return []slog.Handler{
		slog.NewJSONHandler(logFileJSON, &fileHandlerOpts),
		consoleLogger,
	}
}

func CreateLogger(handlers []slog.Handler) *slog.Logger {
	logger := slog.New(slogmulti.Fanout(handlers...))
	logger = logger.With(LoggerProgramInfoGroup)

	return logger
}

func CreateDefaultLogger() *slog.Logger {
	return CreateLogger(SetupDefaultLogHandlers())
}
