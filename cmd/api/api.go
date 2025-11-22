package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	"imagine/api/routes"
	"imagine/internal/config"
	"imagine/internal/db"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/jobs"
	"imagine/internal/jobs/workers"
	imalog "imagine/internal/logger"
	"imagine/internal/images"
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
func (server ImagineMediaServer) Launch(router *chi.Mux) *http.Server {
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
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "x-imagine-key"},
		// Expose Content-Disposition so client JS can read filenames from responses across origins
		ExposedHeaders:   []string{"Set-Cookie", "Content-Disposition"},
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
	imageops.WarmupAllOps()

	server.WSBroker = libhttp.NewWSBroker(logger)

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
		r.Mount("/events", routes.EventsRouter(dbClient, logger, server.WSBroker))
		r.Mount("/collections", routes.CollectionsRouter(dbClient, logger))
		r.Mount("/images", routes.ImagesRouter(dbClient, logger))
		r.Mount("/download", routes.DownloadRouter(dbClient, logger))
		r.Mount("/api-keys", routes.APIKeysRouter(dbClient, logger))
	})

	// Admin routes (auth + admin required)
	router.Mount("/admin", routes.AdminRouter(dbClient, logger))
	router.Mount("/jobs", routes.JobsRouter(dbClient, logger))

	address := fmt.Sprintf("%s:%d", server.Host, server.Port)

	srv := &http.Server{Addr: address, Handler: router}

	go func() {
		logger.Info(fmt.Sprintf("Hey, you want some pics? 👀 - %s: %s", ServerConfig.Key, address))

		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(fmt.Sprintf("failed to start server: %s", err))
				panic(err)
			}
		}
	}()

	// return the server so the caller can gracefully shutdown
	return srv
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
			Address: func() string {
				if host := os.Getenv("DB_HOST"); host != "" {
					return host
				}
				return "localhost"
			}(),
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

	if apiPortEnv := os.Getenv("API_PORT"); apiPortEnv != "" {
		var p int
		if _, err := fmt.Sscanf(apiPortEnv, "%d", &p); err == nil {
			server.ImagineServer.Port = p
		}
	}

	// Lmao I hate this
	client := server.ConnectToDatabase(
		entities.Image{},
		entities.Collection{},
		entities.Session{},
		entities.APIKey{},
		entities.User{},
		entities.DownloadToken{},
		entities.WorkerJob{},
		entities.UserWithPassword{},
	)
	server.ImagineServer.Database.Client = client

	srv := server.Launch(router)

	// create a cancelable context used by background tasks
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start transform cache GC if enabled in config.
	if cfg.IsSet("cache.gc_enabled") {
		if cfg.GetBool("cache.gc_enabled") {
			images.StartTransformCacheGC(ctx, logger)
		} else {
			logger.Debug("transform cache gc: disabled by config")
		}
	} else {
		// default: enabled
		images.StartTransformCacheGC(ctx, logger)
	}

	imageWorker := workers.NewImageWorker(client, server.WSBroker)
	xmpWorker := workers.NewXMPWorker(client, logger, server.WSBroker)
	exifWorker := workers.NewExifWorker(client, server.WSBroker)

	// Run the job router in a goroutine so we can wait for shutdown signals here
	go func() {
		jobs.RunJobQueue(imageWorker, xmpWorker, exifWorker)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	s := <-sigCh
	logger.Info("shutting down", slog.String("signal", s.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if srv != nil {
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", slog.Any("error", err))
		}
	}

	cancel()

	if jobs.Router != nil {
		_ = jobs.Router.Close()
	}

	time.Sleep(500 * time.Millisecond)
	logger.Info("shutdown complete")
}
