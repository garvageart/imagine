package http

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"imagine/internal/db"
	libos "imagine/internal/os"
	"imagine/internal/utils"

	_ "github.com/joho/godotenv/autoload"
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

		config, err := ReadConfig()
		if err != nil {
			panic("Unable to read config file")
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
		logger.Error("error running auto-migration for auth server", slog.Any("error", dbError))
		panic("")
	}

	return client
}

func ReadConfig() (viper.Viper, error) {
	configPath := libos.CurrentWorkingDirectory + "/config"

	viper.SetConfigName(utils.AppName)
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.Viper{}, fmt.Errorf("can't find config file: %w", err)
		}
		// For other errors when reading the config
		return viper.Viper{}, fmt.Errorf("error reading config file: %w", err)
	}

	return *viper.GetViper(), nil
}
