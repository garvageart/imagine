package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/common/entities"
	libhttp "imagine/common/http"
	"imagine/common/uid"
)

type ImagesResponse struct {
	AddedAt time.Time `json:"added_at"`
	AddedBy string    `json:"added_by"`
	entities.Image
}

type Images struct {
	*Pagination
	Items []ImagesResponse `json:"items"`
}

func findCollectionImages(db *gorm.DB, imgUIDs []string, images []entities.Image, collection entities.Collection, limit, offset int) ([]entities.Image, error) {
	if err := db.Where("uid IN ?", imgUIDs).
		Limit(limit).Offset(offset).
		Find(&images).Error; err != nil {
		return nil, err
	}

	imgResponse := make([]ImagesResponse, len(collection.Images))
	for i, img := range images {
		imgResponse[i] = ImagesResponse{
			AddedAt: collection.Images[i].AddedAt,
			AddedBy: collection.Images[i].AddedBy,
			Image:   img,
		}
	}

	return images, nil
}

func CollectionsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var collection entities.Collection

		err := render.DecodeJSON(req.Body, &collection)
		if err != nil {
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		if collection.Name == "" {
			render.JSON(res, req, map[string]string{"error": "name is required"})
			return
		}

		colUid, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate collection ID",
				"Something went wrong, please try again later",
			)
		}

		collection.UID = colUid

		err = db.Create(&collection).Error
		if err != nil {
			render.JSON(res, req, map[string]string{"error": "failed to create collection"})
			return
		}

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, collection)
	})

	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err != nil {
			limit = 20
		}

		offset, err := strconv.Atoi(req.URL.Query().Get("offset"))
		if err != nil {
			offset = 0
		}

		var collections []entities.Collection

		err = db.Limit(limit).Offset(offset).Find(&collections).Error
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to get collections",
				"Something went wrong, please try again later",
			)
			return
		}

		if len(collections) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		nextOffset := min(offset+limit, len(collections))

		prevOffset := new(int)
		*prevOffset = offset - limit

		if *prevOffset < 0 {
			prevOffset = nil
		}

		result := struct {
			Href   string `json:"href"`
			Prev   string `json:"prev"`
			Next   string `json:"next"`
			Limit  int    `json:"limit"`
			Offset int    `json:"offset"`
			Count  int    `json:"count"`
			Items  []entities.Collection
		}{
			Href:   fmt.Sprintf("/collections/?offset=%d&limit=%d", offset, limit),
			Prev:   fmt.Sprintf("/collections/?offset=%d&limit=%d", prevOffset, limit),
			Next:   fmt.Sprintf("/collections/?offset=%d&limit=%d", nextOffset, limit),
			Limit:  limit,
			Offset: offset,
			Count:  len(collections),
			Items:  collections,
		}

		res.WriteHeader(http.StatusOK)
		render.JSON(res, req, result)
	})

	router.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		defaultImageLimit := 100
		defaultImageOffset := 0

		var collection entities.Collection
		var images []entities.Image
		var imgResponse []ImagesResponse

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			imageUIDs := make([]string, len(collection.Images))
			for i, img := range collection.Images {
				imageUIDs[i] = img.UID
			}

			allColImages, err := findCollectionImages(tx, imageUIDs, images, collection, defaultImageLimit, defaultImageOffset)
			if err != nil {
				return err
			}

			images = allColImages
			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				res.WriteHeader(http.StatusNotFound)
				render.JSON(res, req, map[string]string{"error": "collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Something went wrong, please try again later",
			)
			return
		}

		result := struct {
			entities.Collection
			Images Images `json:"images"`
		}{
			Collection: collection,
			Images: Images{
				Pagination: &Pagination{
					Href:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, defaultImageOffset, defaultImageLimit),
					Prev:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, defaultImageOffset-defaultImageLimit, defaultImageLimit),
					Next:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, defaultImageOffset+defaultImageLimit, defaultImageLimit),
					Limit:  defaultImageLimit,
					Offset: defaultImageOffset,
					Count:  len(images),
				},
				Items: imgResponse,
			},
		}

		render.JSON(res, req, result)
	})

	router.Get("/{uid}/images", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err != nil {
			limit = 100
		}

		offset, err := strconv.Atoi(req.URL.Query().Get("offset"))
		if err != nil {
			offset = 0
		}

		var images []entities.Image
		var imgResponse []ImagesResponse
		var collection entities.Collection

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Select("images").First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			imageUIDs := make([]string, len(collection.Images))
			for i, img := range collection.Images {
				imageUIDs[i] = img.UID
			}

			allColImages, err := findCollectionImages(tx, imageUIDs, images, collection, limit, offset)
			if err != nil {
				return err
			}

			images = allColImages
			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				res.WriteHeader(http.StatusNotFound)
				render.JSON(res, req, map[string]string{"error": "collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Something went wrong, please try again later",
			)
			return
		}

		result := Images{
			Pagination: &Pagination{
				Href:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, offset, limit),
				Prev:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, offset-limit, limit),
				Next:   fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, offset+limit, limit),
				Limit:  limit,
				Offset: offset,
				Count:  len(images),
			},
			Items: imgResponse,
		}

		render.JSON(res, req, result)
	})

	return router
}
