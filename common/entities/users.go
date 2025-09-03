package entities

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	CreatedAt     time.Time `json:"created_at"`
	UID           string    `json:"uid"`
	FirstName     *string   `json:"first_name"`
	LastName      *string   `json:"last_name"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Password      string    `json:"password"`
	UsedOAuth     *bool     `json:"used_oauth"`
	OAuthProvider *string   `json:"oauth_provider"`
	OAuthState    *string   `json:"oauth_state"`
}

func (User) TableName() string {
	return "users"
}
