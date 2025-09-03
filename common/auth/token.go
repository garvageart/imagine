package auth

import (
	"fmt"
	"net/http"
	"time"
)

func GenerateAuthToken() string {
	tokenBytes := GenerateRandomBytes(48)
	return fmt.Sprintf("%s", tokenBytes)
}

func CreateAuthTokenCookie(expireTime time.Time, token string) *http.Cookie {
	return &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  expireTime,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}
