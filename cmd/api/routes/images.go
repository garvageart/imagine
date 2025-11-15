package routes

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
	"golang.org/x/sync/singleflight"

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
	"imagine/internal/jobs/workers"
	libos "imagine/internal/os"
	"imagine/internal/uid"
)

// transformGroup coalesces concurrent identical transform requests to avoid duplicate work
var transformGroup singleflight.Group


type ImageUpload struct {
	Name    string `json:"name,omitempty"`
	Private bool   `json:"private"`
}

type ImageUploadFile struct {
	Data     []byte `json:"data"`
	FileName string `json:"filename"`
	Checksum string `json:"checksum,omitempty"`
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

			pageOffset := max(page * limit, 0)
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

	// TODO: Get param values to serve the original file in different
	// formats, resolutions and sizes
	// TODO TODO: Finalize route name. "/file/" isn't exactly honest in my opinion
	router.Get("/{uid}/file", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")

		formatParam := req.FormValue("format")
		widthParam := req.FormValue("w")
		heightParam := req.FormValue("h")
		qualityParam := req.FormValue("quality")

		// Explicitly exclude soft-deleted images
		tx := db.Model(&entities.Image{}).Where("uid = ? AND deleted_at IS NULL", uid)
		var imgEnt entities.Image
		result := tx.Preload("UploadedBy").First(&imgEnt)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
				return
			}
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to fetch image from database"})
			return
		}

		// If no transform parameters are present, serve the original image.
		// Previously this checked len(req.Form) which treated any query param
		// (including download=1) as a transform request and caused parse errors.
		// Build a canonical request URI (path + query) and determine once
		// whether this request matches any stored image path in the DB.
		var reqURI = req.URL.Path
		var isPermanent = false
		if req.URL.RawQuery != "" {
			reqURI = reqURI + "?" + req.URL.RawQuery
		}

		if imgEnt.ImagePaths.Thumbnail == reqURI || imgEnt.ImagePaths.Preview == reqURI || imgEnt.ImagePaths.Original == reqURI {
			isPermanent = true
		}

		if imgEnt.ImagePaths.Raw != nil && *imgEnt.ImagePaths.Raw == reqURI {
			isPermanent = true
		}

		if formatParam == "" && widthParam == "" && heightParam == "" && qualityParam == "" {
			imageData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
			if err != nil {
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: err.Error()})
				return
			}

			// Set cache headers based on permanence
			if isPermanent {
				res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				// Cache for 1 day in browser, but prevent shared caches from storing.
				res.Header().Set("Cache-Control", "private, max-age=86400, no-transform")
			}
			res.Header().Set("ETag", imgEnt.ImageMetadata.Checksum) // Use checksum as ETag
			// Last-Modified from DB metadata so proxies and clients can use If-Modified-Since
			res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))
			res.Header().Set("Content-Type", "image/"+imgEnt.ImageMetadata.FileType)
			res.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
			res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", imgEnt.ImageMetadata.FileName))

			// If this is a download request, validate token details
			// For non-download requests allow normal 304/NotModified behavior.
			if req.FormValue("download") == "1" {
				// Validate opaque token with password support
				token := req.URL.Query().Get("token")
				password := req.URL.Query().Get("password")

				if token == "" {
					render.Status(req, http.StatusBadRequest)
					render.JSON(res, req, dto.ErrorResponse{Error: "missing token query param"})
					return
				}

				uids, tokenEntity, ok := downloads.ValidateTokenWithPassword(db, token, password)
				if !ok {
					if tokenEntity != nil && tokenEntity.Password != nil {
						render.Status(req, http.StatusUnauthorized)
						render.JSON(res, req, dto.ErrorResponse{Error: "invalid or missing password"})
						return
					}
					render.Status(req, http.StatusUnauthorized)
					render.JSON(res, req, dto.ErrorResponse{Error: "invalid or expired token"})
					return
				}

				// Check if downloads are allowed
				if !tokenEntity.AllowDownload {
					render.Status(req, http.StatusForbidden)
					render.JSON(res, req, dto.ErrorResponse{Error: "downloads not permitted for this token"})
					return
				}

				// Validate embed access (prevents hotlinking if AllowEmbed is false)
				if !downloads.ValidateEmbedAccess(tokenEntity, req) {
					render.Status(req, http.StatusForbidden)
					render.JSON(res, req, dto.ErrorResponse{Error: "embedding not allowed for this token"})
					return
				}

				// Ensure token is valid for this UID
				allowed := slices.Contains(uids, uid)
				if !allowed {
					render.Status(req, http.StatusUnauthorized)
					render.JSON(res, req, dto.ErrorResponse{Error: "token not valid for this resource"})
					return
				}
			} else {
				if match := req.Header.Get("If-None-Match"); match == imgEnt.ImageMetadata.Checksum {
					res.Header().Set("ETag", imgEnt.ImageMetadata.Checksum)
					res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))
					res.WriteHeader(http.StatusNotModified)
					return
				}
			}

			res.WriteHeader(http.StatusOK)
			res.Write(imageData)
			return
		}

		width, wErr := strconv.ParseInt(widthParam, 10, 64)
		height, hErr := strconv.ParseInt(heightParam, 10, 64)
		quality, qErr := strconv.ParseInt(qualityParam, 10, 64)

		// Convert quality to 0-10 instead of 0-100 for the Compression option
		// NOTE: This is not final, depending on if this idea is understandable
		// and accepted by the community/users on release
		if formatParam == "png" {
			quality = int64(math.Round(float64(quality/100) * 10))
		}

		if wErr != nil || hErr != nil || qErr != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request parameters"})
			return
		}

		if width < 0 || height < 0 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "width/height cannot be negative"})
			return
		}

		// Generate ETag for transformed image based on params BEFORE processing
		transformETag := fmt.Sprintf("%s-%dx%d-%s-%d", imgEnt.ImageMetadata.Checksum, width, height, formatParam, quality)

		// Check If-None-Match early to avoid expensive VIPS processing (skip for downloads)
		if req.FormValue("download") != "1" {
			if match := req.Header.Get("If-None-Match"); match == transformETag {
				if isPermanent {
					res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					res.Header().Set("Cache-Control", "public, max-age=604800, no-transform")
				}
				res.Header().Set("ETag", transformETag)
				res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}

		// Use singleflight to coalesce concurrent identical transform requests
		key := transformETag
		resVal, err, _ := transformGroup.Do(key, func() (any, error) {
			// Read the original image file directly
			originalData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
			if err != nil {
				logger.Error("failed to read image data", slog.Any("error", err))
				return nil, fmt.Errorf("failed to read image data: %w", err)
			}

			// Load into libvips
			libvipsImg, err := libvips.NewImageFromBuffer(originalData, libvips.DefaultLoadOptions())
			if err != nil {
				logger.Error("failed to create libvips image", slog.Any("error", err))
				return nil, fmt.Errorf("failed to create libvips image: %w", err)
			}
			defer libvipsImg.Close()

			err = libvipsImg.Autorot() // non-fatal
			if err != nil {
				logger.Warn("failed to auto-rotate image", slog.Any("error", err))
			}

			// Resize with libvips
			if err := libvipsImg.Resize(float64(width)/float64(libvipsImg.Width()), &libvips.ResizeOptions{Kernel: libvips.KernelLanczos3}); err != nil {
				logger.Error("failed to resize image", slog.Any("error", err))
				return nil, fmt.Errorf("failed to resize image: %w", err)
			}

			var imageData []byte

			switch formatParam {
			case "webp":
				imageData, err = libvipsImg.WebpsaveBuffer(&libvips.WebpsaveBufferOptions{Q: int(quality)})
			case "png":
				imageData, err = libvipsImg.PngsaveBuffer(&libvips.PngsaveBufferOptions{Filter: libvips.PngFilterNone, Interlace: false, Palette: false, Compression: int(quality)})
			case "jpg", "jpeg":
				imageData, err = libvipsImg.JpegsaveBuffer(&libvips.JpegsaveBufferOptions{Q: int(quality), Interlace: true})
			case "avif", "heif":
				imageData, err = libvipsImg.HeifsaveBuffer(&libvips.HeifsaveBufferOptions{Q: int(quality), Bitdepth: 8, Effort: 5, Lossless: false})
			default:
				imageData, err = libvipsImg.RawsaveBuffer(&libvips.RawsaveBufferOptions{Keep: libvips.KeepAll})
				formatParam = string(libvipsImg.Format())
			}

			if err != nil {
				return nil, fmt.Errorf("failed to encode image: %w", err)
			}

			return imageData, nil
		})

		if err != nil {
			logger.Error("image transform failed", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to process image"})
			return
		}

		imageData := resVal.([]byte)

		// Set response headers for transform result
		res.Header().Set("Content-Type", "image/"+formatParam)
		// If this request matches a stored image path or includes an explicit
		// fingerprint `v=`, serve with immutable long-term caching. Otherwise
		// use the shorter transform TTL.
		if isPermanent || req.URL.Query().Get("v") != "" {
			res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			res.Header().Set("Cache-Control", "public, max-age=604800, no-transform")
		}

		res.Header().Set("ETag", transformETag)
		res.Header().Set("Last-Modified", imgEnt.UpdatedAt.UTC().Format(http.TimeFormat))
		res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", imgEnt.ImageMetadata.FileName))

		// If this is a download request, validate token with details
		if req.FormValue("download") == "1" {
			// Validate opaque token with password support
			token := req.URL.Query().Get("token")
			password := req.URL.Query().Get("password")

			if token == "" {
				render.Status(req, http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "missing token query param"})
				return
			}

			uids, tokenEntity, ok := downloads.ValidateTokenWithPassword(db, token, password)
			if !ok {
				if tokenEntity != nil && tokenEntity.Password != nil {
					render.Status(req, http.StatusUnauthorized)
					render.JSON(res, req, dto.ErrorResponse{Error: "invalid or missing password"})
					return
				}
				render.Status(req, http.StatusUnauthorized)
				render.JSON(res, req, dto.ErrorResponse{Error: "invalid or expired token"})
				return
			}

			// Check if downloads are allowed
			if !tokenEntity.AllowDownload {
				render.Status(req, http.StatusForbidden)
				render.JSON(res, req, dto.ErrorResponse{Error: "downloads not permitted for this token"})
				return
			}

			// Validate embed access (prevents hotlinking if AllowEmbed is false)
			if !downloads.ValidateEmbedAccess(tokenEntity, req) {
				render.Status(req, http.StatusForbidden)
				render.JSON(res, req, dto.ErrorResponse{Error: "embedding not allowed for this token"})
				return
			}

			// Ensure token is valid for this UID
			allowed := slices.Contains(uids, uid)
			if !allowed {
				render.Status(req, http.StatusUnauthorized)
				render.JSON(res, req, dto.ErrorResponse{Error: "token not valid for this resource"})
				return
			}
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
		res.WriteHeader(http.StatusOK)
		res.Write(imageData)
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

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to update image",
				"Something went wrong, please try again later",
			)
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
		var fileImageUpload ImageUploadFile

		// Parse the multipart form in the request
		err := req.ParseMultipartForm(10 << 20) // limit your max input length!
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		fileImageUpload.FileName = req.FormValue("filename")

		file, _, err := req.FormFile("data")
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: err.Error()})
			return
		}
		defer file.Close()
		fileImageUpload.Data, err = io.ReadAll(file)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid file data"})
			return
		}

		if fileImageUpload.FileName == "" || len(fileImageUpload.Data) == 0 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		if fileImageUpload.Checksum != "" {
			hasher := sha1.New()
			hasher.Write(fileImageUpload.Data)
			calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))
			if fileImageUpload.Checksum != calculatedChecksum {
				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, dto.ErrorResponse{Error: "checksum mismatch"})
				return
			}
		}

		libvipsImg, err := libvips.NewImageFromBuffer(fileImageUpload.Data, libvips.DefaultLoadOptions())
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		defer libvipsImg.Close()
		imageEntity, err := createNewImageEntity(logger, fileImageUpload.FileName, libvipsImg)

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image data",
			)
			return
		}

		authUser, _ := libhttp.UserFromContext(req)
		imageEntity.UploadedByID = &authUser.Uid

		hasher := md5.New()
		if _, err := io.Copy(hasher, req.Body); err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		fileSize := int64(len(fileImageUpload.Data))
		imageEntity.ImageMetadata.FileSize = &fileSize
		imageEntity.ImageMetadata.Checksum = fileImageUpload.Checksum

		logger.Info("adding images to database", slog.String("id", imageEntity.Uid))
		dbCreateTx := db.Create(&imageEntity)

		if dbCreateTx.Error != nil {
			libhttp.ServerError(res, req, dbCreateTx.Error, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("starting image processing", slog.String("id", imageEntity.Uid))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		err = images.SaveImage(fileImageUpload.Data, imageEntity.Uid, imageEntity.ImageMetadata.FileName)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		err = workers.EnqueueImageProcessJob(workerJob)

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.Uid))

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, dto.ImageUploadResponse{Id: imageEntity.Uid})
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
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image data",
			)
			return
		}

		// Set the uploader
		authUser, _ := libhttp.UserFromContext(req)
		imageEntity.UploadedByID = &authUser.Uid

		hasher := sha1.New()
		if _, err := io.Copy(hasher, bytes.NewReader(fileBytes)); err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		fileSize := int64(len(fileBytes))
		imageEntity.ImageMetadata.FileSize = &fileSize
		imageEntity.ImageMetadata.Checksum = hex.EncodeToString(hasher.Sum(nil))

		logger.Info("adding image to database", slog.String("id", imageEntity.Uid))
		dbCreateTx := db.Create(&imageEntity)

		if dbCreateTx.Error != nil {
			libhttp.ServerError(res, req, dbCreateTx.Error, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("starting image processing", slog.String("id", imageEntity.Uid))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		err = images.SaveImage(fileBytes, imageEntity.Uid, imageEntity.ImageMetadata.FileName)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		err = workers.EnqueueImageProcessJob(workerJob)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.Uid))

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, dto.ImageUploadResponse{Id: imageEntity.Uid})
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
