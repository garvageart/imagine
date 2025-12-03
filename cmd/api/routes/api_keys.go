package routes

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/auth"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
)

// APIKeysRouter manages API key lifecycle: create, list, get, revoke, rotate, delete.
func APIKeysRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
    r := chi.NewRouter()

    r.Post("/", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        var body dto.APIKeyCreate

        if err := render.DecodeJSON(req.Body, &body); err != nil {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
            return
        }

        var existingCount int64
        if err := db.Model(&entities.APIKey{}).Where("user_uid = ?", authUser.Uid).Count(&existingCount).Error; err == nil {
            if existingCount >= 50 {
                render.Status(req, http.StatusTooManyRequests)
                render.JSON(res, req, dto.ErrorResponse{Error: "API key limit reached"})
                return
            }
        }

        keys, err := auth.GenerateAPIKey()
        if err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to generate api key", "Failed to create API key")
            return
        }

        consumerKey := keys["consumer_key"]

        apiKeyUid, err := uid.Generate()
        if err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to generate uid", "Failed to create API key")
            return
        }

        apiEnt := entities.APIKey{
            Uid:         apiKeyUid,
            KeyHashed:   keys["hashed_key"],
            UserID:      &authUser.Uid,
            Revoked:     false,
            Name:        body.Name,
            Description: body.Description,
            Scopes:      *body.Scopes,
            ExpiresAt:   body.ExpiresAt,
        }

        if err := db.Transaction(func(tx *gorm.DB) error {
            if err := tx.Create(&apiEnt).Error; err != nil {
                return err
            }
            return nil
        }); err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "error inserting api key into database", "Something went wrong, please try again later")
            return
        }

        logger.Info("Generated an API key", slog.String("request_id", libhttp.GetRequestID(req)))
        render.Status(req, http.StatusCreated)
        render.JSON(res, req, dto.APIKeyCreateResponse{ConsumerKey: consumerKey, ExpiresAt: body.ExpiresAt})
    })

    r.Get("/", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        var keys []entities.APIKey
        q := db.Order("created_at desc").Model(&entities.APIKey{})
        if authUser.Role != "admin" && authUser.Role != "superadmin" {
            q = q.Where("user_uid = ?", authUser.Uid)
        }

        if err := q.Find(&keys).Error; err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to list api keys", "Something went wrong")
            return
        }

        items := make([]dto.APIKey, 0, len(keys))
        for _, e := range keys {
            items = append(items, e.DTO())
        }

        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.APIKeyListResponse{Items: items, Count: len(items)})
    })

    r.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        keyUid := chi.URLParam(req, "uid")
        var ent entities.APIKey
        q := db.Where("uid = ?", keyUid).Preload("User")
        if authUser.Role != "admin" && authUser.Role != "superadmin" {
            q = q.Where("user_uid = ?", authUser.Uid)
        }

        if err := q.First(&ent).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                render.Status(req, http.StatusNotFound)
                render.JSON(res, req, dto.ErrorResponse{Error: "API key not found"})
                return
            }
            libhttp.ServerError(res, req, err, logger, nil, "failed to fetch api key", "Something went wrong")
            return
        }

        render.Status(req, http.StatusOK)
        render.JSON(res, req, ent.DTO())
    })

    r.Post("/{uid}/revoke", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        keyUid := chi.URLParam(req, "uid")
        updates := map[string]interface{}{"revoked": true, "revoked_at": time.Now()}
		
        if err := db.Transaction(func(tx *gorm.DB) error {
            tq := tx.Model(&entities.APIKey{}).Where("uid = ?", keyUid)
            if authUser.Role != "admin" && authUser.Role != "superadmin" {
                tq = tq.Where("user_uid = ?", authUser.Uid)
            }

            if err := tq.Updates(updates).Error; err != nil {
                return err
            }

            return nil
        }); err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to revoke api key", "Something went wrong")
            return
        }

        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.MessageResponse{Message: "revoked"})
    })

    // Rotate a key: revoke old, create new and return new consumer key once
    r.Post("/{uid}/rotate", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        keyUid := chi.URLParam(req, "uid")

        var existing entities.APIKey
        q := db.Where("uid = ?", keyUid).Model(&entities.APIKey{}).Preload("User")
        if authUser.Role != "admin" && authUser.Role != "superadmin" {
            q = q.Where("user_uid = ?", authUser.Uid)
        }
		
        if err := q.First(&existing).Error; err != nil {
            if err == gorm.ErrRecordNotFound {
                render.Status(req, http.StatusNotFound)
                render.JSON(res, req, dto.ErrorResponse{Error: "API key not found"})
                return
            }
            libhttp.ServerError(res, req, err, logger, nil, "failed to fetch api key", "Something went wrong")
            return
        }

        // Perform revoke + create in a transaction
        // Generate replacement keys outside the transaction to avoid long txs
        keys, err := auth.GenerateAPIKey()
        if err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to generate api key", "Failed to rotate API key")
            return
        }

        consumerKey := keys["consumer_key"]
        newUid, err := uid.Generate()
        if err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to generate uid", "Failed to rotate API key")
            return
        }

        if err := db.Transaction(func(tx *gorm.DB) error {
            // Revoke old
            if err := tx.Model(&entities.APIKey{}).Where("uid = ?", keyUid).Updates(map[string]interface{}{"revoked": true, "revoked_at": time.Now()}).Error; err != nil {
                return err
            }

            // Create replacement
            apiEnt := entities.APIKey{
                Uid:         newUid,
                KeyHashed:   keys["hashed_key"],
                UserID:      existing.UserID,
                Revoked:     false,
                Name:        existing.Name,
                Description: existing.Description,
                Scopes:      existing.Scopes,
                ExpiresAt:   existing.ExpiresAt,
            }

            if err := tx.Create(&apiEnt).Error; err != nil {
                return err
            }

            return nil
        }); err != nil {
            // If record not found we want to surface a 404-like response
            if err == gorm.ErrRecordNotFound {
                render.Status(req, http.StatusNotFound)
                render.JSON(res, req, dto.ErrorResponse{Error: "API key not found"})
                return
            }
            libhttp.ServerError(res, req, err, logger, nil, "failed to rotate api key", "Something went wrong")
            return
        }

        render.Status(req, http.StatusCreated)
        render.JSON(res, req, dto.APIKeyCreateResponse{ConsumerKey: consumerKey})
    })

    // Delete (hard-delete) an API key -- admin or owner
    r.Delete("/{uid}", func(res http.ResponseWriter, req *http.Request) {
        authUser, ok := libhttp.UserFromContext(req)
        if !ok || authUser == nil {
            render.Status(req, http.StatusUnauthorized)
            render.JSON(res, req, dto.ErrorResponse{Error: "Unauthenticated"})
            return
        }

        keyUid := chi.URLParam(req, "uid")

        if err := db.Transaction(func(tx *gorm.DB) error {
            tq := tx.Where("uid = ?", keyUid)
            if authUser.Role != "admin" && authUser.Role != "superadmin" {
                tq = tq.Where("user_uid = ?", authUser.Uid)
            }
            if err := tq.Delete(&entities.APIKey{}).Error; err != nil {
                return err
            }
            return nil
        }); err != nil {
            libhttp.ServerError(res, req, err, logger, nil, "failed to delete api key", "Something went wrong")
            return
        }

        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.MessageResponse{Message: "deleted"})
    })

    return r
}
