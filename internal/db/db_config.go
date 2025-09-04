package db

import (
	"context"
	"log/slog"

	"gorm.io/gorm"
)

type DB struct {
	Address         string
	Protocol        string
	Port            int
	User            string
	Password        string
	DatabaseName    string
	Client          *gorm.DB
	AppName         string
	TableNameString string
	Table           *gorm.DB
	Context         context.Context
	Logger          *slog.Logger
}
