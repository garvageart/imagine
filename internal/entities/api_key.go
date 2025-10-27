package entities

import (
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	ID        uint           `json:"-" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UID       string         `json:"uid"`
	KeyHashed string         `json:"key_hashed"`
	Scopes    []string       `json:"scopes"`
}
