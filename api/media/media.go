package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	libvips "github.com/davidbyttow/govips/v2/vips"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"imagine/common/entities"
	libhttp "imagine/common/http"
	"imagine/db"
	"imagine/utils"
)

const (
	serverKey = "media-server"
)

type ImagineMediaServer struct {
	*libhttp.ImagineServer
}

// func (server ImagineMediaServer) setupImageRouter() *chi.Mux {
// 	imageRouter := chi.NewRouter()

// 	logger := server.Logger

// 	gcsContext, gcsContextCancel := context.WithCancel(context.Background())
// 	defer gcsContextCancel()

// 	storageClient, err := gcp.SetupClient(gcsContext)
// 	if err != nil {
// 		panic("Failed to setup GCP Storage client" + err.Error())
// 	}

// 	imageRouter.Get("/download", func(res http.ResponseWriter, req *http.Request) {
// 		res.WriteHeader(http.StatusNotImplemented)
// 		res.Header().Add("Content-Type", "text/plain")
// 		res.Write([]byte("not implemented"))
// 	})

// 	imageRouter.Get("/upload", func(res http.ResponseWriter, req *http.Request) {
// 		res.WriteHeader(http.StatusNotImplemented)
// 		res.Header().Add("Content-Type", "text/plain")
// 		res.Write([]byte("not implemented"))
// 	})

// 	return imageRouter
// }

// TODO: This will be the main API server and therefore will have a lot of routes.
// This file and directory will be renamed to "api" and the parent directory to "servers" :)
// Split the different routes into their own files depending on what they server
// For example, a /user* route for the user data etc.

// TODO TODO: Create a `createServer/Router` function that returns a router
// with common defaults for each server type

func (server ImagineMediaServer) Launch(router *chi.Mux) {
	// imageRouter := server.setupImageRouter()
	logger := server.Logger

	correctLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
	}))

	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)

	libvips.Startup(nil)
	defer libvips.Shutdown()

	database := server.Database
	client := database.Client

	// Mount image router to main router
	router.Mount("/collections", CollectionsRouter(client, logger))
	router.Mount("/images", ImagesRouter(client, logger))

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "pong"}
		render.JSON(res, req, jsonResponse)
	})

	// TODO: only admin can do a healthcheck
	router.Post("/healthcheck", func(res http.ResponseWriter, req *http.Request) {
		result := client.Exec("SELECT 1")
		if result.Error != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": "healthcheck failed"})
			return
		}

		res.WriteHeader(http.StatusOK)
		render.JSON(res, req, map[string]string{"message": "ok", "status": "all love and peace 🤍"})
	})

	address := fmt.Sprintf("%s:%d", server.Host, server.Port)

	go func() {
		logger.Info(fmt.Sprintf("Hey, you want some pics? 👀 - %s: %s", serverKey, address))

		err := http.ListenAndServe(address, router)
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(fmt.Sprintf("failed to start server: %s", err))
			}

			panic("")
		}
	}()

	// Taken and adjusted from https://github.com/bluesky-social/social-app/blob/main/bskyweb/cmd/bskyweb/server.go
	// Wait for a signal to exit.
	logger.Info("registering OS exit signal handler")
	quit := make(chan struct{})
	exitSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-exitSignals
		logger.Info(fmt.Sprintf("received OS exit signal: %s", sig))

		// Trigger the return that causes an exit.
		close(quit)
	}()
	<-quit
	logger.Info("graceful shutdown complete")
}

func main() {
	router := chi.NewRouter()
	logger := libhttp.SetupChiLogger(serverKey)

	server := ImagineMediaServer{ImagineServer: libhttp.ImagineServers[serverKey]}
	server.ImagineServer.Logger = logger
	server.Database = &db.DB{
		Address:      "localhost",
		Port:         5432,
		User:         os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		AppName:      utils.AppName,
		DatabaseName: "imagine-dev",
		Logger:       logger,
	}

	// Lmao I hate this
	client := server.ConnectToDatabase(entities.Image{}, entities.Collection{})
	server.ImagineServer.Database.Client = client

	server.Launch(router)
}
