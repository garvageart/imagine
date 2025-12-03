package routes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
)

func findCollectionImages(db *gorm.DB, imgUIDs []string, collection entities.Collection, limit, offset int) ([]dto.ImagesResponse, error) {
	var images []entities.Image

	if err := db.Preload("UploadedBy").Where("uid IN ?", imgUIDs).
		Limit(limit).Offset(offset).
		Find(&images).Error; err != nil {
		return nil, err
	}

	// Build a lookup map to safely pair metadata by UID regardless of DB row order.
	// collection.Images is *[]dto.CollectionImage
	var collectionImages []dto.CollectionImage
	if collection.Images != nil {
		collectionImages = *collection.Images
	}

	meta := make(map[string]dto.CollectionImage, len(collectionImages))
	for _, m := range collectionImages {
		meta[m.Uid] = m
	}

	imgResponse := make([]dto.ImagesResponse, len(images))
	for i, img := range images {
		m := meta[img.Uid]
		imgResponse[i] = dto.ImagesResponse{
			AddedAt: m.AddedAt,
			AddedBy: m.AddedBy,
			Image:   img.DTO(),
		}
	}

	return imgResponse, nil
}

func CollectionsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var create struct {
			Description *string `json:"description,omitempty"`
			Name        string  `json:"name"`
			Private     *bool   `json:"private"`
		}

		err := render.DecodeJSON(req.Body, &create)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
			return
		}

		if create.Name == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Name is required"})
			return
		}

		colUid, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate collection ID",
				"Something went wrong, please try again later",
			)
		}

		authUser, _ := libhttp.UserFromContext(req)

		// Map request -> entity for persistence
		collection := entities.Collection{
			Uid:         colUid,
			Name:        create.Name,
			Private:     create.Private,
			Description: create.Description,
			CreatedByID: &authUser.Uid,
		}

		err = db.Create(&collection).Error
		if err != nil {
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create collection"})
			return
		}

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, collection.DTO())
	})

	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		limit, err := strconv.Atoi(req.URL.Query().Get("limit"))
		if err != nil {
			limit = 20
		}

		page, err := strconv.Atoi(req.URL.Query().Get("page"))
		if err != nil {
			page = 0
		}

		var collections []entities.Collection
		var total int64

		err = db.Transaction(func(tx *gorm.DB) error {
			// Count total collections
			if err := tx.Model(&entities.Collection{}).Count(&total).Error; err != nil {
				return err
			}

			// Fetch current page
			return tx.Preload("Thumbnail").Preload("CreatedBy").
				Limit(limit).
				Offset(page * limit).
				Find(&collections).Error
		})

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to get collections",
				"Something went wrong, please try again later",
			)
			return
		}

		// Convert entities to DTOs for response
		items := make([]dto.Collection, len(collections))
		for i := range collections {
			items[i] = collections[i].DTO()
		}

		// Build pagination links
		href := fmt.Sprintf("/collections/?limit=%d&page=%d", limit, page)
		var prev *string
		var next *string
		hasPrev := page > 0
		hasNext := int64((page+1)*limit) < total
		if hasPrev {
			p := fmt.Sprintf("/collections/?limit=%d&page=%d", limit, page-1)
			prev = &p
		}
		if hasNext {
			nx := fmt.Sprintf("/collections/?limit=%d&page=%d", limit, page+1)
			next = &nx
		}

		count := int(total)
		result := dto.CollectionListResponse{
			Href:  &href,
			Prev:  prev,
			Next:  next,
			Limit: limit,
			Page:  page,
			Count: &count,
			Items: items,
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, result)
	})

	router.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		defaultImageLimit := 100
		defaultImageOffset := 0

		var collection entities.Collection
		var imgResponse []dto.ImagesResponse

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Preload("Thumbnail").Preload("CreatedBy").First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			var collectionImages []dto.CollectionImage
			if collection.Images != nil {
				collectionImages = *collection.Images
			}

			imageUIDs := make([]string, len(collectionImages))
			for i, img := range collectionImages {
				imageUIDs[i] = img.Uid
			}

			allColImages, err := findCollectionImages(tx, imageUIDs, collection, defaultImageLimit, defaultImageOffset)
			if err != nil {
				return err
			}

			imgResponse = allColImages
			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "Collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Something went wrong, please try again later",
			)
			return
		}

		defaultImagePage := 0
		href := fmt.Sprintf("/collections/%s/images/?page=%d&limit=%d", uid, defaultImagePage, defaultImageLimit)

		var next *string
		var totalImages int
		if collection.Images != nil {
			totalImages = len(*collection.Images)
		}

		if totalImages > defaultImageLimit {
			nxPtr := fmt.Sprintf("/collections/%s/images/?page=%d&limit=%d", uid, defaultImagePage+1, defaultImageLimit)
			next = &nxPtr
		}

		var prev *string

		count := len(imgResponse)

		imagesPage := dto.ImagesPage{
			Href:  &href,
			Prev:  prev,
			Next:  next,
			Limit: defaultImageLimit,
			Page:  defaultImagePage,
			Count: &count,
			Items: imgResponse,
		}

		// Use the entity's DTO() method which handles Thumbnail conversion
		collectionDTO := collection.DTO()

		result := dto.CollectionDetailResponse{
			Uid:         collectionDTO.Uid,
			Name:        collectionDTO.Name,
			ImageCount:  &collectionDTO.ImageCount,
			Private:     collectionDTO.Private,
			Images:      imagesPage,
			CreatedBy:   collectionDTO.CreatedBy,
			CreatedAt:   collectionDTO.CreatedAt,
			UpdatedAt:   collectionDTO.UpdatedAt,
			Description: collectionDTO.Description,
			Thumbnail:   collectionDTO.Thumbnail,
		}

		render.JSON(res, req, result)
	})

	router.Patch("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		var update dto.CollectionUpdate
		var collection entities.Collection

		err := render.DecodeJSON(req.Body, &update)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid request body"})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			dbTx := tx.Preload("Thumbnail").Preload("CreatedBy").First(&collection, "uid = ?", uid)
			if dbTx.Error != nil {
				return dbTx.Error
			}

			updateCollectionFromDTO(&collection, update)

			if err := tx.Save(&collection).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "Collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to update collection",
				"Something went wrong, please try again later",
			)
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, collection.DTO())
	})

	router.Delete("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")

		err := db.Transaction(func(tx *gorm.DB) error {
			var collection entities.Collection
			if err := tx.First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			if err := tx.Delete(&collection).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "Collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to delete collection",
				"Something went wrong, please try again later",
			)
			return
		}

		res.WriteHeader(http.StatusNoContent)
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

		var imgResponse []dto.ImagesResponse
		var collection entities.Collection

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Select("images").First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			var collectionImages []dto.CollectionImage
			if collection.Images != nil {
				collectionImages = *collection.Images
			}

			imageUIDs := make([]string, len(collectionImages))
			for i, img := range collectionImages {
				imageUIDs[i] = img.Uid
			}

			imgResponse, err = findCollectionImages(tx, imageUIDs, collection, limit, offset)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "Collection not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Something went wrong, please try again later",
			)
			return
		}

		href := fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, offset, limit)

		var prev *string
		if offset > 0 {
			pv := fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, max(offset-limit, 0), limit)
			prev = &pv
		}

		var next *string
		var totalImages int
		if collection.Images != nil {
			totalImages = len(*collection.Images)
		}

		if offset+limit < totalImages {
			nx := fmt.Sprintf("/collections/%s/images/?offset=%d&limit=%d", uid, offset+limit, limit)
			next = &nx
		}

		count := len(imgResponse)

		result := dto.ImagesPage{
			Href:  &href,
			Prev:  prev,
			Next:  next,
			Limit: limit,
			Page:  offset / limit, // derive page index from original row offset
			Count: &count,
			Items: imgResponse,
		}

		render.JSON(res, req, result)
	})

	router.Put("/{uid}/images", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		var colImage struct {
			UIDs []string `json:"uids"`
		}

		err := render.DecodeJSON(req.Body, &colImage)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.AddImagesResponse{Added: false, Error: ptrString("invalid request body")})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var collection entities.Collection
			if err := tx.First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			for _, imgUID := range colImage.UIDs {
				var img entities.Image

				if err := tx.First(&img, "uid = ?", imgUID).Error; err != nil {
					return err
				}

				colImageEnt := dto.CollectionImage{
					Uid:     imgUID,
					AddedAt: time.Now(),
				}

				// Append to the slice
				var images []dto.CollectionImage
				if collection.Images != nil {
					images = *collection.Images
				}
				images = append(images, colImageEnt)
				collection.Images = &images
			}

			collection.ImageCount = len(*collection.Images)

			return tx.Save(&collection).Error
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.AddImagesResponse{Added: false, Error: ptrString("collection or image not found")})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Something went wrong, please try again later",
			)
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.AddImagesResponse{Added: true})
	})

	router.Delete("/{uid}/images", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		var body struct {
			UIDs []string `json:"uids"`
		}

		if err := render.DecodeJSON(req.Body, &body); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.AddImagesResponse{Added: false, Error: ptrString("invalid request body")})
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			var collection entities.Collection
			if err := tx.First(&collection, "uid = ?", uid).Error; err != nil {
				return err
			}

			var images []dto.CollectionImage
			if collection.Images != nil {
				images = *collection.Images
			}

			if len(images) == 0 || len(body.UIDs) == 0 {
				collection.ImageCount = len(images)
				return tx.Save(&collection).Error
			}

			toRemove := make(map[string]struct{}, len(body.UIDs))
			for _, u := range body.UIDs {
				toRemove[u] = struct{}{}
			}

			j := 0
			for i := 0; i < len(images); i++ {
				img := images[i]
				if _, found := toRemove[img.Uid]; !found {
					images[j] = img
					j++
				}
			}

			if j == 0 {
				collection.Images = nil
			} else {
				tmp := make([]dto.CollectionImage, j)
				copy(tmp, images[:j])
				collection.Images = &tmp
			}
			collection.ImageCount = j

			return tx.Save(&collection).Error
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.DeleteImagesResponse{Deleted: false, Error: ptrString("collection not found")})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to remove images from collection",
				"Something went wrong, please try again later",
			)
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.DeleteImagesResponse{Deleted: true})
	})

	return router
}

// Helper function to create a string pointer
func ptrString(s string) *string {
	return &s
}

// updateCollectionFromDTO updates collection entity fields from a CollectionUpdate DTO
func updateCollectionFromDTO(collection *entities.Collection, update dto.CollectionUpdate) {
	if update.Name != nil {
		collection.Name = *update.Name
	}
	if update.Description != nil {
		collection.Description = update.Description
	}
	if update.Private != nil {
		collection.Private = update.Private
	}
	if update.ThumbnailUID != nil {
		if *update.ThumbnailUID == "" {
			collection.ThumbnailID = nil
		} else {
			collection.ThumbnailID = update.ThumbnailUID
		}
	}
	if update.OwnerUID != nil {
		collection.CreatedByID = update.OwnerUID
	}
}
