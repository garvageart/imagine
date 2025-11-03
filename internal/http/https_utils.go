package http

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

const (
	AuthTokenCookie    = "imag-auth_token"
	StateCookie        = "imag-state"
	RedirectCookie     = "imag-redirect_state"
	RefreshTokenCookie = "imag-refresh_token"
)

func GetRequestID(request *http.Request) string {
	return middleware.GetReqID(request.Context())
}

func ClearCookie(name string, w http.ResponseWriter) {
	// Respect configured cookie domain/secure/samesite settings when clearing
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
	}

	// Optional domain for cross-subdomain cookies
	if d := os.Getenv("IMAGINE_COOKIE_DOMAIN"); d != "" {
		cookie.Domain = d
	}

	// Secure must be true for SameSite=None to work in browsers
	if os.Getenv("IMAGINE_COOKIE_SECURE") == "true" {
		cookie.Secure = true
	}

	// Allow configuring SameSite via IMAGINE_COOKIE_SAMESITE
	switch os.Getenv("IMAGINE_COOKIE_SAMESITE") {
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	default:
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

func CreateAuthTokenCookie(expireTime time.Time, token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     AuthTokenCookie,
		Value:    token,
		Expires:  expireTime,
		HttpOnly: true,
		Path:     "/",
	}

	// Optional domain for cross-subdomain cookies
	if d := os.Getenv("IMAGINE_COOKIE_DOMAIN"); d != "" {
		cookie.Domain = d
	}

	// Secure must be true for SameSite=None to work in browsers
	if os.Getenv("IMAGINE_COOKIE_SECURE") == "true" {
		cookie.Secure = true
	}

	// Allow configuring SameSite via IMAGINE_COOKIE_SAMESITE
	switch os.Getenv("IMAGINE_COOKIE_SAMESITE") {
	case "None":
		cookie.SameSite = http.SameSiteNoneMode
	case "Strict":
		cookie.SameSite = http.SameSiteStrictMode
	default:
		cookie.SameSite = http.SameSiteLaxMode
	}

	return cookie
}
