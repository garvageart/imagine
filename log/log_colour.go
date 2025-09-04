package log

// Modified from https://github.com/dusted-go/logging/blob/main/prettylog/prettylog.go
// Apache 2.0 License https://github.com/dusted-go/logging/blob/main/LICENSE

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"imagine/internal/ansi"
	"io"
	"log/slog"
	"strings"
	"sync"
)

const (
	TimeFormat = "[01-02-2006 15:04:05.000]"
)

type Handler struct {
	handler          slog.Handler
	recorder         func([]string, slog.Attr) slog.Attr
	buf              *bytes.Buffer
	mu               *sync.Mutex
	writer           io.Writer
	colorize         bool
	outputEmptyAttrs bool
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{handler: h.handler.WithAttrs(attrs), buf: h.buf, recorder: h.recorder, mu: h.mu, writer: h.writer, colorize: h.colorize}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{handler: h.handler.WithGroup(name), buf: h.buf, recorder: h.recorder, mu: h.mu, writer: h.writer, colorize: h.colorize}
}

func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.mu.Lock()
	defer func() {
		h.buf.Reset()
		h.mu.Unlock()
	}()
	if err := h.handler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buf.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	colorize := func(code int, value string) string {
		return value
	}
	if h.colorize {
		colorize = ansi.Colorize
	}

	var level string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}
	if h.recorder != nil {
		levelAttr = h.recorder([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		level = levelAttr.Value.String() + ":"

		if r.Level <= slog.LevelDebug {
			level = colorize(ansi.LightGray, level)
		} else if r.Level <= slog.LevelInfo {
			level = colorize(ansi.Cyan, level)
		} else if r.Level < slog.LevelWarn {
			level = colorize(ansi.LightBlue, level)
		} else if r.Level < slog.LevelError {
			level = colorize(ansi.LightYellow, level)
		} else if r.Level <= slog.LevelError+1 {
			level = colorize(ansi.LightRed, level)
		} else if r.Level > slog.LevelError+1 {
			level = colorize(ansi.LightMagenta, level)
		}
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(r.Time.Format(TimeFormat)),
	}
	if h.recorder != nil {
		timeAttr = h.recorder([]string{}, timeAttr)
	}
	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(ansi.LightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}
	if h.recorder != nil {
		msgAttr = h.recorder([]string{}, msgAttr)
	}
	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(ansi.White, msgAttr.Value.String())
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	var attrsAsBytes []byte
	if h.outputEmptyAttrs || len(attrs) > 0 {
		attrsAsBytes, err = json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			return fmt.Errorf("error when marshaling attrs: %w", err)
		}
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
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
		out.WriteString(colorize(ansi.DarkGray, string(attrsAsBytes)))
	}

	_, err = io.WriteString(h.writer, out.String()+"\n")
	if err != nil {
		return err
	}

	return nil
}

func SuppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
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

func NewColourHandler(handlerOptions *slog.HandlerOptions, options ...Option) *Handler {
	if handlerOptions == nil {
		handlerOptions = &slog.HandlerOptions{}
	}

	buf := &bytes.Buffer{}
	handler := &Handler{
		buf: buf,
		handler: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level:       handlerOptions.Level,
			AddSource:   handlerOptions.AddSource,
			ReplaceAttr: SuppressDefaults(handlerOptions.ReplaceAttr),
		}),
		recorder: handlerOptions.ReplaceAttr,
		mu:       &sync.Mutex{},
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

type Option func(h *Handler)

func WithDestinationWriter(writer io.Writer) Option {
	return func(h *Handler) {
		h.writer = writer
	}
}

func WithColor() Option {
	return func(h *Handler) {
		h.colorize = true
	}
}

func WithOutputEmptyAttrs(outputEmptyAttrs bool) Option {
	return func(h *Handler) {
		h.outputEmptyAttrs = outputEmptyAttrs
	}
}
