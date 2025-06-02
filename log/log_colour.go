package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"io"
	"log/slog"
	"strings"
	"sync"
)

const (
	Reset = "\033[0m"

	Black        = 30
	Red          = 31
	Green        = 32
	Yellow       = 33
	Blue         = 34
	Magenta      = 35
	Cyan         = 36
	LightGray    = 37
	DarkGray     = 90
	LightRed     = 91
	LightGreen   = 92
	LightYellow  = 93
	LightBlue    = 94
	LightMagenta = 95
	LightCyan    = 96
	White        = 97
)

const (
	timeFormat = "[01-02-2006 15:04:05.000]"
)

type LogMessage struct {
	Time        string
	LevelString string
	Level       slog.Level
	Message     string
	Bytes       string
}

type SlogColourHandler struct {
	handler          slog.Handler
	buffer           *bytes.Buffer
	writer           io.Writer
	mutex            *sync.Mutex
	showRecord       bool
	outputEmptyAttrs bool
	replaceAttrs     func([]string, slog.Attr) slog.Attr
}


func RGBToAnsiString(r, g, b uint8) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%dm%s%s", colorCode, v, Reset)
}

func (h *SlogColourHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *SlogColourHandler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.mutex.Lock()
	defer func() {
		h.buffer.Reset()
		h.mutex.Unlock()
	}()

	if err := h.handler.Handle(ctx, r); err != nil {
	return map[string]any{}, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buffer.Bytes(), &attrs)

	if err != nil {
	return map[string]any{}, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}

	return attrs, nil
}

func SuppressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
			}
		if next == nil {
			return a
		}

		return next(groups, a)
		}
}

func (h *SlogColourHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *SlogColourHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *SlogColourHandler) Handle(ctx context.Context, record slog.Record) error {
	var attrs map[string]any
	var err error
	var attrsAsBytes []byte

	var levelString string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(record.Level),
	}

	if h.replaceAttrs != nil {
		levelAttr = h.replaceAttrs([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		levelString = levelAttr.Value.String() + ":"

		if record.Level <= slog.LevelDebug {
			levelString = colorize(LightGreen, levelString)
		} else if record.Level <= slog.LevelInfo {
			levelString = colorize(Cyan, levelString)
		} else if record.Level < slog.LevelWarn {
			levelString = colorize(LightBlue, levelString)
		} else if record.Level < slog.LevelError {
			levelString = colorize(LightYellow, levelString)
		} else if record.Level <= slog.LevelError+1 {
			levelString = colorize(LightRed, levelString)
		} else if record.Level > slog.LevelError+1 {
			levelString = colorize(LightMagenta, levelString)
		}
	}

	attrs, err = h.computeAttrs(ctx, record)
	if err != nil {
		return err
	}

	if h.outputEmptyAttrs || len(attrs) > 0 {
		attrsAsBytes, err = json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			return fmt.Errorf("error when marshaling attrs: %w", err)
		}
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(record.Time.Format(timeFormat)),
	}

	if h.replaceAttrs != nil {
		timeAttr = h.replaceAttrs([]string{}, timeAttr)
	}

	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(LightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(record.Message),
	}

	if h.replaceAttrs != nil {
		msgAttr = h.replaceAttrs([]string{}, msgAttr)
	}

	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(White, msgAttr.Value.String())
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}

	if len(levelString) > 0 {
		out.WriteString(levelString)
		out.WriteString(" ")
	}

	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}

	// 2 is the minimum length of a []byte when created as a literal
	// Check this so that we don't display an empty []byte string in the console
	// when there are no properties to show
	//
	// TODO: Use the ShowSource field or check if some sort of debug env variable is set,
	// in addition the length check to determine whether we should output
	// any additional messages. This handles any weird potential edges cases
	// Which? Idk but I'd rather not find out lmao
	// This is fine for now
	if len(attrsAsBytes) > 2 {
		out.WriteString(colorize(DarkGray, string(attrsAsBytes)))
	}

	_, err = io.WriteString(h.writer, out.String()+"\n")
	return err
}


func NewColourLogger(opts *ImalogHandlerOptions) *SlogColourHandler {
	if opts == nil {
		opts = &ImalogHandlerOptions{}
	}

	buffer := &bytes.Buffer{}

	return &SlogColourHandler{
		buffer:     buffer,
		writer:     opts.Writer,
		mutex:      &sync.Mutex{},
		showRecord: opts.ShowSource,
		handler: slog.NewJSONHandler(buffer, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: SuppressDefaults(opts.ReplaceAttr),
		}),
		outputEmptyAttrs: opts.OutputEmptyAttrs,
	}
}


