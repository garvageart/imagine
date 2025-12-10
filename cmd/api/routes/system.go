package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/config"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/settings"
)

// SystemRouter creates a router for system-related endpoints
func SystemRouter(db *gorm.DB, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Use(systemCacheMiddleware)

	// This is lowkey complicated and a mess but whatever
	r.Get("/status", func(res http.ResponseWriter, req *http.Request) {
		status, err := GetSystemStatus(db, logger, req)
		if err != nil {
			logger.Error("failed to get system status", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to get system status"})
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, status)
	})

	r.Group(func(authRouter chi.Router) {
		authRouter.Use(libhttp.AuthMiddleware(db, logger))
		authRouter.Use(libhttp.AdminMiddleware)

		authRouter.Get("/config", func(res http.ResponseWriter, req *http.Request) {
			cfg := config.AppConfig
			// Sanitize sensitive fields
			cfg.Database.Password = "***"
			cfg.Queue.Password = "***"

			render.Status(req, http.StatusOK)
			render.JSON(res, req, cfg)
		})
	})

	return r
}

func GetSystemStatus(db *gorm.DB, logger *slog.Logger, req *http.Request) (*dto.SystemStatusResponse, error) {
	var (
		initialized            bool
		userOnboardingRequired bool
		needsSuperadmin        bool
	)

	userCount, err := entities.CountUsers(db)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	firstRunCompleteStr, err := settings.GetSetting(db, "first_run_complete", nil)
	if err != nil {
		logger.Warn("could not retrieve 'first_run_complete' setting, assuming false", slog.Any("error", err))
		firstRunCompleteStr = "false"
	}

	firstRunComplete, err := strconv.ParseBool(firstRunCompleteStr)
	if err != nil {
		logger.Error("failed to parse 'first_run_complete' setting to bool, assuming false", slog.String("value", firstRunCompleteStr), slog.Any("error", err))
		firstRunComplete = false
	}

	if userCount == 0 && !firstRunComplete {
		needsSuperadmin = true
		initialized = false
	} else {
		needsSuperadmin = false
		initialized = true
	}

	// For authenticated users, check 'onboarding_complete'
	// Since this endpoint is public, we must manually check for a session
	var user *entities.User
	if cookie, err := req.Cookie(libhttp.AuthTokenCookie); err == nil && cookie.Value != "" {
		if cachedUser, ok := libhttp.GetSessionCache(cookie.Value); ok {
			user = cachedUser
		} else {
			var sess entities.Session
			if err := db.Where("token = ?", cookie.Value).First(&sess).Error; err == nil {
				if sess.ExpiresAt == nil || sess.ExpiresAt.IsZero() || time.Now().Before(*sess.ExpiresAt) {
					var dbUser entities.User
					if err := db.Where("uid = ?", sess.UserUid).First(&dbUser).Error; err == nil {
						user = &dbUser
						libhttp.SetSessionCache(cookie.Value, user, sess.ExpiresAt)
					}
				}
			}
		}
	}

	if user != nil {
		onboardingCompleteStr, err := settings.GetSetting(db, "onboarding_complete", &user.Uid)
		if err != nil {
			logger.Warn("could not retrieve 'onboarding_complete' setting for user, assuming true", slog.String("user_id", user.Uid), slog.Any("error", err))
			// If setting not found for user, assume onboarding is NOT required to avoid blocking
			userOnboardingRequired = false
		} else {
			onboardingComplete, err := strconv.ParseBool(onboardingCompleteStr)
			if err != nil {
				logger.Error("failed to parse 'onboarding_complete' setting to bool, assuming true", slog.String("value", onboardingCompleteStr), slog.Any("error", err))
				userOnboardingRequired = false
			} else {
				userOnboardingRequired = !onboardingComplete
			}
		}
	} else {
		// Not authenticated, so user onboarding not applicable
		userOnboardingRequired = false
	}

	return &dto.SystemStatusResponse{
		Initialized:             initialized,
		UserOnboardingRequired:  userOnboardingRequired,
		NeedsSuperadmin:         needsSuperadmin,
		AllowManualRegistration: config.AppConfig.UserManagement.AllowManualRegistration,
	}, nil
}

func systemCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Cache-Control", "private, max-age=300")
			w.Header().Add("Vary", "Cookie")
		}
		next.ServeHTTP(w, r)
	})
}
