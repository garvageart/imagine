package entities

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UID           string  `json:"uid"`
	FirstName     *string `json:"first_name"`
	LastName      *string `json:"last_name"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	UsedOAuth     *bool   `json:"used_oauth"`
	OAuthProvider *string `json:"oauth_provider"`
	OAuthState    *string `json:"oauth_state"`
	UserToken     *string `json:"user_token"`
}

func (User) TableName() string {
	return "users"
}

type Session struct {
	gorm.Model
	UID          string `json:"uid"`
	SessionToken string `json:"session_token"`
	ExpiresAt    int64  `json:"expires_at"`
	RevokedAt    int64  `json:"revoked_at"`
	Revoked      bool   `json:"revoked"`
}
