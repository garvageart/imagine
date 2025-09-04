package routes

import (
	"encoding/hex"
	"imagine/internal/crypto"
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
		var createdUser entities.User

		err := render.DecodeJSON(req.Body, &createdUser)
		if err != nil {
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		if createdUser.Name == "" || createdUser.Password == "" || createdUser.Email == "" {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "required fields are missing"})
			return
		}

		createdUser.UID, err = uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate user ID",
				"Something went wrong, please try again later",
			)
		}

		//todo: fix this mess. get a string from the salt and hash seperately
		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		salt := argon.GenerateSalt()
		hashedPass, _ := argon.Hash([]byte(createdUser.Password), salt)
		createdUser.Password = hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

		err = db.Create(&createdUser).Error
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to create user",
				"Something went wrong, please try again later",
			)

			return
		}

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, createdUser)
	})

	router.Get("/{id}", func(res http.ResponseWriter, req *http.Request) {
		userID := chi.URLParam(req, "id")

		var user entities.User

		err := db.Where("uid = ?", userID).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				res.WriteHeader(http.StatusNotFound)
				render.JSON(res, req, map[string]string{"error": "user not found"})
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

	return router
}
