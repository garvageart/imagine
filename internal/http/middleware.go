package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/go-chi/render"

	imaAuth "imagine/internal/auth"
	"imagine/internal/dto"
	"imagine/internal/entities"
)

// Simple in-memory session->user cache to avoid DB lookups on every request.
// This is intentionally short-lived and best-effort. For multi-instance
// deployments prefer a centralized cache (Redis) to share invalidations.
type sessionCacheEntry struct {
	user      *entities.User
	expiresAt time.Time
}

var (
	sessionCacheMu sync.RWMutex
	sessionCache   = make(map[string]*sessionCacheEntry)
	// Default maximum time to cache an authenticated session locally.
	// Keep small to reduce window where revocations aren't seen.
	sessionCacheTTL = 60 * time.Second
)

// SetSessionCache stores a user for a session token. expiresAt will be the
// minimum of the session's server-side expiry (if provided) and now+TTL.
func SetSessionCache(token string, user *entities.User, serverExpires *time.Time) {
	sessionCacheMu.Lock()
	defer sessionCacheMu.Unlock()

	expiry := time.Now().Add(sessionCacheTTL)
	if serverExpires != nil && !serverExpires.IsZero() {
		if serverExpires.Before(expiry) {
			expiry = *serverExpires
		}
	}

	sessionCache[token] = &sessionCacheEntry{user: user, expiresAt: expiry}
}

// GetSessionCache returns cached user for token if present and not expired.
func GetSessionCache(token string) (*entities.User, bool) {
	sessionCacheMu.RLock()
	entry, ok := sessionCache[token]
	sessionCacheMu.RUnlock()
	if !ok || entry == nil {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		sessionCacheMu.Lock()
		delete(sessionCache, token)
		sessionCacheMu.Unlock()
		return nil, false
	}

	return entry.user, true
}

// ClearSessionCache removes a cached entry — used during logout or explicit revocation.
func ClearSessionCache(token string) {
	sessionCacheMu.Lock()
	defer sessionCacheMu.Unlock()
	delete(sessionCache, token)
}

// context keys
type ctxKey string

const (
	ctxUserKey    ctxKey = "currentUser"
	ctxAPIKey     ctxKey = "apiKey"
	ctxAPIKeyAuth ctxKey = "apiKeyAuth"
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

// APIKeyFromContext returns the authenticated API Key from the request context, if present.
func APIKeyFromContext(r *http.Request) (*entities.APIKey, bool) {
	v := r.Context().Value(ctxAPIKey)
	if v == nil {
		return nil, false
	}
	u, ok := v.(*entities.APIKey)
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
				query := db.Where("key_hashed = ?", hashed).Preload("User").First(&key)
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

				r = r.WithContext(context.WithValue(r.Context(), ctxAPIKey, &key))
				r = r.WithContext(context.WithValue(r.Context(), ctxAPIKeyAuth, true))
				next.ServeHTTP(w, r)
				return
			}

			// Fallback to cookie-based session authentication.
			cookie, err := r.Cookie(AuthTokenCookie)
			if err != nil || cookie == nil || cookie.Value == "" {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "token missing"})
				return
			}

			// Try in-memory cache first to avoid a DB roundtrip for every request.
			// If the cache misses, fall back to DB lookup and populate the cache.
			var sess entities.Session
			var userPtr *entities.User

			if u, ok := GetSessionCache(cookie.Value); ok {
				// Use cached user — attach pointer to context.
				userPtr = u
			} else {
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

				if sess.UserUid == "" {
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, dto.ErrorResponse{Error: "user missing"})
					return
				}

				var user entities.User
				if err := db.Where("uid = ?", sess.UserUid).First(&user).Error; err != nil {
					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, dto.ErrorResponse{Error: "user not found"})
					return
				}

				// Attach to request context and cache the resolved user for short time.
				userPtr = &user
				SetSessionCache(cookie.Value, userPtr, sess.ExpiresAt)
			}

			// Attach the resolved user pointer to the request context.
			if userPtr == nil {
				// Defensive: if for some reason we don't have a user, treat as unauthenticated.
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "not authenticated"})
				return
			}

			r = WithUser(r, userPtr)

			// For authenticated GET requests, expose a short-lived user-version so
			// clients can avoid re-fetching user details if they haven't changed.
			// We set ETag and an opt-in X-User-Version header, but we avoid an
			// automatic 304 here because many endpoints don't return the user
			// representation; handlers that do want to perform conditional
			// responses can compare If-None-Match themselves.
			if r.Method == http.MethodGet {
				if u, ok := UserFromContext(r); ok && u != nil {
					etag := fmt.Sprintf("W/\"%d-%s\"", u.UpdatedAt.UnixNano(), u.Uid)
					w.Header().Set("Cache-Control", "private, max-age=60, must-revalidate")
					w.Header().Set("ETag", etag)
					w.Header().Set(APIUserVersion, etag)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ScopeMiddleware requires that the request context contains an authenticated
// API Key with the required scopes. It assumes AuthMiddleware has run
// earlier in the chain to populate the API Key in context.
func ScopeMiddleware(requiredScopes []imaAuth.Scope) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Cookie-based auth should bypass scope checks
			if _, ok := UserFromContext(r); ok {
				next.ServeHTTP(w, r)
				return
			}

			apiKey, ok := APIKeyFromContext(r)
			if !ok || apiKey == nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, dto.ErrorResponse{Error: "unauthenticated"})
				return
			}

			// Superadmin bypass
			if apiKey.User != nil && (apiKey.User.Role == "admin" || apiKey.User.Role == "superadmin") {
				next.ServeHTTP(w, r)
				return
			}

			if len(apiKey.Scopes) == 0 {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, dto.ErrorResponse{Error: "forbidden: insufficient scopes"})
				return
			}

			userScopes := make(map[imaAuth.Scope]bool)
			for _, s := range apiKey.Scopes {
				userScopes[imaAuth.Scope(s)] = true
			}

			for _, requiredScope := range requiredScopes {
				if !userScopes[requiredScope] {
					render.Status(r, http.StatusForbidden)
					render.JSON(w, r, dto.ErrorResponse{Error: "forbidden: insufficient scopes"})
					return
				}
			}

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

	apiKey := r.Header.Get(APIKeyName)
	if apiKey != "" {
		return apiKey
	}

	return ""
}
