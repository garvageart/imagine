package auth

import (
	"imagine/internal/utils"
	"net/http"
	"time"
)

func GenerateAuthToken() string {
	tokenBytes := utils.GenerateRandomBytes(48)
	return string(tokenBytes)
}

func CreateAuthTokenCookie(expireTime time.Time, token string) *http.Cookie {
	return &http.Cookie{
		Name:     "imag-auth_token",
		Value:    token,
		Expires:  expireTime,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}
