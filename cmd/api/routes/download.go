package routes

import (
    "archive/zip"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    "gorm.io/gorm"

    "imagine/internal/downloads"
    "imagine/internal/dto"
    "imagine/internal/entities"
    "imagine/internal/images"
    "imagine/internal/utils"
)

// writeImagesToZip queries images for the given uids and writes them into the provided zip.Writer
// in the order of the provided uids slice. Missing or unreadable files are skipped and logged.
func writeImagesToZip(ctx context.Context, db *gorm.DB, logger *slog.Logger, zw *zip.Writer, uids []string) error {
    if len(uids) == 0 {
        return nil
    }

    var imgs []entities.Image
    if err := db.WithContext(ctx).Where("uid IN ? AND deleted_at IS NULL", uids).Find(&imgs).Error; err != nil {
        return err
    }

    imgMap := make(map[string]entities.Image, len(imgs))
    for _, im := range imgs {
        imgMap[im.Uid] = im
    }

    for _, uid := range uids {
        imageEntity, ok := imgMap[uid]
        if !ok {
            logger.Warn("image not found for export", slog.String("uid", uid))
            continue
        }

        diskPath := images.GetImagePath(imageEntity.Uid, imageEntity.ImageMetadata.FileName)
        f, err := os.Open(diskPath)
        if err != nil {
            logger.Error("failed to open image file for export", slog.Any("error", err), slog.String("path", diskPath))
            continue
        }

    safeName := filepath.Base(imageEntity.ImageMetadata.FileName)
    // Use the original filename inside the ZIP (do not prefix with UID)
    zipFileName := safeName

        fileHeader := &zip.FileHeader{
            Name:   zipFileName,
            Method: zip.Deflate,
        }

        // try to set a meaningful mod time
        fileHeader.Modified = imageEntity.UpdatedAt

        w, err := zw.CreateHeader(fileHeader)
        if err != nil {
            f.Close()
            logger.Error("failed to create zip entry", slog.Any("error", err))
            continue
        }

        if _, err := io.Copy(w, f); err != nil {
            f.Close()
            logger.Error("failed to write image to zip", slog.Any("error", err))
            continue
        }

        f.Close()
    }

    return nil
}

// streamZipResponse streams a zip of the given uids to the http.ResponseWriter using an io.Pipe
// to avoid buffering the entire archive in memory.
func streamZipResponse(res http.ResponseWriter, req *http.Request, db *gorm.DB, logger *slog.Logger, uids []string, filename string) {
	if filename == "" {
		filename = fmt.Sprintf("%s_export_%s.zip", utils.AppName, time.Now().Format("20060102T150405"))
	}

	res.Header().Set("Content-Type", "application/octet-stream")
	// Provide both filename and filename* (UTF-8) to improve browser compatibility
	base := filepath.Base(filename)
	res.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", base, url.PathEscape(filename)))
    pr, pw := io.Pipe()

    go func() {
        // Ensure any writer-side errors are propagated to the reader via CloseWithError
        zw := zip.NewWriter(pw)
        if err := writeImagesToZip(req.Context(), db, logger, zw, uids); err != nil {
            logger.Error("error while creating zip", slog.Any("error", err))
            _ = zw.Close()
            _ = pw.CloseWithError(err)
            return
        }

        if err := zw.Close(); err != nil {
            logger.Error("failed to close zip writer", slog.Any("error", err))
            _ = pw.CloseWithError(err)
            return
        }

        // Normal close
        _ = pw.Close()
    }()

    // flush headers early if possible so the client sees Content-Disposition
    if f, ok := res.(http.Flusher); ok {
        f.Flush()
    }

    // copy the pipe to the response writer (this will block until pw is closed)
    if _, err := io.Copy(res, pr); err != nil {
        logger.Error("failed to stream zip to client", slog.Any("error", err))
        return
    }
}

func DownloadRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
    router := chi.NewRouter()

    router.Post("/sign", func(res http.ResponseWriter, req *http.Request) {
        var body dto.SignDownloadRequest

        if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
            return
        }

        // require uids present
        if body.Uids == nil || len(*body.Uids) == 0 {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "uids is required"})
            return
        }

        var ttl time.Duration
        if body.ExpiresIn != nil && *body.ExpiresIn > 0 {
            ttl = time.Duration(*body.ExpiresIn) * time.Second
        } else {
            ttl = 15 * time.Minute
        }

        // Build token options
        opts := downloads.TokenOptions{
            TTL:           ttl,
            AllowDownload: body.AllowDownload == nil || *body.AllowDownload, // Default: true
            AllowEmbed:    body.AllowEmbed != nil && *body.AllowEmbed,       // Default: false
            ShowMetadata:  body.ShowMetadata == nil || *body.ShowMetadata,   // Default: true
        }

        if body.Password != nil {
            opts.Password = *body.Password
        }
        if body.Description != nil {
            opts.Description = *body.Description
        }

		token, err := downloads.CreateTokenWithOptions(db, *body.Uids, opts)
		if err != nil {
			logger.Error("failed to create download token", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to create download token"})
			return
		}

		var tokenEntity entities.DownloadToken
		if err := db.First(&tokenEntity, "uid = ?", token).Error; err != nil {
			logger.Error("failed to fetch created token", slog.Any("error", err))
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "failed to retrieve token details"})
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, tokenEntity.DTO())
	}) 
	
    router.Post("/", func(res http.ResponseWriter, req *http.Request) {
        tokenParam := req.URL.Query().Get("token")
        passwordParam := req.URL.Query().Get("password")

        if tokenParam == "" {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "missing token query param"})
            return
        }

        allowedUIDs, tokenEntity, ok := downloads.ValidateTokenWithPassword(db, tokenParam, passwordParam)
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

        if !tokenEntity.AllowDownload {
            render.Status(req, http.StatusForbidden)
            render.JSON(res, req, dto.ErrorResponse{Error: "downloads not permitted for this token"})
            return
        }

        var body dto.DownloadRequest

        if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
            return
        }

        if len(body.Uids) == 0 {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "uids is required"})
            return
        }

        allowedMap := make(map[string]bool, len(allowedUIDs))
        for _, uid := range allowedUIDs {
            allowedMap[uid] = true
        }

        for _, uid := range body.Uids {
            if !allowedMap[uid] {
                render.Status(req, http.StatusUnauthorized)
                render.JSON(res, req, dto.ErrorResponse{Error: "token does not authorize all requested UIDs"})
                return
            }
        }

        // Stream ZIP using a pipe to avoid buffering the entire archive in memory
        var filename string
        if body.Filename != nil {
            filename = *body.Filename
        }
        streamZipResponse(res, req, db, logger, body.Uids, filename)
    })

    return router
}