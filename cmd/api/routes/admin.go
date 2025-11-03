// This file contains the routes for the admin API
package routes

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    "log/slog"
    "gorm.io/gorm"

    libhttp "imagine/internal/http"
    "imagine/internal/utils"
    "imagine/internal/dto"
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
            res.WriteHeader(http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "healthcheck failed"})
            return
        }

        randomPositiveMessage := []string{
            "all love and peace ",
            "take care of yourself",
            "love is in the air",
            "support open source <3",
        }

        loveMessage := randomPositiveMessage[utils.RandomInt(0, len(randomPositiveMessage)-1)]

        res.WriteHeader(http.StatusOK)
        render.JSON(res, req, dto.MessageResponse{Message: loveMessage})
    })


    return r
}

