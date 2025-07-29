package main

import (
	"imagine/common/entities"
	libhttp "imagine/common/http"
	"imagine/common/uid"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
)

func AccountsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/user", func(res http.ResponseWriter, req *http.Request) {
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

	router.Get("/user/{user_id}", func(res http.ResponseWriter, req *http.Request) {
		userID := chi.URLParam(req, "user_id")

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
