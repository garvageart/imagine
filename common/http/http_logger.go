package http

import (
	"fmt"
	"log/slog"
	"os"

	"strings"

	"imagine/utils"
	imalog "imagine/log"
)

func setupChiLogHandler(name string) []slog.Handler {
	httpLogFileDefaults := imalog.LogFileDefaults
	logLevel := func() slog.Level {
		if !utils.IsProduction {
			return slog.LevelDebug
		}
		return slog.LevelInfo
	}()

	// Setup file logger
	logFileWriter := imalog.FileLog{
		Directory: httpLogFileDefaults.Directory + "/http",
		Filename:  fmt.Sprintf("%s-http-%s", utils.AppName, strings.ReplaceAll(name, "-", "_")),
	}

	fileHandler := imalog.NewFileLogger(&imalog.ImalogHandlerOptions{
		Writer:     logFileWriter, 
		HandlerOptions: &slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
			// Strip all ANSI codes from the log output set by the
			// go-chi logger middleware. Even if colour logging is disabled is production
			// during development the middleware wraps strings in colour
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.MessageKey {
					a.Value = slog.StringValue(imalog.StripAnsi(a.Value.String()))
				}
				return a
			},
		},
	})

	// Setup console logger
	consoleHandler := imalog.NewColourLogger(&imalog.ImalogHandlerOptions{
		HandlerOptions: &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: false,
		},
		Writer:           os.Stderr,
		OutputEmptyAttrs: true,
	})

	return []slog.Handler{
		fileHandler,
		consoleHandler,
	}
}
func SetupChiLogger(name string) *slog.Logger {
	handlers := setupChiLogHandler(name)

	logger := imalog.CreateLogger(handlers)
	return logger
}
