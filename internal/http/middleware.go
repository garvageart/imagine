package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/go-chi/render"

	imaAuth "imagine/internal/auth"
	"imagine/internal/dto"
	"imagine/internal/entities"
)

// context keys
type ctxKey string

const (
	ctxUserKey ctxKey = "currentUser"
	ctxAPIKey  ctxKey = "apiKeyAuth"
)

// WithUser returns a request with the authenticated user added to the context.
func WithUser(r *http.Request, user *entities.User) *http.Request {
	ctx := context.WithValue(r.Context(), ctxUserKey, user)
	return r.WithContext(ctx)
}

// UserFromContext returns the authenticated user from the request context, if present.
func UserFromContext(r *http.Request) (*entities.User, bool) {
	v := r.Context().Value(ctxUserKey)
	if v == nil {
		return nil, false
	}
	u, ok := v.(*entities.User)
	return u, ok
}

// AuthMiddleware validates the auth cookie against the sessions table, loads the user,
// and injects it into the request context. 401 is returned for missing/invalid/expired sessions.
func AuthMiddleware(db *gorm.DB, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := getAPIKeyFromRequest(r)
			if apiKey != "" {
				hashed, _ := imaAuth.HashSecret(apiKey)

				var key entities.APIKey
				query := db.Where("key_hashed = ?", hashed).First(&key)
				if query.Error != nil {
					if query.Error == gorm.ErrRecordNotFound {
						render.Status(r, http.StatusUnauthorized)
						render.JSON(w, r, dto.ErrorResponse{Error: "invalid api key"})
						return
					}

					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, dto.ErrorResponse{Error: "invalid api key"})
					return
				}

				if key.Revoked {
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, dto.ErrorResponse{Error: "api key has been revoked"})
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), ctxAPIKey, true))
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(AuthTokenCookie)
			if err != nil || cookie == nil || cookie.Value == "" {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "token missing"})
				return
			}

			var sess entities.Session
			if err := db.Where("token = ?", cookie.Value).First(&sess).Error; err != nil {
				// clear auth cookie
				ClearCookie(AuthTokenCookie, w)
				ClearCookie(StateCookie, w)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "invalid session"})
				return
			}

			if sess.ExpiresAt != nil && !sess.ExpiresAt.IsZero() && time.Now().After(*sess.ExpiresAt) {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "session expired"})
				return
			}

			var user entities.User
			if sess.UserUid == "" {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "user missing"})
				return
			}

			err = db.Where("uid = ?", sess.UserUid).First(&user).Error

			if err != nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "user not found"})
				return
			}

			r = WithUser(r, &user)
			next.ServeHTTP(w, r)
		})
	}
}

// AdminMiddleware requires that the request context contains an authenticated
// user with role "admin" or "superadmin". It assumes AuthMiddleware has run
// earlier in the chain to populate the user in context.
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r)
		if !ok || user == nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, dto.ErrorResponse{Error: "unauthenticated"})
			return
		}

		if user.Role != "admin" && user.Role != "superadmin" {
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, dto.ErrorResponse{Error: "forbidden"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getAPIKeyFromRequest checks common header locations for an API key
func getAPIKeyFromRequest(r *http.Request) string {
	// Prefer Authorization: Bearer <key>
	authHead := r.Header.Get("Authorization")
	if authHead != "" {
		// Case-insensitive prefix match
		if len(authHead) > 7 && (authHead[:7] == "Bearer " || authHead[:7] == "bearer ") {
			return authHead[7:]
		}
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.Header.Get(APIKeyName)
	}
	if apiKey != "" {
		return apiKey
	}

	return ""
}
