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
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"imagine/api/routes"
	"imagine/internal/db"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/utils"
)

var (
	ServerConfig = libhttp.ImagineServers["api"]
)

type ImagineMediaServer struct {
	*libhttp.ImagineServer
}

// TODO: This will be the main API server and therefore will have a lot of routes.
// This file and directory will be renamed to "api" and the parent directory to "servers" :)
// Split the different routes into their own files depending on what they server
// For example, a /user* route for the user data etc.

// TODO TODO: Create a `createServer/Router` function that returns a router
// with common defaults for each server type
func (server ImagineMediaServer) Launch(router *chi.Mux) {
	logger := server.Logger

	serverLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: serverLogger,
	}))

	// Setup general middleware
	// router.Use(libhttp.AuthedMiddleware)
	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)
	router.Use(cors.Handler(cors.Options{
		// TODO: Replace with config addresses instead of the hardcoded values
		AllowedOrigins:   []string{"https://localhost:7777", "http://localhost:7777"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "set-cookie"},
		AllowCredentials: true,
	}))

	database := server.Database
	dbClient := database.Client

	var libvipsLogLevel libvips.LogLevel = libvips.LogLevelInfo
	if os.Getenv("LIBVIPS_LOG_LEVEL") != "" {
		switch os.Getenv("LIBVIPS_LOG_LEVEL") {
		case "critical":
			libvipsLogLevel = libvips.LogLevelCritical
		case "error":
			libvipsLogLevel = libvips.LogLevelError
		case "warning":
			libvipsLogLevel = libvips.LogLevelWarning
		case "message":
			libvipsLogLevel = libvips.LogLevelMessage
		case "info":
			libvipsLogLevel = libvips.LogLevelInfo
		case "debug":
			libvipsLogLevel = libvips.LogLevelDebug
		}
	}

	libvips.LoggingSettings(func(messageDomain string, messageLevel libvips.LogLevel, message string) {
		switch messageLevel {
		case libvips.LogLevelCritical:
			logger.Error(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelError:
			logger.Error(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelWarning:
			logger.Warn(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelMessage:
			logger.Info(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelInfo:
			logger.Info(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelDebug:
			logger.Debug(fmt.Sprintf("%s: %s", messageDomain, message))
		}
	}, libvipsLogLevel)

	// TODO: Migrate to https://github.com/cshum/vipsgen
	libvips.Startup(&libvips.Config{
		ConcurrencyLevel: 4,
		MaxCacheFiles:    100,
		MaxCacheMem:      500 * 1024 * 1024,
		MaxCacheSize:     1000,
		ReportLeaks:      true,
	})
	defer libvips.Shutdown()

	// Mount image router to main router
	router.Mount("/collections", routes.CollectionsRouter(dbClient, logger))
	router.Mount("/images", routes.ImagesRouter(dbClient, logger))
	router.Mount("/accounts", routes.AccountsRouter(dbClient, logger))

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "pong"}
		render.JSON(res, req, jsonResponse)
	})

	// TODO: only admin can do a healthcheck
	router.Post("/healthcheck", func(res http.ResponseWriter, req *http.Request) {
		result := dbClient.Exec("SELECT 1")
		if result.Error != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": "healthcheck failed"})
			return
		}

		randomPositiveMessage := []string{
			"all love and peace ",
			"take care of yourself",
			"love is in the air",
			"support open source <3",
		}

		loveMessage := randomPositiveMessage[utils.RandomInt(0, len(randomPositiveMessage)-1)]

		res.WriteHeader(http.StatusOK)
		render.JSON(res, req, map[string]string{"message": "ok", "status": loveMessage})
	})

	address := fmt.Sprintf("%s:%d", server.Host, server.Port)

	go func() {
		logger.Info(fmt.Sprintf("Hey, you want some pics? 👀 - %s: %s", ServerConfig.Key, address))

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
	logger := libhttp.SetupChiLogger(ServerConfig.Key)

	server := ImagineMediaServer{ImagineServer: ServerConfig}
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
