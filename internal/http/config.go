package http

import (
	"fmt"
	"log/slog"

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
		"api":  "api-server",
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
	Logger    *slog.Logger
	Database  *db.DB
	SSEBroker *SSEBroker
}

func (server ImagineServer) ConnectToDatabase(dst ...any) *gorm.DB {
	logger := server.Logger
	database := server.Database

	client, dbError := database.Connect()
	if dbError != nil {
		logger.Error("error connecting to postgres", slog.Any("error", dbError))
		panic("")
	}

	logger.Info("Running auto-migration for auth server")
	dbError = client.AutoMigrate(dst...)
	if dbError != nil {
		logger.Error("error running auto-migration", slog.Any("error", dbError))
		panic("")
	}

	return client
}
