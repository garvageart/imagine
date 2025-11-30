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
	"imagine/internal/auth"
	"imagine/internal/config"
	"imagine/internal/db"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/jobs/workers"
	imalog "imagine/internal/logger"
	"imagine/internal/utils"
)

var (
	ServerConfig = config.ImagineServers["api"]
)

type APIServer struct {
	*config.ImagineServer
}

// TODO: Create a `createServer/Router` function that returns a router
// with common defaults for each server type
func (server APIServer) Launch(router *chi.Mux) *http.Server {
	logLevel := server.LogLevel
	logger := server.Logger
	serverLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)

	// Setup general middleware - CORS must be first!
	router.Use(cors.Handler(cors.Options{
		// TODO: maybe make this configurable by admin since this might
		// some people might not want to allow all origins for API
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", libhttp.APIKeyName, "If-None-Match", "If-Modified-Since"},
		// Expose Content-Disposition so client JS can read filenames from responses across origins
		ExposedHeaders:   []string{"Set-Cookie", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: serverLogger,
	}))
	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.GetHead)
	router.Use(middleware.Timeout(60 * time.Second))
	// Note: AuthMiddleware is applied per-route, not globally

	database := server.Database
	dbClient := database.Client

	var libvipsLogLevel libvips.LogLevel = libvips.LogLevelInfo
	var matchServerLogLevel = os.Getenv("LIBVIPS_MATCH_LOG_LEVEL") == "true"
	if matchServerLogLevel {
		switch logLevel {
		case slog.LevelDebug:
			libvipsLogLevel = libvips.LogLevelDebug
		case slog.LevelInfo:
			libvipsLogLevel = libvips.LogLevelInfo
		case slog.LevelWarn:
			libvipsLogLevel = libvips.LogLevelWarning
		case slog.LevelError:
			libvipsLogLevel = libvips.LogLevelError
		default:
			libvipsLogLevel = libvips.LogLevelInfo
		}
	} else {
		// TODO: fix this error message, it sucks and is confusing
		logger.Info("libvipsLogLevel: matching server level is off. using default: info")
	}

	var libvipsLogHandler libvips.LoggingHandlerFunction = func(messageDomain string, messageLevel libvips.LogLevel, message string) {
		switch messageLevel {
		case libvips.LogLevelCritical:
			imalog.Fatal(logger, fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelError:
			logger.Error(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelWarning:
			logger.Warn(fmt.Sprintf("%s: %s", messageDomain, message))
		case libvips.LogLevelMessage, libvips.LogLevelInfo:
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
		r.Group(func(r chi.Router) {
			r.Use(libhttp.ScopeMiddleware([]auth.Scope{auth.EventsReadScope}))
			r.Mount("/events", routes.EventsRouter(dbClient, logger, server.WSBroker))
		})
		r.Group(func(r chi.Router) {
			r.Use(libhttp.ScopeMiddleware([]auth.Scope{
				auth.CollectionsCreateScope, 
				auth.CollectionsDeleteScope, 
				auth.CollectionsReadScope, 
				auth.CollectionsUpdateScope,
			}))
			r.Mount("/collections", routes.CollectionsRouter(dbClient, logger))
		})
		r.Group(func(r chi.Router) {
			r.Use(libhttp.ScopeMiddleware([]auth.Scope{
				auth.ImagesReadScope, 
				auth.ImagesDownloadScope, 
				auth.ImagesDeleteScope, 
				auth.ImagesUpdateScope, 
				auth.ImagesUploadScope,
			}))
			r.Mount("/images", routes.ImagesRouter(dbClient, logger))
		})

		r.Group(func(r chi.Router) {
			r.Use(libhttp.ScopeMiddleware([]auth.Scope{
				auth.DownloadsCreateScope,
			}))
			r.Mount("/download", routes.DownloadRouter(dbClient, logger))
		})
		r.Group(func(r chi.Router) {
			r.Use(libhttp.ScopeMiddleware([]auth.Scope{
				auth.APIKeysReadScope, 
				auth.APIKeysCreateScope,
				auth.APIKeysRevokeScope,
				auth.APIKeysListScope,
				auth.APIKeysRotateScope,
				auth.APIKeysDeleteScope,
			}))
			r.Mount("/api-keys", routes.APIKeysRouter(dbClient, logger))
		})
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

	v, err := config.ReadConfig()
	if err != nil {
		errorMsg := fmt.Sprintf("failed to read config file: %v", err)
		panic(errorMsg)
	}

	var appConfig config.ImagineConfig
	if err := v.Unmarshal(&appConfig); err != nil {
		errorMsg := fmt.Sprintf("failed to unmarshal config: %v", err)
		panic(errorMsg)
	}

	// setup logging stuff
	logLevel := imalog.GetLevelFromString(appConfig.Logging.Level)
	logger := libhttp.SetupChiLogger("api", logLevel)

	apiServer := APIServer{ImagineServer: ServerConfig}
	apiServer.ImagineServer.LogLevel = logLevel
	apiServer.ImagineServer.Logger = logger

	// db stuff
	if os.Getenv("DB_PASSWORD") != "" {
		appConfig.Database.Password = os.Getenv("DB_PASSWORD")
	} else {
		logger.Error("Database password not set, please set a password")
		panic("Database password not set")
	}

	apiServer.Database = &db.DB{
		Address: func() string {
			if host := os.Getenv("DB_HOST"); host != "" {
				return host
			}
			return "localhost"
		}(),
		Port: func() int {
			if appConfig.Database.Port == 0 {
				return 5432
			}
			return appConfig.Database.Port
		}(),
		User:     appConfig.Database.User,
		Password: appConfig.Database.Password,
		AppName:  utils.AppName,
		DatabaseName: func() string {
			dbName := appConfig.Database.Name
			if dbName != "" {
				// TODO: this nonsense will change in future
				if utils.IsDevelopment {
					return dbName + "-dev"
				}
				return dbName
			}

			return "imagine"
		}(),
		Logger:   logger,
		LogLevel: logLevel,
	}

	// Lmao I hate this
	client := apiServer.ConnectToDatabase(
		entities.Image{},
		entities.Collection{},
		entities.Session{},
		entities.APIKey{},
		entities.User{},
		entities.DownloadToken{},
		entities.WorkerJob{},
		entities.UserWithPassword{},
	)
	apiServer.ImagineServer.Database.Client = client

	// http server stuff
	if apiPortEnv := os.Getenv("API_PORT"); apiPortEnv != "" {
		var p int
		if _, err := fmt.Sscanf(apiPortEnv, "%d", &p); err == nil {
			apiServer.ImagineServer.Port = p
		}
	}
	httpServer := apiServer.Launch(router)

	// create a cancelable context used by background tasks
	ctx, globalCancel := context.WithCancel(context.Background())
	defer globalCancel()

	// Start transform cache GC if enabled in config.
	if appConfig.Cache.GCEnabled {
		images.StartTransformCacheGC(ctx, logger)
	} else {
		logger.Debug("transform cache gc: disabled by config")
	}

	imageWorker := workers.NewImageWorker(client, apiServer.WSBroker)
	xmpWorker := workers.NewXMPWorker(client, apiServer.WSBroker)
	exifWorker := workers.NewExifWorker(client, apiServer.WSBroker)

	// Run the job router in a goroutine so we can wait for shutdown signals here
	go func() {
		jobs.RunJobQueue(appConfig.Queue, logger, imageWorker, xmpWorker, exifWorker)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	s := <-sigCh
	logger.Info("shutting down", slog.String("signal", s.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", slog.Any("error", err))
		}
	}

	globalCancel()

	if jobs.Router != nil {
		_ = jobs.Router.Close()
	}

	time.Sleep(500 * time.Millisecond)
	logger.Info("shutdown complete")
}
