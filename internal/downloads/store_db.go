package downloads

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"imagine/internal/entities"
)

// TokenOptions configures the download token creation
type TokenOptions struct {
	TTL           time.Duration
	AllowDownload bool
	AllowEmbed    bool
	ShowMetadata  bool
	Password      string // Plain text password (will be hashed)
	Description   string
}

// CreateToken stores a random opaque token (32 bytes) in the database.
func CreateToken(db *gorm.DB, uids []string, ttl time.Duration) (string, error) {
	return CreateTokenWithOptions(db, uids, TokenOptions{
		TTL:           ttl,
		AllowDownload: true,
		AllowEmbed:    false,
		ShowMetadata:  true,
	})
}

// CreateTokenWithOptions creates a token with full configuration.
func CreateTokenWithOptions(db *gorm.DB, uids []string, opts TokenOptions) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	tok := hex.EncodeToString(b) // 64 hex characters

	var expires *time.Time
	if opts.TTL > 0 {
		t := time.Now().Add(opts.TTL)
		expires = &t
	}

	var passwordHash *string
	if opts.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.Password), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	var description *string
	if opts.Description != "" {
		description = &opts.Description
	}

	dt := entities.DownloadToken{
		Uid:           tok,
		ImageUids:     uids,
		AllowDownload: opts.AllowDownload,
		AllowEmbed:    opts.AllowEmbed,
		ShowMetadata:  opts.ShowMetadata,
		Password:      passwordHash,
		Description:   description,
		ExpiresAt:     expires,
		CreatedAt:     time.Now(),
	}

	if err := db.Create(&dt).Error; err != nil {
		return "", err
	}
	return tok, nil
}

// ValidateToken checks token validity and returns associated data.
// Returns: uids, token entity, valid flag
func ValidateToken(db *gorm.DB, token string) ([]string, bool) {
	var dt entities.DownloadToken
	if err := db.First(&dt, "token = ?", token).Error; err != nil {
		return nil, false
	}
	if dt.ExpiresAt != nil && dt.ExpiresAt.Before(time.Now()) {
		// expired, remove
		_ = db.Delete(&dt)
		return nil, false
	}
	return dt.ImageUids, true
}

// ValidateTokenWithPassword validates token and checks password if required.
// Returns: uids, token entity, valid flag
func ValidateTokenWithPassword(db *gorm.DB, token, password string) ([]string, *entities.DownloadToken, bool) {
	var dt entities.DownloadToken
	if err := db.First(&dt, "uid = ?", token).Error; err != nil {
		return nil, nil, false
	}

	// Check expiry
	if dt.ExpiresAt != nil && dt.ExpiresAt.Before(time.Now()) {
		_ = db.Delete(&dt)
		return nil, nil, false
	}

	// Check password if required
	if dt.Password != nil {
		if password == "" {
			return nil, &dt, false // Password required but not provided
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*dt.Password), []byte(password)); err != nil {
			return nil, &dt, false // Invalid password
		}
	}

	return dt.ImageUids, &dt, true
}

// ValidateEmbedAccess checks if token allows embedding based on Referer header.
// If AllowEmbed is false, only direct access (no referer) is permitted.
func ValidateEmbedAccess(dt *entities.DownloadToken, req *http.Request) bool {
	if dt.AllowEmbed {
		return true // Embedding explicitly allowed
	}

	// If embedding not allowed, check referer
	referer := req.Header.Get("Referer")
	origin := req.Header.Get("Origin")

	// Direct access (no referer/origin) is OK for downloads
	if referer == "" && origin == "" {
		return true
	}

	// Check if referer is from the same origin (our own domain)
	host := req.Host
	if strings.Contains(referer, host) || strings.Contains(origin, host) {
		return true // Same-origin requests allowed
	}

	// Third-party embedding not allowed
	return false
}
