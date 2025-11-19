package http

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"imagine/internal/config"
	"imagine/internal/db"
	"imagine/internal/utils"

	_ "github.com/joho/godotenv/autoload"
)

const (
	APIKeyName = "x-imagine-key"
)

var (
	ServerKeys = map[string]string{
		"auth": "auth-server",
		"api":  "api",
		"viz":  "viz",
	}

	ImagineServers = func() map[string]*ImagineServer {
		var host string
		if utils.IsProduction {
			host = "0.0.0.0"
		} else {
			host = "localhost"
		}

		config, err := config.ReadConfig()
		if err != nil {
			panic("Unable to read config file " + err.Error())
		}

		result := map[string]*ImagineServer{}
		for _, serverKey := range ServerKeys {
			result[serverKey] = &ImagineServer{Server: &Server{}}

			result[serverKey].Port = config.GetInt(fmt.Sprintf("servers.%s.port", serverKey))
			result[serverKey].Host = host
			result[serverKey].Key = serverKey
		}

		return result
	}()
)

type Server struct {
	Port int
	Host string
	Key  string
}

type ImagineServer struct {
	*Server
	Logger   *slog.Logger
	Database *db.DB
	WSBroker *WSBroker
}

func (server ImagineServer) ConnectToDatabase(dst ...any) *gorm.DB {
	logger := server.Logger
	database := server.Database

	timeoutSeconds := 60
	if v := os.Getenv("DB_CONNECT_TIMEOUT"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	start := time.Now()
	var client *gorm.DB
	var dbError error
	for {
		client, dbError = database.Connect()
		if dbError == nil {
			break
		}

		logger.Error("error connecting to postgres, will retry", slog.Any("error", dbError))

		if time.Since(start) > time.Duration(timeoutSeconds)*time.Second {
			logger.Error("timed out waiting for database to become available", slog.Int("timeout_seconds", timeoutSeconds))
			panic("")
		}

		time.Sleep(2 * time.Second)
	}

	logger.Info("Running auto-migration for auth server")
	dbError = client.AutoMigrate(dst...)
	if dbError != nil {
		logger.Error("error running auto-migration", slog.Any("error", dbError))
		panic("")
	}

	return client
}
