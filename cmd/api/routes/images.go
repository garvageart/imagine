package routes

import (
	"slices"
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
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	"imagine/internal/dto"
	"imagine/internal/entities"
	"imagine/internal/downloads"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs/workers"
	libos "imagine/internal/os"
	"imagine/internal/uid"
)

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
		offsetStr := req.URL.Query().Get("offset")

		limit := 100
		offset := 0

		if limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
				limit = parsed
			}
		}

		if offsetStr != "" {
			if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
				offset = parsed
			}
		}

		var images []entities.Image

		if err := db.WithContext(req.Context()).Transaction(func(tx *gorm.DB) error {
			// Fetch non-deleted images
			err := tx.Preload("UploadedBy").
				Where("deleted_at IS NULL").
				Order("created_at DESC, uid DESC").
				Limit(limit).
				Offset(offset * limit).
				Find(&images).Error;

			if err != nil {
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

		count := len(items)
		response := dto.ImagesPage{
			Limit:  limit,
			Offset: offset,
			Count:  &count,
			Items:  items,
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
		if formatParam == "" && widthParam == "" && heightParam == "" && qualityParam == "" {
			imageData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
			if err != nil {
				render.Status(req, http.StatusInternalServerError)
				render.JSON(res, req, dto.ErrorResponse{Error: err.Error()})
				return
			}

			// Set cache headers for browser caching (1 year)
			res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			res.Header().Set("ETag", imgEnt.ImageMetadata.Checksum) // Use checksum as ETag
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
				res.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", time.Hour*24*365/time.Second))
				res.Header().Set("ETag", transformETag)
				res.WriteHeader(http.StatusNotModified)
				return
			}
		}

		// Read the original image file directly
		originalData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
		if err != nil {
			logger.Error("failed to read image data", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to read image data"})
			return
		}

		// Load into libvips
		libvipsImg, err := libvips.NewImageFromBuffer(originalData, libvips.DefaultLoadOptions())
		if err != nil {
			logger.Error("failed to create libvips image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to create libvips image"})
			return
		}
		defer libvipsImg.Close()

		err = libvipsImg.Autorot()
		if err != nil {
			logger.Error("failed to auto-rotate image", slog.Any("error", err))
		}

		// Resize with libvips
		err = libvipsImg.Resize(float64(width)/float64(libvipsImg.Width()), &libvips.ResizeOptions{
			Kernel: libvips.KernelLanczos3,
		})

		if err != nil {
			logger.Error("failed to resize image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to resize image"})
			return
		}

		var imageData []byte

		switch formatParam {
		case "webp":
			imageData, err = libvipsImg.WebpsaveBuffer(&libvips.WebpsaveBufferOptions{
				Q: int(quality),
			})
		case "png":
			imageData, err = libvipsImg.PngsaveBuffer(&libvips.PngsaveBufferOptions{
				Filter:      libvips.PngFilterNone,
				Interlace:   false,
				Palette:     false,
				Compression: int(quality),
			})
		case "jpg":
		case "jpeg":
			imageData, err = libvipsImg.JpegsaveBuffer(&libvips.JpegsaveBufferOptions{
				Q:         int(quality),
				Interlace: true,
			})
		case "avif":
		case "heif":
			imageData, err = libvipsImg.HeifsaveBuffer(&libvips.HeifsaveBufferOptions{
				Q:        int(quality),
				Bitdepth: 8,
				Effort:   5,
				Lossless: false,
			})
		default:
			imageData, err = libvipsImg.RawsaveBuffer(&libvips.RawsaveBufferOptions{
				Keep: libvips.KeepAll,
			})

		formatParam = string(libvipsImg.Format())
	}

	// Set response headers
	res.Header().Set("Content-Type", "image/"+formatParam)
	res.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", time.Hour*24*365/time.Second))
	res.Header().Set("ETag", transformETag)
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

		if err != nil {
			logger.Error("failed to encode image", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to encode image"})
			return
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
		res.WriteHeader(http.StatusOK)
		res.Write(imageData)
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
