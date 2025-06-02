package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	gcp "imagine/common/gcp/storage"
	libhttp "imagine/common/http"
)

type ImagineMediaServer struct {
	*libhttp.ImagineServer
}

func (server ImagineMediaServer) setupImageRouter() *chi.Mux {
	imageRouter := chi.NewRouter()
	logger := server.Logger

	gcsContext, gcsContextCancel := context.WithCancel(context.Background())
	defer gcsContextCancel()

	storageClient, err := gcp.SetupClient(gcsContext)
	if err != nil {
		panic("Failed to setup GCP Storage client" + err.Error())
	}

	imageRouter.Get("/download", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotImplemented)
		res.Header().Add("Content-Type", "text/plain")
		res.Write([]byte("not implemented"))
	})

	imageRouter.Get("/upload", func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotImplemented)
		res.Header().Add("Content-Type", "text/plain")
		res.Write([]byte("not implemented"))
	})

	return imageRouter
}

func (server ImagineMediaServer) Launch(router *chi.Mux) {
	imageRouter := server.setupImageRouter()
	logger := server.Logger

	correctLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
	}))

	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)

	// Mount image router to main router
	router.Mount("/image", imageRouter)

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "pong"}
		render.JSON(res, req, jsonResponse)
	})

	logger.Info(fmt.Sprint("Starting server on port ", server.Port))
	address := fmt.Sprintf("%s:%d", server.Host, server.Port)
	err := http.ListenAndServe(address, router)
	// This line will only be reached if ListenAndServe returns, usually with an error.
	logger.Error(fmt.Sprintf("failed to start server: %s", err.Error()))
}

func main() {
	key := "media-server"
	router := chi.NewRouter()
	logger := libhttp.SetupChiLogger(key)

	var server = &ImagineMediaServer{
		ImagineServer: &libhttp.ImagineServer{
			Host:   "localhost",
			Key:    key,
			Logger: logger,
		}}

	config, err := server.ReadConfig()
	if err != nil {
		panic("Unable to read config file: " + err.Error())
	}
	server.Port = config.GetInt(fmt.Sprintf("servers.%s.port", server.Key))
	server.Launch(router)
}
