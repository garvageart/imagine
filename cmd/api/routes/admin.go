// This file contains the routes for the admin API
package routes

import (
	"encoding/hex"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
	"log/slog"

	"imagine/internal/crypto"

	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/images"
	libos "imagine/internal/os"
	"imagine/internal/uid"
	"imagine/internal/utils"
)

type AdminUserUpdate struct {
	dto.UserUpdate
	Role *dto.UserRole `json:"role,omitempty"`
}

// For me, start time is described once the config has been read in safely without fail
// From there on, the server will launch and the database will connect which is when things
// really start
var StartTime = time.Now()

// AdminRouter returns a router with admin-only endpoints. It applies AuthMiddleware
// and AdminMiddleware so handlers inside can assume the request is from an admin.
func AdminRouter(db *gorm.DB, logger *slog.Logger, storageStats *images.StorageStatsHolder) *chi.Mux {
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
		render.JSON(res, req, dto.MessageResponse{Message: "Image cache cleared"})
	})

	// System Stats
	r.Get("/system/stats", func(res http.ResponseWriter, req *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// Simple uptime check
		uptime := time.Since(StartTime)

		totalSystemSpace, err := libos.GetTotalDiskSpace(storageStats.GetPath())
		if err != nil {
			logger.Error("failed to get total system space", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to get total system space"})
			return
		}

		stats := dto.SystemStatsResponse{
			UptimeSeconds:    int64(uptime.Seconds()),
			NumGoroutine:     runtime.NumGoroutine(),
			AllocMemory:      int64(m.Alloc),
			SysMemory:        int64(m.Sys),
			StorageUsedBytes: storageStats.GetTotalSize(),
			StoragePath:      storageStats.GetPath(),
			TotalSystemSpaceBytes: func(u uint64) *int64 {
				i := int64(u)
				return &i
			}(totalSystemSpace),
		}

		render.JSON(res, req, stats)
	})

	// Database Stats
	r.Get("/db/stats", func(res http.ResponseWriter, req *http.Request) {
		var userCount int64
		var imageCount int64
		var collectionCount int64

		db.Model(&entities.User{}).Count(&userCount)
		db.Model(&entities.Image{}).Count(&imageCount)
		db.Model(&entities.Collection{}).Count(&collectionCount)

		var dbSize int64
		var activeConnections int64

		// Postgres specific stats
		if db.Dialector.Name() == "postgres" {
			db.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize)
			db.Raw("SELECT count(*) FROM pg_stat_activity").Scan(&activeConnections)
		}

		stats := dto.DatabaseStatsResponse{
			UserCount:       userCount,
			ImageCount:      imageCount,
			CollectionCount: collectionCount,
		}

		if db.Dialector.Name() == "postgres" {
			stats.DbSizeBytes = &dbSize
			stats.ActiveConnections = &activeConnections
		}

		render.JSON(res, req, stats)
	})

	// User Management
	r.Route("/users", func(r chi.Router) {
		r.Get("/", func(res http.ResponseWriter, req *http.Request) {
			var users []entities.User
			if err := db.Find(&users).Error; err != nil {
				logger.Error("failed to list users", slog.Any("error", err))
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: "Failed to list users"})
				return
			}

			// Simple DTO mapping
			userDTOs := make([]dto.User, len(users))
			for i, u := range users {
				userDTOs[i] = u.DTO()
			}

			render.JSON(res, req, userDTOs)
		})

		r.Post("/", func(res http.ResponseWriter, req *http.Request) {
			var adminCreate dto.AdminUserCreate

			if err := render.DecodeJSON(req.Body, &adminCreate); err != nil {
				render.Status(req, http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
				return
			}

			if adminCreate.Name == "" || adminCreate.Password == "" || string(adminCreate.Email) == "" {
				render.Status(req, http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "Required fields are missing"})
				return
			}

			// Check if user exists
			var existingUser entities.User
			if err := db.Where("email = ?", string(adminCreate.Email)).First(&existingUser).Error; err == nil {
				render.Status(req, http.StatusConflict)
				render.JSON(res, req, dto.ErrorResponse{Error: "User already exists"})
				return
			}

			id, err := uid.Generate()
			if err != nil {
				libhttp.ServerError(res, req, err, logger, nil, "Failed to generate UID", "Internal server error")
				return
			}

			role := dto.UserRoleUser
			if adminCreate.Role != nil {
				role = dto.UserRole(*adminCreate.Role)
			}

			userEnt := entities.User{
				Uid:      id,
				Email:    string(adminCreate.Email),
				Username: adminCreate.Name,
				Role:     role,
			}

			argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
			salt := argon.GenerateSalt()
			hashedPass, _ := argon.Hash([]byte(adminCreate.Password), salt)
			hashed := hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

			uwp := entities.FromUser(userEnt, &hashed)

			txErr := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&uwp).Error; err != nil {
					return err
				}

				// Reload the user to get the auto-generated CreatedAt timestamp
				if err := tx.First(&userEnt, "uid = ?", userEnt.Uid).Error; err != nil {
					return err
				}

				// Initialize onboarding_complete to false for new users created by admin
				onboardingOverride := entities.SettingOverride{
					UserId: id,
					Name:   "onboarding_complete",
					Value:  "false",
				}
				if err := tx.Create(&onboardingOverride).Error; err != nil {
					return err
				}

				return nil
			})

			if txErr != nil {
				libhttp.ServerError(res, req, txErr, logger, nil, "Failed to create user", "Internal server error")
				return
			}

			render.Status(req, http.StatusCreated)
			render.JSON(res, req, userEnt.DTO())
		})

		r.Patch("/{uid}", func(res http.ResponseWriter, req *http.Request) {
			uid := chi.URLParam(req, "uid")
			var update AdminUserUpdate
			if err := render.DecodeJSON(req.Body, &update); err != nil {
				render.Status(req, http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
				return
			}

			var user entities.User
			if err := db.Where("uid = ?", uid).First(&user).Error; err != nil {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "User not found"})
				return
			}

			updates := entities.User{
				Username:  *update.Username,
				FirstName: *update.FirstName,
				LastName:  *update.LastName,
				Email:     string(*update.Email),
				Role:      *update.Role,
			}

			if err := db.Model(&user).Updates(updates).Error; err != nil {
				logger.Error("failed to update user", slog.Any("error", err))
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: "Failed to update user"})
				return
			}

			// Force token revocation if role changed? maybe
			if update.Role != nil {
				// Revoke sessions for security
				if err := db.Where("user_uid = ?", user.Uid).Delete(&entities.Session{}).Error; err != nil {
					// failing the request is unecessary
					logger.Error("failed to revoke user sessions after role change", slog.String("uid", user.Uid), slog.Any("error", err))
				}
			}

			render.JSON(res, req, user.DTO())
		})

		r.Delete("/{uid}", func(res http.ResponseWriter, req *http.Request) {
			uid := chi.URLParam(req, "uid")

			// Don't allow deleting self
			requester, _ := libhttp.UserFromContext(req)
			if requester.Uid == uid {
				render.Status(req, http.StatusForbidden)
				render.JSON(res, req, dto.ErrorResponse{Error: "Cannot delete your own account"})
				return
			}

			// Check for force delete in body
			// Since body is optional for DELETE in many clients, we check if there's content
			var deleteReq struct {
				Force bool `json:"force"`
			}
			if req.ContentLength > 0 {
				if err := render.DecodeJSON(req.Body, &deleteReq); err != nil {
					// Invalid body but we continue with safe delete or fail?
					// Let's warn and fail to be safe
					render.Status(req, http.StatusBadRequest)
					render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
					return
				}
			}

			if deleteReq.Force {
				if err := entities.HardDeleteUser(db, uid); err != nil {
					logger.Error("failed to force delete user", slog.String("uid", uid), slog.Any("error", err))
					render.Status(req, http.StatusInternalServerError)
					render.JSON(res, req, dto.ErrorResponse{Error: "Failed to force delete user"})
					return
				}
				render.Status(req, http.StatusOK)
				render.JSON(res, req, dto.MessageResponse{Message: "User permanently deleted"})
				return
			}

			if err := db.Where("uid = ?", uid).Delete(&entities.User{}).Error; err != nil {
				logger.Error("failed to delete user", slog.Any("error", err))
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: "Failed to delete user"})
				return
			}

			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.MessageResponse{Message: "User deleted"})
		})
	})

	return r
}
