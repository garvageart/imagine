package routes

import (
	"encoding/hex"
	"fmt"
	"imagine/internal/crypto"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

func AccountsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var create dto.UserCreate

		err := render.DecodeJSON(req.Body, &create)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		if create.Name == "" || create.Password == "" || string(create.Email) == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "required fields are missing"})
			return
		}

		var existingUser entities.User
		tx := db.Where("email = ?", string(create.Email)).First(&existingUser)
		if existingUser.Email != "" {
			render.Status(req, http.StatusConflict)
			render.JSON(res, req, dto.ErrorResponse{Error: "user already exists"})
			return
		} else if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
			libhttp.ServerError(res, req, tx.Error, logger, nil,
				"Failed to check existing user",
				"Something went wrong, please try again later",
			)
			return
		}

		id, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate user ID",
				"Something went wrong, please try again later",
			)
			return
		}

		userEnt := entities.User{
			Uid:       id,
			Email:     string(create.Email),
			Username:  create.Name,
			FirstName: "",
			LastName:  "",
		}

		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		salt := argon.GenerateSalt()
		hashedPass, _ := argon.Hash([]byte(create.Password), salt)
		hashed := hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&userEnt).Error; err != nil {
				return err
			}

			if err := tx.Table("users").Where("uid = ?", id).Update("password", hashed).Error; err != nil {
				return err
			}

			return nil
		})

		if txErr != nil {
			libhttp.ServerError(res, req, txErr, logger, nil,
				"Failed to create user",
				"Something went wrong, please try again later",
			)
			return
		}

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, userEnt.DTO())
	})

	router.Get("/{id}", func(res http.ResponseWriter, req *http.Request) {
		userID := chi.URLParam(req, "id")

		var user entities.User

		err := db.Where("uid = ?", userID).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "user not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to get user",
				"Something went wrong, please try again later",
			)
			return
		}

		render.JSON(res, req, user)
	})

	// Authenticated routes
	router.Group(func(r chi.Router) {
		r.Use(libhttp.AuthMiddleware(db, logger))

		// GET /accounts/me - return the currently authenticated user
		r.Get("/me", func(res http.ResponseWriter, req *http.Request) {
			user, ok := libhttp.UserFromContext(req)
			if !ok || user == nil {
				render.Status(req, http.StatusUnauthorized)
				render.JSON(res, req, dto.ErrorResponse{Error: "not authenticated"})
				return
			}

			// Compute ETag based on user's UpdatedAt and UID and support conditional
			// requests for bandwidth savings.
			etag := fmt.Sprintf("W/\"%d-%s\"", user.UpdatedAt.UnixNano(), user.Uid)
			res.Header().Set("Cache-Control", "private, max-age=60, must-revalidate")
			res.Header().Set("ETag", etag)

			if match := req.Header.Get("If-None-Match"); match == etag {
				res.WriteHeader(http.StatusNotModified)
				return
			}

			render.JSON(res, req, user.DTO())
		})
	})

	return router
}
