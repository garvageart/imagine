// This file contains the routes for the admin API
package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
	"log/slog"

	"imagine/internal/dto"
	libhttp "imagine/internal/http"
	"imagine/internal/images"
	"imagine/internal/utils"
)

// AdminRouter returns a router with admin-only endpoints. It applies AuthMiddleware
// and AdminMiddleware so handlers inside can assume the request is from an admin.
func AdminRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Apply authentication and admin role checks to all routes in this router
	r.Use(libhttp.AuthMiddleware(db, logger))
	r.Use(libhttp.AdminMiddleware)

	// Admin-only healthcheck
	r.Post("/healthcheck", func(res http.ResponseWriter, req *http.Request) {
		result := db.Exec("SELECT 1")
		if result.Error != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Healthcheck failed"})
			return
		}

		randomPositiveMessage := []string{
			"all love and peace ",
			"take care of yourself",
			"love is in the air",
			"support open source <3",
		}

		loveMessage := randomPositiveMessage[utils.RandomInt(0, len(randomPositiveMessage)-1)]

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: loveMessage})
	})

	// Get cache status
	r.Get("/cache/status", func(res http.ResponseWriter, req *http.Request) {
		status, err := images.GetCacheStatus()
		if err != nil {
			logger.Error("failed to get cache status", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to get cache status"})
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.CacheStatusResponse{
			Size:     status.Size,
			Items:    status.Items,
			Hits:     status.Hits,
			Misses:   status.Misses,
			HitRatio: status.HitRatio,
		})
	})

	// Clear image cache
	r.Delete("/cache", func(res http.ResponseWriter, req *http.Request) {
		err := images.ClearCache(logger)
		if err != nil {
			logger.Error("failed to clear image cache", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to clear image cache"})
			return
		}
		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: "image cache cleared"})
	})

	return r
}
