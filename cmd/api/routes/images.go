package routes

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	"imagine/internal/downloads"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/jobs/workers"
	libos "imagine/internal/os"
	"imagine/internal/uid"
	"imagine/internal/utils"
)

type ImageUpload struct {
	Name    string `json:"name,omitempty"`
	Private bool   `json:"private"`
}

type ImageUploadError struct {
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
	Error     string `json:"error"`
}

func createNewImageEntity(logger *slog.Logger, fileName string, libvipsImg *libvips.Image) (*entities.Image, error) {
	logger.Info("Generating ID", slog.String("file", fileName))
	id, err := uid.Generate()

	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	if strings.Trim(fileName, " ") == "" {
		fileName = id
	}

	logger = logger.With(
		slog.String("name", fileName),
		slog.String("id", id),
	)

	logger.Info("reading exif data")
	exifData := libvipsImg.Exif()

	if len(exifData) == 0 {
		logger.Warn("No exif data found. Blank fields", slog.String("file", fileName))
	} else {
		logger.Debug("exif data", slog.Any("data", exifData), slog.Int("length", len(exifData)))
	}

	exif, fileCreatedAt, fileModifiedAt := imageops.BuildImageEXIF(exifData)

	// If EXIF contains a rating-like value, parse it and set the initial
	// canonical rating on the image entity (clamped to 0..5). We store the
	// raw EXIF rating in Exif.Rating as provenance but the top-level Rating
	// becomes the canonical value once DB column exists / migration runs.
	var initialRating *int
	if exif.Rating != nil {
		if r, err := strconv.Atoi(*exif.Rating); err == nil {
			if r < 0 {
				r = 0
			} else if r > 5 {
				r = 5
			}
			initialRating = &r
		}
	}

	var keywords []string
	keywordsPtr := imageops.FindExif(exifData, "Keywords", "Subject")
	if keywordsPtr != nil {
		keywords = strings.Split(*keywordsPtr, ",")
	}

	label := "None"

	metadata := dto.ImageMetadata{
		FileName:         fileName,
		OriginalFileName: &fileName,
		FileType:         string(libvipsImg.Format()),
		ColorSpace:       imageops.GetColourSpaceString(libvipsImg),
		FileModifiedAt:   fileModifiedAt,
		FileCreatedAt:    fileCreatedAt,
		Keywords:         &keywords,
		Label:            &label,
	}

	// Seed canonical rating into the stored image metadata (NULL = unrated)
	metadata.Rating = initialRating

	// Construct paths with reasonable defaults matching the {uid}/file route params
	originalPath := fmt.Sprintf("/images/%s/file", id)
	// Thumbnail: 400px wide, maintain aspect, webp, good quality for small previews
	thumbnailPath := fmt.Sprintf("/images/%s/file?format=webp&w=400&h=400&quality=85", id)
	// Preview: 1920px wide, maintain aspect, webp, balanced quality
	previewPath := fmt.Sprintf("/images/%s/file?format=webp&w=1920&h=1920&quality=90", id)

	paths := dto.ImagePaths{
		Original:  originalPath,
		Thumbnail: thumbnailPath,
		Preview:   previewPath,
	}

	allImageData := entities.Image{
		Uid:           id,
		Name:          fileName,
		Private:       false,
		Processed:     false,
		Exif:          &exif,
		ImageMetadata: &metadata,
		ImagePaths:    paths,
		Width:         int32(libvipsImg.Width()),
		Height:        int32(libvipsImg.Height()),
		Description:   nil, // TODO: evaluate if necessary, blank for now
	}

	ta := imageops.GetTakenAt(allImageData)
	allImageData.TakenAt = &ta

	return &allImageData, nil
}

// moveDirWithFallback attempts to rename src->dst. If rename fails (e.g., cross-device),
// it copies the directory contents to dst and removes the src directory.
func moveDirWithFallback(src, dst string) error {
	return libos.MoveDirWithFallback(src, dst)
}

func ImagesRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	// List images with pagination
	router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		limitStr := req.URL.Query().Get("limit")
		pageStr := req.URL.Query().Get("page")

		limit := 100
		page := 0

		if limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		if pageStr != "" {
			if parsed, err := strconv.Atoi(pageStr); err == nil && parsed >= 0 {
				page = parsed
			}
		}

		var images []entities.Image
		var total int64

		if err := db.WithContext(req.Context()).Transaction(func(tx *gorm.DB) error {
			// Count total non-deleted images for pagination metadata
			if err := tx.Model(&entities.Image{}).Where("deleted_at IS NULL").Count(&total).Error; err != nil {
				return err
			}

			pageOffset := max(page*limit, 0)
			if err := tx.Preload("UploadedBy").Where("deleted_at IS NULL").Order("taken_at DESC NULLS LAST, name DESC").Offset(pageOffset).Limit(limit).Find(&images).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to fetch images"})
			return
		}

		items := make([]dto.ImagesResponse, len(images))
		for i, img := range images {
			items[i] = dto.ImagesResponse{
				AddedAt: img.CreatedAt,
				AddedBy: func() *dto.User {
					if img.UploadedBy != nil {
						d := img.UploadedBy.DTO()
						return &d
					}
					return nil
				}(),
				Image: img.DTO(),
			}
		}

		// Build pagination links
		href := fmt.Sprintf("/images?limit=%d&page=%d", limit, page)
		var prev *string
		var next *string
		hasPrev := page > 0
		hasNext := int64((page+1)*limit) < total
		if hasPrev {
			p := fmt.Sprintf("/images?limit=%d&page=%d", limit, page-1)
			prev = &p
		}
		if hasNext {
			nx := fmt.Sprintf("/images?limit=%d&page=%d", limit, page+1)
			next = &nx
		}

		count := int(total)
		response := dto.ImagesPage{
			Href:  &href,
			Prev:  prev,
			Next:  next,
			Limit: limit,
			Page:  page,
			Count: &count,
			Items: items,
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, response)
	})

	router.Get("/{uid}/file", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		logger := logger.With(slog.String("uid", uid))

		params, err := imageops.ParseTransformParams(req.URL.String())
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid transform parameters"})
			return
		}

		if params.Height <= 0 || params.Width <= 0 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid width/height parameters"})
			return
		}

		var imgEnt entities.Image
		if result := db.Model(&entities.Image{}).Where("uid = ? AND deleted_at IS NULL", uid).First(&imgEnt); result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
				return
			}

			logger.Error("failed to fetch image from database", slog.Any("error", result.Error))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to fetch image from database"})
			return
		}

		isDownload := req.URL.Query().Get("download") == "1"
		if isDownload {
			if !validateDownloadRequest(res, req, db, uid) {
				return
			}
		}

		hasTransformParams := params.Format != "" || params.Width > 0 || params.Height > 0 || params.Quality > 0 || params.Rotate > 0 || params.Flip != ""
		if !hasTransformParams {
			serveOriginalImage(res, req, logger, &imgEnt, isDownload)
			return
		}

		serveTransformedImage(res, req, logger, &imgEnt, params, isDownload)
	})

	router.Get("/{uid}/exif", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		simple := req.URL.Query().Get("simple") == "true"

		var imgEnt entities.Image
		result := db.Model(&entities.Image{}).Where("uid = ? AND deleted_at IS NULL", uid).First(&imgEnt)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
				return
			}

			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to retrieve image"})
			return
		}

		if simple {
			if imgEnt.Exif == nil {
				render.Status(req, http.StatusOK)
				render.JSON(res, req, dto.ErrorResponse{Error: "no exif data"})
				return
			}

			render.Status(req, http.StatusOK)
			render.JSON(res, req, imgEnt.Exif)
			return
		}

		imageFile, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to read image"})
			return
		}

		libvipsImg, err := libvips.NewImageFromBuffer(imageFile, libvips.DefaultLoadOptions())

		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to process image"})
			return
		}
		defer libvipsImg.Close()

		exifData := libvipsImg.Exif()

		render.Status(req, http.StatusOK)
		render.JSON(res, req, exifData)
	})

	router.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")

		var imgEnt entities.Image
		result := db.Preload("UploadedBy").Model(&entities.Image{}).Where("uid = ? AND deleted_at IS NULL", uid).First(&imgEnt)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
				return
			}

			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to retrieve image"})
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, imgEnt.DTO())
	})

	router.Patch("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		var update dto.ImageUpdate

		if err := render.DecodeJSON(req.Body, &update); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		var img entities.Image
		err := db.Transaction(func(tx *gorm.DB) error {
			if e := tx.First(&img, "uid = ? AND deleted_at IS NULL", uid); e.Error != nil {
				return e.Error
			}

			updateImageFromDTO(&img, update)

			if err := tx.Save(&img).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
				return
			}

			logger.Error("Failed to update image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Something went wrong, please try again later"})
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, img.DTO())
	})

	// Dedicated download route: creates a short-lived signed redirect to the
	// file endpoint with download=1 so clients (or browsers) can follow a URL
	// that forces a download and is authorized by HMAC signature.
	router.Get("/{uid}/download", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")

		// Create a short-lived opaque token and redirect to the file URL
		token, err := downloads.CreateToken(db, []string{uid}, 5*time.Minute)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to create download token"})
			return
		}

		redirectURL := fmt.Sprintf("/images/%s/file?download=1&token=%s", uid, token)
		http.Redirect(res, req, redirectURL, http.StatusFound)
	})

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var fileImageUpload dto.ImageUploadRequest

		// Parse the multipart form in the request
		err := req.ParseMultipartForm(10 << 20) // limit your max input length!
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		fileImageUpload.FileName = req.FormValue("file_name")
		fileImageUpload.Checksum = utils.StringPtr(req.FormValue("checksum"))

		// Get the file and header from the form
		file, header, err := req.FormFile("data")
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "missing file data"})
			return
		}
		defer file.Close()

		// Initialize the DTO
		(&fileImageUpload.Data).InitFromMultipart(header)

		imageFileData, err := fileImageUpload.Data.Bytes()
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to read file data"})
			return
		}

		if fileImageUpload.FileName == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "missing filename"})
			return
		}

		if len(imageFileData) == 0 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "empty file data"})
			return
		}

		if fileImageUpload.Checksum != nil && *fileImageUpload.Checksum != "" {
			calculatedChecksum, err := images.CalculateImageChecksum(imageFileData)
			if err != nil {
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: "failed to calculate checksum"})
				return
			}

			if *fileImageUpload.Checksum != calculatedChecksum {
				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "checksum mismatch"})
				return
			}
		}

		libvipsImg, err := libvips.NewImageFromBuffer(imageFileData, libvips.DefaultLoadOptions())
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid image data"})
			return
		}
		defer libvipsImg.Close()

		imageEntity, err := createNewImageEntity(logger, fileImageUpload.FileName, libvipsImg)
		if err != nil {
			logger.Error("Failed to process image data", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to process image data"})
			return
		}

		authUser, _ := libhttp.UserFromContext(req)
		imageEntity.UploadedByID = &authUser.Uid

		var checksum string
		if fileImageUpload.Checksum != nil && *fileImageUpload.Checksum != "" {
			checksum = *fileImageUpload.Checksum
		} else {
			checksum, err = images.CalculateImageChecksum(imageFileData)
			if err != nil {
				logger.Error("Failed to create image", slog.Any("error", err))
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create image"})
				return
			}
		}

		fileSize := int64(len(imageFileData))
		imageEntity.ImageMetadata.FileSize = &fileSize
		imageEntity.ImageMetadata.Checksum = checksum

		var existing entities.Image
		dupErr := db.Where("image_metadata->>'checksum' = ?", checksum).First(&existing).Error
		if dupErr == nil {
			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.ImageUploadResponse{Uid: existing.Uid})
			return
		} else if dupErr != gorm.ErrRecordNotFound {
			logger.Error("Failed to check for duplicates", slog.Any("error", dupErr))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to check for duplicates"})
			return
		}

		logger.Info("adding images to database", slog.String("uid", imageEntity.Uid))
		dbCreateTx := db.Create(&imageEntity)
		if dbCreateTx.Error != nil {
			logger.Error("Failed to create image", slog.Any("error", dbCreateTx.Error))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create image"})
			return
		}

		logger.Info("starting image processing", slog.String("uid", imageEntity.Uid))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		err = images.SaveImage(imageFileData, imageEntity.Uid, imageEntity.ImageMetadata.FileName)
		if err != nil {
			logger.Error("Failed to save image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to save image"})
			return
		}

		jobUid, err := jobs.Enqueue(db, workers.TopicImageProcess, workerJob, nil, &imageEntity.Uid)
		if err != nil {
			logger.Error("Failed to create image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create image"})
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.Uid))

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, dto.ImageUploadResponse{
			Uid: imageEntity.Uid,
			Metadata: &map[string]interface{}{
				"job_uid":   jobUid,
				"file_name": fileImageUpload.FileName,
				"duplicate": dupErr == nil,
			},
		})
	})

	// TODO: either this should be a config option or removed entirely.
	// it's possible that URL uploads could be a security problem.
	// cool idea in theory i guess tho
	router.Post("/url", func(res http.ResponseWriter, req *http.Request) {
		if os.Getenv("ENABLE_URL_UPLOAD") == "false" {
			render.Status(req, http.StatusForbidden)
			render.JSON(res, req, dto.ErrorResponse{Error: "url uploads are disabled"})
			return
		}

		imageUrlBytes, err := io.ReadAll(req.Body)

		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		imageUrl := string(imageUrlBytes)

		// Any failed files should be added to a failed files array/slice
		// and returned to the user after all files have been uploaded
		// at this point in the process, if something has gone so
		// catastrophically wrong to crash the server or whatever, then
		// return a 5XX error but this shouldn't happen

		// Download image first:
		logger.Info("Downloading image", slog.String("url", imageUrl))
		fileBytes, err := libhttp.DownloadFile(imageUrl)

		if err != nil {
			logger.Error("Failed to download file", slog.Any("error", err), slog.String("url", imageUrl))
			return
		}

		urlParsed, err := url.Parse(imageUrl)

		if err != nil {
			logger.Error("Failed to parse url", slog.Any("error", err), slog.String("url", imageUrl))
			return
		}

		fileName, _ := strings.CutPrefix(urlParsed.Path, "/")
		libvipsImg, err := libvips.NewImageFromBuffer(fileBytes, libvips.DefaultLoadOptions())
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}
		defer libvipsImg.Close()
		imageEntity, err := createNewImageEntity(logger, fileName, libvipsImg)

		if err != nil {
			logger.Error("Failed to process image data", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to process image data"})
			return
		}

		// Set the uploader
		authUser, _ := libhttp.UserFromContext(req)
		imageEntity.UploadedByID = &authUser.Uid

		hasher := sha1.New()
		if _, err := io.Copy(hasher, bytes.NewReader(fileBytes)); err != nil {
			logger.Error("Failed to create image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create image"})
			return
		}

		fileSize := int64(len(fileBytes))
		imageEntity.ImageMetadata.FileSize = &fileSize
		checksum := hex.EncodeToString(hasher.Sum(nil))
		imageEntity.ImageMetadata.Checksum = checksum

		var existing entities.Image
		dupErr := db.Where("image_metadata->>'checksum' = ?", checksum).First(&existing).Error
		if dupErr == nil {
			// Duplicate: return existing UID as an ImageUploadResponse (200)
			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.ImageUploadResponse{Uid: existing.Uid})
			return
		} else if dupErr != gorm.ErrRecordNotFound {
			logger.Error("Failed to check for duplicates", slog.Any("error", dupErr))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to check for duplicates"})
			return
		}

		logger.Info("adding image to database", slog.String("id", imageEntity.Uid))
		dbCreateTx := db.Create(&imageEntity)

		if dbCreateTx.Error != nil {
			logger.Error("Failed to create image", slog.Any("error", dbCreateTx.Error))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to create image"})
			return
		}

		logger.Info("starting image processing", slog.String("id", imageEntity.Uid))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		err = images.SaveImage(fileBytes, imageEntity.Uid, imageEntity.ImageMetadata.FileName)
		if err != nil {
			logger.Error("Failed to process image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to process image"})
			return
		}

		_, err = jobs.Enqueue(db, workers.TopicImageProcess, workerJob, nil, &imageEntity.Uid)
		if err != nil {
			logger.Error("Failed to process image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to process image"})
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.Uid))

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, dto.ErrorResponse{Error: imageEntity.Uid})
	})

	router.Delete("/", func(res http.ResponseWriter, req *http.Request) {
		var body struct {
			Uids  []string `json:"uids"`
			Force bool     `json:"force,omitempty"`
		}

		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.DeleteAssetsResponse{Results: nil})
			return
		}

		if len(body.Uids) == 0 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.DeleteAssetsResponse{Results: nil})
			return
		}

		libDir := images.Directory
		trashDir := images.TrashDirectory

		if !body.Force {
			if err := os.MkdirAll(trashDir, 0755); err != nil {
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.DeleteAssetsResponse{Results: nil})
				return
			}
		}

		resultsArr := make([]map[string]any, 0, len(body.Uids))
		var anyFailed bool

		for _, id := range body.Uids {
			src := filepath.Join(libDir, id)
			var deleted bool
			var errMsg *string

			if body.Force {
				// Force delete: Remove from DB permanently and delete files
				if err := db.Unscoped().Where("uid = ?", id).Delete(&entities.Image{}).Error; err != nil {
					logger.Error("failed to hard delete from DB", slog.String("uid", id), slog.Any("error", err))
					e := err.Error()
					errMsg = &e
					deleted = false
					anyFailed = true
				} else if err := os.RemoveAll(src); err != nil {
					logger.Error("failed to force delete asset dir", slog.String("uid", id), slog.Any("error", err))
					e := err.Error()
					errMsg = &e
					deleted = false
					anyFailed = true
				} else {
					deleted = true
				}
			} else {
				// Soft delete: Set DeletedAt in DB and move files to trash
				if err := db.Where("uid = ?", id).Delete(&entities.Image{}).Error; err != nil {
					logger.Error("failed to soft delete from DB", slog.String("uid", id), slog.Any("error", err))
					e := err.Error()
					errMsg = &e
					deleted = false
					anyFailed = true
				} else {
					dst := filepath.Join(trashDir, id)
					if err := moveDirWithFallback(src, dst); err != nil {
						logger.Error("failed to move asset to trash", slog.String("uid", id), slog.Any("error", err))
						e := err.Error()
						errMsg = &e
						deleted = false
						anyFailed = true
					} else {
						deleted = true
					}
				}
			}

			entry := map[string]any{
				"uid":     id,
				"deleted": deleted,
			}
			if errMsg != nil {
				entry["error"] = *errMsg
			}
			resultsArr = append(resultsArr, entry)
		}

		var msg *string
		if anyFailed {
			m := fmt.Sprintf("failed to delete/move ids: %s", func() string {
				var failed []string
				for _, r := range resultsArr {
					if d, ok := r["deleted"].(bool); ok && !d {
						if uid, ok := r["uid"].(string); ok {
							failed = append(failed, uid)
						}
					}
				}
				return strings.Join(failed, ",")
			}())
			msg = &m
		}

		tmp := map[string]any{"results": resultsArr}
		if msg != nil {
			tmp["message"] = *msg
		}

		b, _ := json.Marshal(tmp)
		var resp dto.DeleteAssetsResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			if anyFailed {
				render.Status(req, http.StatusMultiStatus)
			} else {
				render.Status(req, http.StatusOK)
			}
			render.JSON(res, req, tmp)
			return
		}

		if anyFailed {
			render.Status(req, http.StatusMultiStatus)
			render.JSON(res, req, resp)
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, resp)
	})

	return router
}

func serveOriginalImage(res http.ResponseWriter, req *http.Request, logger *slog.Logger, imgEnt *entities.Image, isDownload bool) {
	imageData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
	if err != nil {
		logger.Error("failed to read original image", slog.Any("error", err))
		render.Status(req, http.StatusInternalServerError)
		render.JSON(res, req, dto.ErrorResponse{Error: "failed to read original image"})
		return
	}

	res.Header().Set("Etag", fmt.Sprintf(`"%s"`, imgEnt.ImageMetadata.Checksum))
	res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))

	if isDownload {
		res.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
		res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, imgEnt.ImageMetadata.FileName))
	} else {
		res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	http.ServeContent(res, req, imgEnt.ImageMetadata.FileName, imgEnt.UpdatedAt, bytes.NewReader(imageData))
}

func serveTransformedImage(res http.ResponseWriter, req *http.Request, logger *slog.Logger, imgEnt *entities.Image, params *imageops.TransformParams, isDownload bool) {
	// 1. Determine if this is a "permanent" transform path
	reqURI := req.URL.String()
	isPermanent := imgEnt.ImagePaths.Thumbnail == reqURI || imgEnt.ImagePaths.Preview == reqURI

	// Check if a 'v' (version/checksum) query parameter is present
	hasVersionParam := req.URL.Query().Get("v") != ""

	// 2. Generate ETag for the transform
	transformETag := *imageops.CreateTransformEtag(*imgEnt, params)

	// 3. Check client-side cache first
	if match := req.Header.Get("If-None-Match"); match != "" {
		// Strip quotes from the If-None-Match header value for comparison
		match = strings.Trim(match, `"`)
		if match == transformETag {
			logger.Debug("client-side cache hit (If-None-Match)", slog.String("etag", transformETag))
			res.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if modifiedSince := req.Header.Get("If-Modified-Since"); modifiedSince != "" {
		if t, err := http.ParseTime(modifiedSince); err == nil {
			if !imgEnt.UpdatedAt.After(t) {
				logger.Debug("client-side cache hit (If-Modified-Since)", slog.Time("last-modified", imgEnt.UpdatedAt))
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	ext := params.Format
	if ext == "" {
		ext = imgEnt.ImageMetadata.FileType
	}

	res.Header().Set("Content-Type", "image/"+ext)

	// 4. Check our server-side cache
	cacheKey := strings.Trim(transformETag, `"`)
	cachedData, err := images.ReadCachedTransform(imgEnt.Uid, cacheKey, ext)

	if err == nil {
		// Cache HIT: Serve the cached file
		logger.Debug("server-side cache hit", slog.String("key", cacheKey))
		res.Header().Set("Etag", transformETag)
		res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))

		if isDownload {
			res.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
			res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, imgEnt.ImageMetadata.FileName))
		} else if isPermanent || hasVersionParam {
			// If it's a permanent path or has a version parameter, it's immutable
			res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// Otherwise, use a shorter cache time
			res.Header().Set("Cache-Control", "public, max-age=604800")
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(cachedData)))
		res.WriteHeader(http.StatusOK)
		res.Write(cachedData)
		return
	}

	// 5a. If permanent path and not in cache, it's still processing.
	if isPermanent {
		logger.Info("permanent transform not ready, telling client to retry", slog.String("path", reqURI), slog.String("uid", imgEnt.Uid))
		res.Header().Set("Retry-After", "10") // Tell client to retry after 10 seconds
		res.WriteHeader(http.StatusAccepted)
		return
	}

	// 5b. If not permanent, generate on-the-fly
	logger.Info("server-side cache miss, generating on-demand transform", slog.String("key", cacheKey))
	originalData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
	if err != nil {
		logger.Error("failed to read original for transform", slog.Any("error", err))
		render.Status(req, http.StatusInternalServerError)
		render.JSON(res, req, dto.ErrorResponse{Error: "failed to read original for transform"})
		return
	}

	tresult, err := imageops.GenerateTransform(params, *imgEnt, originalData)
	if err != nil {
		logger.Error("failed to generate transform", slog.Any("error", err))
		render.Status(req, http.StatusInternalServerError)
		render.JSON(res, req, dto.ErrorResponse{Error: "failed to generate transform"})
		return
	}

	// Write to cache in the background
	go func() {
		if err := images.WriteCachedTransform(imgEnt.Uid, cacheKey, ext, tresult.ImageData); err != nil {
			logger.Warn("failed to write on-demand transform to cache", slog.Any("error", err))
		}
	}()

	// Serve the newly generated image
	res.Header().Set("Etag", transformETag)
	res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))
	if isDownload {
		res.Header().Set("Cache-Control", "private, no-cache, no-store, must-revalidate")
		res.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, imgEnt.ImageMetadata.FileName))
	} else if isPermanent || hasVersionParam {
		res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		res.Header().Set("Cache-Control", "public, max-age=604800")
	}

	res.Header().Set("Content-Length", strconv.Itoa(len(tresult.ImageData)))
	res.WriteHeader(http.StatusOK)
	res.Write(tresult.ImageData)
}

func validateDownloadRequest(res http.ResponseWriter, req *http.Request, db *gorm.DB, uid string) bool {
	token := req.URL.Query().Get("token")
	password := req.URL.Query().Get("password")

	if token == "" {
		render.Status(req, http.StatusBadRequest)
		render.JSON(res, req, dto.ErrorResponse{Error: "missing token query param"})
		return false
	}

	uids, tokenEntity, ok := downloads.ValidateTokenWithPassword(db, token, password)
	if !ok {
		if tokenEntity != nil && tokenEntity.Password != nil {
			render.Status(req, http.StatusUnauthorized)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid or missing password"})
			return false
		}
		render.Status(req, http.StatusUnauthorized)
		render.JSON(res, req, dto.ErrorResponse{Error: "invalid or expired token"})
		return false
	}

	if !tokenEntity.AllowDownload {
		render.Status(req, http.StatusForbidden)
		render.JSON(res, req, dto.ErrorResponse{Error: "downloads not permitted for this token"})
		return false
	}

	if !downloads.ValidateEmbedAccess(tokenEntity, req) {
		render.Status(req, http.StatusForbidden)
		render.JSON(res, req, dto.ErrorResponse{Error: "embedding not allowed for this token"})
		return false
	}

	if !slices.Contains(uids, uid) {
		render.Status(req, http.StatusUnauthorized)
		render.JSON(res, req, dto.ErrorResponse{Error: "token not valid for this resource"})
		return false
	}

	return true
}

// updateImageFromDTO updates image entity fields from a small ImageUpdate
func updateImageFromDTO(image *entities.Image, update dto.ImageUpdate) {
	if update.Name != nil {
		image.Name = *update.Name
	}

	if update.Description != nil {
		image.Description = update.Description
	}

	if update.Private != nil {
		image.Private = *update.Private
	}

	if update.Exif != nil {
		image.Exif = update.Exif
	}

	if update.ImageMetadata != nil && update.ImageMetadata.Rating != nil {
		// Clamp rating 0..5
		r := *update.ImageMetadata.Rating
		if r < 0 {
			r = 0
		} else if r > 5 {
			r = 5
		}

		if image.ImageMetadata == nil {
			image.ImageMetadata = &dto.ImageMetadata{}
		}

		image.ImageMetadata.Rating = &r
	}

	if update.ImageMetadata != nil && update.ImageMetadata.Label != nil {
		if image.ImageMetadata == nil {
			image.ImageMetadata = &dto.ImageMetadata{}
		}

		image.ImageMetadata.Label = update.ImageMetadata.Label
	}

	if update.ImageMetadata != nil && update.ImageMetadata.Keywords != nil {
		if image.ImageMetadata == nil {
			image.ImageMetadata = &dto.ImageMetadata{}
		}
		image.ImageMetadata.Keywords = update.ImageMetadata.Keywords
	}

}
