package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"imagine/api/routes"
	"imagine/internal/config"
	"imagine/internal/db"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/jobs"
	"imagine/internal/jobs/workers"
	imalog "imagine/internal/logger"
	"imagine/internal/utils"
)

var (
	ServerConfig = libhttp.ImagineServers["api-server"]
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

	// Setup general middleware - CORS must be first!
	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "x-imagine-key"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)
	// Note: AuthMiddleware is applied per-route, not globally

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

	var libvipsLogHandler libvips.LoggingHandlerFunction = func(messageDomain string, messageLevel libvips.LogLevel, message string) {
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
	}

	libvips.SetLogging(libvipsLogHandler, libvipsLogLevel)

	// Public routes (no auth required)
	router.Mount("/auth", routes.AuthRouter(dbClient, logger))
	router.Mount("/accounts", routes.AccountsRouter(dbClient, logger))
	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "pong"}
		render.JSON(res, req, jsonResponse)
	})

	// Protected routes (auth required)
	router.Group(func(r chi.Router) {
		r.Use(libhttp.AuthMiddleware(server.Database.Client, logger))
		r.Mount("/collections", routes.CollectionsRouter(dbClient, logger))
		r.Mount("/images", routes.ImagesRouter(dbClient, logger))
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

			panic(err)
		}
	}()
}

func main() {
	router := chi.NewRouter()
	logger := imalog.CreateDefaultLogger()

	cfg, err := config.ReadConfig()
	if err != nil {
		errorMsg := fmt.Sprintf("failed to read config file: %v", err)
		logger.Error(errorMsg, slog.String("error", err.Error()))
		panic(errorMsg)
	}

	server := ImagineMediaServer{ImagineServer: ServerConfig}
	server.ImagineServer.Logger = logger
	server.Database = &db.DB{
		Address: "localhost",
		Port: func() int {
			if os.Getenv("DB_PORT") != "" {
				var port int
				if cfgPort := cfg.GetInt("database.port"); cfgPort != 0 {
					port = cfgPort
				} else if envPort := os.Getenv("DB_PORT"); envPort != "" {
					fmt.Sscanf(envPort, "%d", &port)
				}

				return port
			}

			return 5432
		}(),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		AppName:  utils.AppName,
		DatabaseName: func() string {
			if os.Getenv("DB_NAME") != "" {
				return os.Getenv("DB_NAME")
			}

			dbName := cfg.GetString("database.name")
			if dbName != "" {
				if utils.IsDevelopment {
					return dbName + "-dev"
				}
				return dbName
			}

			return "imagine"
		}(),
		Logger: logger,
	}

	// Lmao I hate this
	client := server.ConnectToDatabase(entities.Image{}, entities.Collection{}, entities.Session{}, entities.User{})
	server.ImagineServer.Database.Client = client

	server.Launch(router)
	jobWorkers := workers.NewImageWorker(client)
	jobs.RunJobQueue(jobWorkers)
}
