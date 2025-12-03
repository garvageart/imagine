package routes

import (
	"slices"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/jobs/workers"
)

// handleImageProcessing processes image processing job requests
func handleImageProcessing(db *gorm.DB, logger *slog.Logger, body dto.WorkerJobCreateRequest, res http.ResponseWriter, req *http.Request) {
	command := string(body.Command)
	if command == "" {
		command = "all"
	}

	var count int64
	var err error

	// If UIDs provided and only one, treat as a single target
	if body.Uids != nil && len(*body.Uids) == 1 {
		var img entities.Image
		if err = db.Where("uid = ?", (*body.Uids)[0]).First(&img).Error; err != nil {
			render.Status(req, http.StatusNotFound)
			render.JSON(res, req, dto.ErrorResponse{Error: "Image not found"})
			return
		}

		job := &workers.ImageProcessJob{Image: img}
		_, err := jobs.Enqueue(db, workers.TopicImageProcess, job, nil, &img.Uid)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to enqueue job"})
			return
		}

		count := 1
		render.Status(req, http.StatusAccepted)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "thumbnail generation job enqueued", Count: &count})
		return
	}

	switch command {
	case "missing":
		// Find UIDs of images missing thumbhash
		var thumbhashMissing []string
		if err = db.Model(&entities.Image{}).Where("image_metadata->>'thumbhash' IS NULL").Pluck("uid", &thumbhashMissing).Error; err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to identify images missing thumbhash"})
			return
		}

		// Find UIDs of images with missing permanent transforms
		transformsMissing, err := findMissingTransforms(db, logger)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to identify images with missing permanent transforms"})
			return
		}

		// Combine and deduplicate UIDs
		uniqueUids := make(map[string]struct{})
		for _, uid := range thumbhashMissing {
			uniqueUids[uid] = struct{}{}
		}

		for _, uid := range transformsMissing {
			uniqueUids[uid] = struct{}{}
		}

		var finalUids []string
		for uid := range uniqueUids {
			finalUids = append(finalUids, uid)
		}

		count = int64(len(finalUids))

		if count == 0 {
			zeroCount := 0
			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "no images to process", Count: &zeroCount})
			return
		}

		// Enqueue jobs in background
		go func(targetUids []string) {
			// Batch process targetUids
			batchSize := 100
			for i := 0; i < len(targetUids); i += batchSize {
				end := min(i+batchSize, len(targetUids))
				batchUids := targetUids[i:end]

				var imgs []entities.Image
				if err := db.Where("uid IN ?", batchUids).Find(&imgs).Error; err != nil {
					logger.Error("failed to fetch images for batch processing", slog.Any("error", err))
					continue
				}

				for _, img := range imgs {
					job := &workers.ImageProcessJob{Image: img}
					_, _ = jobs.Enqueue(db, workers.TopicImageProcess, job, nil, &img.Uid)
				}
			}
			logger.Info("image processing jobs enqueued", "command", "missing", "count", count)
		}(finalUids)

		jobCount := int(count)
		render.Status(req, http.StatusAccepted)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{
			Message: fmt.Sprintf("thumbnail generation jobs enqueued (%s)", command),
			Count:   &jobCount,
		})
	}
}

// handleXMPGeneration processes XMP sidecar file generation job requests
func handleXMPGeneration(db *gorm.DB, logger *slog.Logger, body dto.WorkerJobCreateRequest, res http.ResponseWriter, req *http.Request) {
	command := string(body.Command)
	if command == "" {
		command = "all"
	}

	var count int64
	var err error
	var uidsWithoutXMP []string

	if body.Uids != nil && len(*body.Uids) == 1 {
		var img entities.Image
		if err := db.Where("uid = ?", (*body.Uids)[0]).First(&img).Error; err != nil {
			render.Status(req, http.StatusNotFound)
			render.JSON(res, req, dto.ErrorResponse{Error: "Image not found"})
			return
		}

		job := &workers.XMPGenerationJob{Image: img}
		_, err := jobs.Enqueue(db, workers.TopicXMPGeneration, job, nil, &img.Uid)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to enqueue job"})
			return
		}

		count := 1
		render.Status(req, http.StatusAccepted)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "XMP sidecar generation job enqueued", Count: &count})
		return
	}

	switch command {
	case "missing":
		// Scan filesystem first to find UIDs without XMP files
		libraryPath := images.Directory
		entries, err := os.ReadDir(libraryPath)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to read library directory"})
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			uid := entry.Name()
			dirPath := filepath.Join(libraryPath, uid)

			// Check if any .xmp file exists in this directory
			files, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}

			hasXMP := false
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".xmp") {
					hasXMP = true
					break
				}
			}

			if !hasXMP {
				uidsWithoutXMP = append(uidsWithoutXMP, uid)
			}
		}

		count = int64(len(uidsWithoutXMP))
	case "all":
		// All images will get XMP files (regenerate existing ones)
		err = db.Model(&entities.Image{}).Count(&count).Error
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to count images"})
			return
		}
	// 'single' replaced by `uids`: handled above if provided.
	default:
		render.Status(req, http.StatusBadRequest)
		render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unknown command: %s", command)})
		return
	}

	if count == 0 {
		zeroCount := 0
		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "no images to process", Count: &zeroCount})
		return
	}

	go func(cmd string, uids []string) {
		processed := 0

		if cmd == "missing" {
			if len(uids) == 0 {
				logger.Info("XMP sidecar generation completed", "command", cmd, "enqueued", 0)
				return
			}

			// Process in batches of 100 UIDs
			for i := 0; i < len(uids); i += 100 {
				end := min(i+100, len(uids))
				batch := uids[i:end]

				var batchImgs []entities.Image
				if err := db.Where("uid IN ?", batch).Find(&batchImgs).Error; err != nil {
					logger.Error("failed to fetch images", "error", err)
					continue
				}

				for _, img := range batchImgs {
					job := &workers.XMPGenerationJob{Image: img}
					if _, err := jobs.Enqueue(db, workers.TopicXMPGeneration, job, nil, &img.Uid); err != nil {
						logger.Error("failed to enqueue XMP job", "image_uid", img.Uid, "error", err)
					} else {
						processed++
					}
				}
			}
		} else {
			query := db.Session(&gorm.Session{})
			var imgs []entities.Image
			query.FindInBatches(&imgs, 100, func(tx *gorm.DB, batch int) error {
				for _, img := range imgs {
					job := &workers.XMPGenerationJob{Image: img}
					if _, err := jobs.Enqueue(db, workers.TopicXMPGeneration, job, nil, &img.Uid); err != nil {
						logger.Error("failed to enqueue XMP job", "image_uid", img.Uid, "error", err)
					} else {
						processed++
					}
				}
				return nil
			})
		}

		logger.Info("XMP sidecar generation jobs enqueued", "command", cmd, "enqueued", processed)
	}(command, uidsWithoutXMP)

	jobCount := int(count)
	render.Status(req, http.StatusAccepted)
	render.JSON(res, req, dto.WorkerJobEnqueueResponse{
		Message: fmt.Sprintf("XMP sidecar generation jobs enqueued (%s)", command),
		Count:   &jobCount,
	})
}

// handleExifProcessing processes EXIF extraction job requests
func handleExifProcessing(db *gorm.DB, logger *slog.Logger, body dto.WorkerJobCreateRequest, res http.ResponseWriter, req *http.Request) {
	command := string(body.Command)
	if command == "" {
		command = "all"
	}

	var count int64
	var err error

	if body.Uids != nil && len(*body.Uids) == 1 {
		var img entities.Image
		if err := db.Where("uid = ?", (*body.Uids)[0]).First(&img).Error; err != nil {
			render.Status(req, http.StatusNotFound)
			render.JSON(res, req, dto.ErrorResponse{Error: "Image not found"})
			return
		}

		job := &workers.ExifProcessJob{Image: img}
		_, err := jobs.Enqueue(db, workers.TopicExifProcess, job, nil, &img.Uid)
		if err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to enqueue job"})
			return
		}

		c := 1
		render.Status(req, http.StatusAccepted)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "EXIF processing job enqueued", Count: &c})
		return
	}

	// looool i hate this so bad
	missingQuery := "exif IS NULL OR (exif IS NOT NULL AND (exif->>'aperture' IS NULL OR exif->>'date_time' IS NULL OR exif->>'date_time_original' IS NULL OR exif->>'exif_version' IS NULL OR exif->>'exposure_time' IS NULL OR exif->>'exposure_value' IS NULL OR exif->>'f_number' IS NULL OR exif->>'flash' IS NULL OR exif->>'focal_length' IS NULL OR exif->>'iso' IS NULL OR exif->>'latitude' IS NULL OR exif->>'lens_model' IS NULL OR exif->>'longitude' IS NULL OR exif->>'make' IS NULL OR exif->>'model' IS NULL OR exif->>'modify_date' IS NULL OR exif->>'orientation' IS NULL OR exif->>'rating' IS NULL OR exif->>'resolution' IS NULL OR exif->>'software' IS NULL OR exif->>'white_balance' IS NULL))"
	switch command {
	case "missing":
		// images without exif
		err = db.Model(&entities.Image{}).Where(missingQuery).Count(&count).Error
	case "all":
		err = db.Model(&entities.Image{}).Count(&count).Error
	// 'single' replaced by `uids`: handled above if provided.
	default:
		render.Status(req, http.StatusBadRequest)
		render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unknown command: %s", command)})
		return
	}

	if err != nil {
		render.Status(req, http.StatusInternalServerError)
		render.JSON(res, req, dto.ErrorResponse{Error: "Failed to count images"})
		return
	}

	if count == 0 {
		zeroCount := 0
		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.WorkerJobEnqueueResponse{Message: "no images to process", Count: &zeroCount})
		return
	}

	go func(cmd string) {
		var query *gorm.DB
		if cmd == "missing" {
			query = db.Where(missingQuery)
		} else {
			query = db.Session(&gorm.Session{})
		}

		var imgs []entities.Image
		query.FindInBatches(&imgs, 100, func(tx *gorm.DB, batch int) error {
			for _, img := range imgs {
				job := &workers.ExifProcessJob{Image: img}
				_, _ = jobs.Enqueue(db, workers.TopicExifProcess, job, nil, &img.Uid)
			}
			return nil
		})

		logger.Info("exif processing jobs enqueued", "command", cmd, "count", count)
	}(command)

	jobCount := int(count)
	render.Status(req, http.StatusAccepted)
	render.JSON(res, req, dto.WorkerJobEnqueueResponse{
		Message: fmt.Sprintf("EXIF processing jobs enqueued (%s)", command),
		Count:   &jobCount,
	})
}

// containsUid is a helper for checking if a slice contains a UID.
func containsUid(s []string, e string) bool {
	return slices.Contains(s, e)
}

// checkMissingTransforms checks if a given image path's transform is missing from the cache.
// If missing, it adds the image's UID to the missingUids slice.
func checkMissingTransforms(img entities.Image, transformPath string, missingUids *[]string, logger *slog.Logger) error {
	if transformPath == "" {
		return nil // No path to check
	}

	if img.ImageMetadata == nil || img.ImageMetadata.Checksum == "" || img.ImageMetadata.FileType == "" {
		logger.Debug("skipping transform check due to incomplete metadata", slog.String("uid", img.Uid), slog.String("path", transformPath))
		return nil // Cannot check without complete metadata
	}

	params, err := imageops.ParseTransformParams(transformPath)
	if err != nil {
		logger.Warn("failed to parse transform params", slog.String("uid", img.Uid), slog.String("path", transformPath), slog.Any("error", err))
		// If params cannot be parsed, treat as missing/invalid and enqueue for reprocessing
		if !containsUid(*missingUids, img.Uid) {
			*missingUids = append(*missingUids, img.Uid)
		}
		return fmt.Errorf("failed to parse transform params for %s: %w", transformPath, err)
	}

	transformEtag := *imageops.CreateTransformEtag(img, params)
	ext := params.Format
	if ext == "" {
		ext = img.ImageMetadata.FileType // Fallback if format isn't specified in path
	}

	_, exists, err := images.FindCachedTransform(img.Uid, transformEtag, ext)
	if err != nil {
		logger.Error("failed to check for cached transform", slog.String("uid", img.Uid), slog.String("path", transformPath), slog.Any("error", err))
		// Treat error as missing
		if !containsUid(*missingUids, img.Uid) {
			*missingUids = append(*missingUids, img.Uid)
		}
		return fmt.Errorf("failed to check cached transform for %s: %w", transformPath, err)
	}

	if !exists {
		if !containsUid(*missingUids, img.Uid) {
			*missingUids = append(*missingUids, img.Uid)
		}
	}
	return nil
}

// findMissingTransforms finds UIDs of images that have permanent paths (thumbnail/preview) defined
// but the corresponding cached transform files are missing.
func findMissingTransforms(db *gorm.DB, logger *slog.Logger) ([]string, error) {
	var allImages []entities.Image
	var err error
	// Fetch uid, image_paths, and image_metadata needed for cache key calculation
	if err = db.Select("uid", "image_paths", "image_metadata").Find(&allImages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all images: %w", err)
	}

	var missing []string
	for _, img := range allImages {
		// Check thumbnail path
		err = checkMissingTransforms(img, img.ImagePaths.Thumbnail, &missing, logger)
		if err != nil {
			return nil, err
		}

		// Check preview path (only if not already added by thumbnail check)
		if !containsUid(missing, img.Uid) { // Only check preview if not already identified as missing
			err = checkMissingTransforms(img, img.ImagePaths.Preview, &missing, logger)
			if err != nil {
				return nil, err
			}
		}
	}

	return missing, nil
}

// JobsRouter returns a router with admin-only job endpoints.
// It applies AuthMiddleware and AdminMiddleware internally so it can be
// mounted anywhere (we mount it under /admin/jobs in api.go).
func JobsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(libhttp.AuthMiddleware(db, logger))
	r.Use(libhttp.AdminMiddleware)

	r.Get("/count", func(res http.ResponseWriter, req *http.Request) {
		count := jobs.GetRunningJobs()
		render.Status(req, http.StatusOK)
		render.JSON(res, req, map[string]int{"running": count})
	})

	r.Get("/stats", func(res http.ResponseWriter, req *http.Request) {
		stats := jobs.GetCounts()
		render.Status(req, http.StatusOK)
		render.JSON(res, req, stats)
	})

	// Atomic snapshot for UI bootstrap: active jobs, counters, and next event cursor
	r.Get("/snapshot", func(res http.ResponseWriter, req *http.Request) {
		// active jobs
		activeMap := jobs.GetAllJobs()
		type ActiveBrief struct {
			Id     string `json:"id"`
			Topic  string `json:"topic"`
			Status string `json:"status"`
		}

		active := make([]ActiveBrief, 0, len(activeMap))
		for id, j := range activeMap {
			active = append(active, ActiveBrief{Id: id, Topic: j.Topic(), Status: j.GetStatus()})
		}

		// counters
		stats := jobs.GetCounts()

		// next cursor from SSE broker via context key; the server wires it under /events router.
		// We can’t access the broker instance directly here without a global. For now, omit nextCursor in response.
		// Frontend can fetch /events/since with cursor=0 to bootstrap missed events if needed.
		snap := map[string]any{
			"active":           active,
			"running_by_topic": stats.RunningByTopic,
			"queued_by_topic":  stats.QueuedByTopic,
			// nextCursor intentionally omitted due to scope isolation
		}
		render.Status(req, http.StatusOK)
		render.JSON(res, req, snap)
	})

	// GET /jobs: list worker jobs (DB-backed). Supports ?status=&topic=&limit=&page=
	r.Get("/", func(res http.ResponseWriter, req *http.Request) {
		status := req.URL.Query().Get("status")
		topic := req.URL.Query().Get("topic")

		limit := 25
		page := 0
		if q := req.URL.Query().Get("limit"); q != "" {
			fmt.Sscanf(q, "%d", &limit)
		}
		if q := req.URL.Query().Get("page"); q != "" {
			fmt.Sscanf(q, "%d", &page)
		}

		var ents []entities.WorkerJob
		query := db.Session(&gorm.Session{})
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if topic != "" {
			query = query.Where("topic = ?", topic)
		}

		var total int64
		if err := query.Model(&entities.WorkerJob{}).Count(&total).Error; err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to list jobs"})
			return
		}

		if err := query.Order("enqueued_at desc").Limit(limit).Offset(page * limit).Find(&ents).Error; err != nil {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Failed to list jobs"})
			return
		}

		items := make([]dto.WorkerJob, 0, len(ents))
		for _, e := range ents {
			items = append(items, e.DTO())
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.WorkerJobsResponse{Items: items, Total: int(total)})
	})

	// GET /workers: list registered worker types
	r.Get("/workers", func(res http.ResponseWriter, req *http.Request) {
		stats := jobs.GetCounts()
		workersList := jobs.GetAllWorkers()

		items := make([]dto.WorkerInfo, 0, len(workersList))
		for _, w := range workersList {
			count := 0
			if v := stats.RunningByTopic[w.Topic]; v > 0 {
				count = v
			}
			cptr := count
			items = append(items, dto.WorkerInfo{Concurrency: w.Concurrency, Count: &cptr, DisplayName: w.DisplayName, Name: w.Name})
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.WorkersListResponse{Items: items})
	})

	r.Post("/workers", func(res http.ResponseWriter, req *http.Request) {
		var body dto.WorkerRegisterRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid body"})
			return
		}

		if body.Name == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Name required"})
			return
		}

		// Apply optional concurrency update
		if body.Concurrency != nil {
			jobs.SetConcurrency(body.Name, *body.Concurrency)
		}

		stats := jobs.GetCounts()
		count := 0
		if v := stats.RunningByTopic[body.Name]; v > 0 {
			count = v
		}
		cptr := count

		ci := dto.WorkerInfo{Concurrency: jobs.GetConcurrency(body.Name), Count: &cptr, DisplayName: body.Name, Name: body.Name}

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, ci)
	})

	r.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var body dto.WorkerJobCreateRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid body"})
			return
		}

		switch body.Type {
		case workers.JobTypeImageProcess:
			handleImageProcessing(db, logger, body, res, req)
			return

		case workers.JobTypeXMPGeneration:
			handleXMPGeneration(db, logger, body, res, req)
			return

		case workers.JobTypeExifProcess:
			handleExifProcessing(db, logger, body, res, req)
			return

		default:
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unsupported job type: %s", body.Type)})
			return
		}
	})

	r.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		var ent entities.WorkerJob

		if err := db.Where("uid = ?", uid).First(&ent).Error; err == nil {
			resp := ent.DTO()

			render.Status(req, http.StatusOK)
			render.JSON(res, req, resp)
			return
		}

		// Fallback to in-memory active jobs snapshot
		all := jobs.GetAllJobs()
		if j, ok := all[uid]; ok {
			resp := map[string]any{
				"uid":    uid,
				"type":   j.Topic(),
				"topic":  j.Topic(),
				"status": j.GetStatus(),
			}

			render.Status(req, http.StatusOK)
			render.JSON(res, req, resp)
			return
		}

		render.Status(req, http.StatusNotFound)
		render.JSON(res, req, dto.ErrorResponse{Error: "Job not found"})
	})

	r.Delete("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		all := jobs.GetAllJobs()
		if j, ok := all[uid]; ok {
			j.SetStatus("cancelled")
			delete(all, uid)
			_ = jobs.UpdateWorkerJobStatus(db, uid, "cancelled", nil, nil, nil, nil)

			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.MessageResponse{Message: "cancelled"})
			return
		}

		if err := jobs.UpdateWorkerJobStatus(db, uid, "cancelled", nil, nil, nil, nil); err == nil {
			render.Status(req, http.StatusOK)
			render.JSON(res, req, dto.MessageResponse{Message: "cancelled"})
			return
		}

		render.Status(req, http.StatusNotFound)
		render.JSON(res, req, dto.ErrorResponse{Error: "Job not found"})
	})

	r.Post("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		render.Status(req, http.StatusNotImplemented)
		render.JSON(res, req, dto.ErrorResponse{Error: "Retry not implemented"})
	})

	r.Post("/types/{type}/stop", func(res http.ResponseWriter, req *http.Request) {
		jobType := chi.URLParam(req, "type")
		if jobType == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Job type required"})
			return
		}

		topic := jobType

		all := jobs.GetAllJobs()
		cancelled := 0
		for id, j := range all {
			if j.Topic() == topic {
				j.SetStatus("cancelled")
				delete(all, id)
				cancelled++
			}
		}

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: fmt.Sprintf("stopped %d jobs of type %s", cancelled, jobType)})
	})

	r.Put("/types/{type}/concurrency", func(res http.ResponseWriter, req *http.Request) {
		jobType := chi.URLParam(req, "type")
		if jobType == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Job type required"})
			return
		}

		var body struct {
			Concurrency int `json:"concurrency"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid body"})
			return
		}

		if body.Concurrency < 1 || body.Concurrency > 100 {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Concurrency must be between 1 and 100"})
			return
		}

		topic := jobType

		jobs.SetConcurrency(topic, body.Concurrency)
		logger.Info("concurrency updated",
			slog.String("jobType", jobType),
			slog.String("topic", topic),
			slog.Int("concurrency", body.Concurrency),
		)

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: fmt.Sprintf("concurrency for %s set to %d", jobType, body.Concurrency)})
	})

	return r
}
