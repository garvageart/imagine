package entities

import (
	"time"

	"gorm.io/gorm"
)

type CollectionImage struct {
	UID     string    `json:"uid"`
	AddedAt time.Time `json:"added_at"`
	AddedBy string    `json:"added_by"`
}

type Collection struct {
	ID          uint              `json:"-" gorm:"primarykey"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `json:"-" gorm:"index"`
	UID         string            `json:"uid"`
	Name        string            `json:"name"`
	ImageCount  int               `json:"image_count"`
	Private     *bool             `json:"private"`
	Images      []CollectionImage `json:"images" gorm:"type:JSONB"`
	CreatedBy   string            `json:"created_by"`
	Description string            `json:"description"`
}
