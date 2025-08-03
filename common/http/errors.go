package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

var (
	ErrTokenInvalid = CreateDefaultMessage("invalid token", 1000)
	ErrTokenExpired = CreateDefaultMessage("token expired", 1001)
	ErrTokenMissing = CreateDefaultMessage("token missing", 1002)
	ErrRequestBodyInvalid = CreateDefaultMessage("invalid request body", 1003)
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
