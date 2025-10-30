package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

var (
	ErrTokenInvalid             = CreateDefaultMessage("invalid token", 1000)
	ErrTokenExpired             = CreateDefaultMessage("token expired", 1001)
	ErrTokenMissing             = CreateDefaultMessage("token missing", 1002)
	ErrRequestBodyInvalid       = CreateDefaultMessage("invalid request body", 1003)
	ErrSomethingWentWrongServer = CreateDefaultMessage("something went wrong", 1004)
)

type DefaultMessage struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func CreateDefaultMessage(message string, code int) DefaultMessage {
	return DefaultMessage{
		Message: message,
		Code:    code,
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    int    `json:"code"`
}

func CreateErrorResponse(message string, status int, code int) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Status:  status,
		Code:    code,
	}
}

// Logs the error to all sources and creates an HTTP 500 response
// if there is a user message to send. HTTP responses will be sent in JSON
func ServerError(res http.ResponseWriter, req *http.Request, err error, logger *slog.Logger, logArgs []slog.Attr, msg string, userMsg string) {
	jsonResponse := map[string]any{
		"user_message": "Something went wrong on our side, please try again later",
		"message":      msg,
	}

	msg = strings.TrimSpace(msg)
	userMsg = strings.TrimSpace(userMsg)

	if logArgs == nil {
		logArgs = []slog.Attr{}
	}

	// Get the request ID
	logArgs = append(logArgs, slog.String("request_id", middleware.GetReqID(req.Context())))

	// Add the user message to the response if it's not blank
	if userMsg != "" {
		jsonResponse["user_message"] = userMsg
		logArgs = append(logArgs, slog.String("user_message", userMsg))
	}

	// If the server-specific error message is blank, return
	// the error as a string
	if msg != "" {
		jsonResponse["message"] = msg
		logArgs = append(logArgs, slog.String("message", msg))
	} else {
		msg = err.Error()
	}

	logger.Error(msg, slog.Any("args", logArgs))

	if userMsg == "" {
		return
	}

	// Usually theres no other kind of http error code for stuff breaking on the server
	res.WriteHeader(http.StatusInternalServerError)
	render.JSON(res, req, jsonResponse)
}

// ServerErrorOptions configures the ServerError response
type ServerErrorOptions struct {
	// Err is the original error that occurred
	Err error

	// InternalMsg is the detailed error message for logging (not shown to users)
	InternalMsg string

	// UserMsg is the user-facing error message. If empty, uses a default message.
	UserMsg string

	// LogAttrs are additional structured logging attributes
	LogAttrs []slog.Attr

	// StatusCode allows overriding the default 500 status
	StatusCode int
}

const defaultUserMessage = "Something went wrong on our side, please try again later"

// ServerError logs an error and sends an HTTP error response.
// Always sends an HTTP response with at least a default user message.
func ServerErrorNew(w http.ResponseWriter, r *http.Request, logger *slog.Logger, opts ServerErrorOptions) {
	if opts.StatusCode == 0 {
		opts.StatusCode = http.StatusInternalServerError
	}

	logMsg := strings.TrimSpace(opts.InternalMsg)
	if logMsg == "" && opts.Err != nil {
		logMsg = opts.Err.Error()
	}

	if logMsg == "" {
		logMsg = "unknown error"
	}

	userMsg := strings.TrimSpace(opts.UserMsg)
	if userMsg == "" {
		userMsg = defaultUserMessage
	}

	logAttrs := make([]slog.Attr, 0, len(opts.LogAttrs)+3)
	logAttrs = append(logAttrs, opts.LogAttrs...)
	logAttrs = append(logAttrs,
		slog.String("request_id", middleware.GetReqID(r.Context())),
		slog.String("user_message", userMsg),
	)

	if opts.Err != nil {
		logAttrs = append(logAttrs, slog.Any("error", opts.Err))
	}

	logger.LogAttrs(r.Context(), slog.LevelError, logMsg, logAttrs...)

	jsonResponse := map[string]any{
		"user_message": userMsg,
		"message":      logMsg,
	}

	w.WriteHeader(opts.StatusCode)
	render.JSON(w, r, jsonResponse)
}

// Convenience wrapper for common case
func ServerErrorSimple(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error, userMsg string) {
	ServerErrorNew(w, r, logger, ServerErrorOptions{
		Err:     err,
		UserMsg: userMsg,
	})
}
