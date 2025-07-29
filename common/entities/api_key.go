package entities

import "gorm.io/gorm"

type APIKey struct {
	gorm.Model
	UID       string   `json:"uid"`
	KeyHashed string   `json:"key_hashed"`
	Scopes    []string `json:"scopes"`
}
