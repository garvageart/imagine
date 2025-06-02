package http

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/render"
)

func ServerError(res http.ResponseWriter, req *http.Request, err error, logger *slog.Logger, logArgs *slog.Attr, msg string, userMsg string) {
	jsonResponse := map[string]any{
		"user_message": "something went wrong on our side, please try again later",
		"message":      msg,
	}

	msg = strings.TrimSpace(msg)
	userMsg = strings.TrimSpace(userMsg)

	if userMsg != "" {
		jsonResponse["user_message"] = userMsg
	}

	if msg != "" {
		jsonResponse["message"] = msg
	} else {
		msg = userMsg
	}

	if msg != "" {
		jsonResponse["error"] = err.Error()
	}

	if logArgs == nil {
		logArgs = &slog.Attr{}
	}

	logger.Error(msg, *logArgs, slog.String("err", err.Error()))

	res.WriteHeader(http.StatusInternalServerError)
	render.JSON(res, req, jsonResponse)
}
