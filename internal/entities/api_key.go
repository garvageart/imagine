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
	UID       string         `json:"uid" gorm:"uniqueIndex"`
	KeyHashed string         `json:"key_hashed"`
	UserUID   string         `json:"user_uid"`
	User      *User          `json:"user" gorm:"foreignKey:UserUID;references:Uid"`
	Scopes    []string       `json:"scopes" gorm:"serializer:json;type:JSONB"`
	Revoked   bool           `json:"revoked"`
	RevokedAt *time.Time     `json:"revoked_at,omitempty"`
}
