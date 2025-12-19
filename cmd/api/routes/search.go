package routes

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/search"
)

// SearchRouter creates a new router for search-related endpoints
func SearchRouter(db *gorm.DB, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()
	r.Get("/", func(res http.ResponseWriter, req *http.Request) {
		queryParam := req.URL.Query().Get("q")
		limitParam := req.URL.Query().Get("limit")
		pageParam := req.URL.Query().Get("page")

		criteria := search.ParseQuery(queryParam)
		engine := search.NewEngine()

		// security filters (private = false OR (private = true AND owner_id = :user))
		securityScope := func(db *gorm.DB) *gorm.DB {
			userID := ""
			if user, ok := libhttp.UserFromContext(req); ok && user != nil {
				userID = user.Uid
			} else if apiKey, ok := libhttp.APIKeyFromContext(req); ok && apiKey != nil && apiKey.User != nil {
				userID = apiKey.User.Uid
			}

			if userID != "" {
				// allow public items OR their own private items
				return db.Where("private = ? OR (private = ? AND owner_id = ?)", false, true, userID)
			}

			// Fallback (should be covered by middleware, but safe default): only public
			return db.Where("private = ?", false)
		}

		imagesQuery := engine.Apply(db, criteria).Scopes(securityScope)

		limit := 100
		page := 0
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}

		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}

		imagesQuery = imagesQuery.Limit(limit).Offset((page - 1) * limit)

		var images []entities.Image
		if err := imagesQuery.Find(&images).Error; err != nil {
			logger.Error("failed to search images", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{
				Error: "Failed to search images",
			})

			return
		}

		collectionsQuery := engine.ApplyCollections(db, criteria).Scopes(securityScope)
		collectionsQuery = collectionsQuery.Limit(limit).Offset((page - 1) * limit)

		var collections []entities.Collection
		if err := collectionsQuery.Find(&collections).Error; err != nil {
			logger.Error("Failed to search collections", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{
				Error: "Failed to search collections",
			})

			return
		}

		imagesDTO := make([]dto.Image, 0)
		for _, img := range images {
			imagesDTO = append(imagesDTO, img.DTO())
		}

		collectionsDTO := make([]dto.Collection, 0)
		for _, col := range collections {
			collectionsDTO = append(collectionsDTO, col.DTO())
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.SearchListResponse{
			Images:      imagesDTO,
			Collections: collectionsDTO,
		})
	})

	return r
}
