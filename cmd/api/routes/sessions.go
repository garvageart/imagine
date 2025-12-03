package routes

import (
	"imagine/internal/dto"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	gocache "github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
)

var SessionCacheDefaultDuration = 30 * time.Minute
var SessionCache = gocache.New(30*time.Minute, 10*time.Minute)

func SessionsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var session entities.Session

		err := render.DecodeJSON(req.Body, &session)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
			return
		}

		if session.Token == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Token is required"})
			return
		}

		session.Uid, err = uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate session ID",
				"Something went wrong, please try again later",
			)
		}

		session.CreatedAt = time.Now()
		t := time.Now().Add(SessionCacheDefaultDuration)
		session.ExpiresAt = &t

		err = db.Create(&session).Error
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create session"})
			return
		}

		SessionCache.Add(session.Uid, session, SessionCacheDefaultDuration)

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, session)
	})

	return router
}
