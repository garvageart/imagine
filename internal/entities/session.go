package entities

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID         uint           `json:"-" gorm:"primarykey"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	Token      string         `json:"token"`
	UID        string         `json:"uid"`
	UserUID    string         `json:"user_id"`
	ClientID   string         `json:"client_id"`
	ClientName string         `json:"client_name"`
	ClientIP   string         `json:"client_ip"`
	LastActive time.Time      `json:"last_active"`
	ExpiresAt  time.Time      `json:"expires_at"`
	Timeout    time.Duration  `json:"timeout"`
	UserAgent  string         `json:"user_agent"`
	RefID      string         `json:"ref_id"`
	LoginIP    string         `json:"login_ip"`
	LoginAt    time.Time      `json:"login_at"`
	Status     int            `json:"status"`
}
