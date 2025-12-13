package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
	
	"imagine/internal/db"
	"imagine/internal/settings"
	libhttp "imagine/internal/http"
	_ "github.com/joho/godotenv/autoload"
)

type ImagineServer struct {
	*ServerConfig
	Logger   *slog.Logger
	Database *db.DB
	WSBroker *libhttp.WSBroker
	LogLevel slog.Level
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
			panic("timed out waiting for database to become available")
		}

		time.Sleep(2 * time.Second)
	}

	// Run cleanup for setting defaults before auto-migration
	settings.CleanupSettingDefaults(client, logger)
	settings.CleanupSettingOverrides(client, logger)

	dbError = client.AutoMigrate(dst...)
	if dbError != nil {
		logger.Error("error running auto-migration", slog.Any("error", dbError))
		panic("error running auto-migration: " + dbError.Error())
	}

	// Seed default settings after migration
	settings.SeedDefaultSettings(client, logger)

	// Run backfill for ownership
	db.BackfillOwnership(client, logger)

	return client
}