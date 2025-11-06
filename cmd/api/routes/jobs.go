package routes

import (
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
    "imagine/internal/images"
    "imagine/internal/jobs"
    "imagine/internal/jobs/workers"
)

// handleThumbnailGeneration processes thumbnail generation job requests
func handleThumbnailGeneration(db *gorm.DB, logger *slog.Logger, body dto.JobCreateRequest, res http.ResponseWriter, req *http.Request) {
    command := "all" // default
    if body.Command != nil {
        command = string(*body.Command)
    }

    var count int64
    var err error

    switch command {
    case "missing":
        // Only images without thumbnails/thumbhash
        err = db.Model(&entities.Image{}).Where("image_metadata->>'thumbhash' IS NULL").Count(&count).Error
    case "all":
        // All images
        err = db.Model(&entities.Image{}).Count(&count).Error
    case "single":
        // Single image - requires image_uid
        if body.ImageUid == nil || *body.ImageUid == "" {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "image_uid required for single command"})
            return
        }
        var img entities.Image
        if err := db.Where("uid = ?", *body.ImageUid).First(&img).Error; err != nil {
            render.Status(req, http.StatusNotFound)
            render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
            return
        }

        job := &workers.ImageProcessJob{Image: img}
        if err := workers.EnqueueImageProcessJob(job); err != nil {
            render.Status(req, http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to enqueue job"})
            return
        }

        count := 1
        render.Status(req, http.StatusAccepted)
        render.JSON(res, req, dto.JobEnqueueResponse{Message: "thumbnail generation job enqueued", Count: &count})
        return
    default:
        render.Status(req, http.StatusBadRequest)
        render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unknown command: %s", command)})
        return
    }

    if err != nil {
        render.Status(req, http.StatusInternalServerError)
        render.JSON(res, req, dto.ErrorResponse{Error: "failed to count images"})
        return
    }

    if count == 0 {
        zeroCount := 0
        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.JobEnqueueResponse{Message: "no images to process", Count: &zeroCount})
        return
    }

    // Enqueue jobs in background
    go func(cmd string) {
        var query *gorm.DB
        if cmd == "missing" {
            query = db.Where("image_metadata->>'thumbhash' IS NULL")
        } else {
            query = db.Session(&gorm.Session{})
        }

        var imgs []entities.Image
        query.FindInBatches(&imgs, 100, func(tx *gorm.DB, batch int) error {
            for _, img := range imgs {
                job := &workers.ImageProcessJob{Image: img}
                _ = workers.EnqueueImageProcessJob(job)
            }
            return nil
        })

        logger.Info("thumbnail generation jobs enqueued", "command", cmd, "count", count)
    }(command)

    jobCount := int(count)
    render.Status(req, http.StatusAccepted)
    render.JSON(res, req, dto.JobEnqueueResponse{
        Message: fmt.Sprintf("thumbnail generation jobs enqueued (%s)", command),
        Count:   &jobCount,
    })
}

// handleXMPGeneration processes XMP sidecar file generation job requests
func handleXMPGeneration(db *gorm.DB, logger *slog.Logger, body dto.JobCreateRequest, res http.ResponseWriter, req *http.Request) {
    command := "all" // default
    if body.Command != nil {
        command = string(*body.Command)
    }

    var count int64
    var err error
    var uidsWithoutXMP []string

    switch command {
    case "missing":
        // Scan filesystem first to find UIDs without XMP files
        libraryPath := images.Directory
        entries, err := os.ReadDir(libraryPath)
        if err != nil {
            render.Status(req, http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to read library directory"})
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
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to count images"})
            return
        }
    case "single":
        if body.ImageUid == nil || *body.ImageUid == "" {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "image_uid required for single command"})
            return
        }
        var img entities.Image
        if err := db.Where("uid = ?", *body.ImageUid).First(&img).Error; err != nil {
            render.Status(req, http.StatusNotFound)
            render.JSON(res, req, dto.ErrorResponse{Error: "image not found"})
            return
        }

        job := &workers.XMPGenerationJob{Image: img}
        if err := workers.EnqueueXMPGenerationJob(job); err != nil {
            render.Status(req, http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to enqueue job"})
            return
        }

        count := 1
        render.Status(req, http.StatusAccepted)
        render.JSON(res, req, dto.JobEnqueueResponse{Message: "XMP sidecar generation job enqueued", Count: &count})
        return
    default:
        render.Status(req, http.StatusBadRequest)
        render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unknown command: %s", command)})
        return
    }

    if count == 0 {
        zeroCount := 0
        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.JobEnqueueResponse{Message: "no images to process", Count: &zeroCount})
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
                end := min(i + 100, len(uids))
                batch := uids[i:end]
                
                var batchImgs []entities.Image
                if err := db.Where("uid IN ?", batch).Find(&batchImgs).Error; err != nil {
                    logger.Error("failed to fetch images", "error", err)
                    continue
                }
                
                for _, img := range batchImgs {
                    job := &workers.XMPGenerationJob{Image: img}
                    if err := workers.EnqueueXMPGenerationJob(job); err != nil {
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
                    if err := workers.EnqueueXMPGenerationJob(job); err != nil {
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
    render.JSON(res, req, dto.JobEnqueueResponse{
        Message: fmt.Sprintf("XMP sidecar generation jobs enqueued (%s)", command),
        Count:   &jobCount,
    })
}

// JobsRouter returns a router with admin-only job endpoints.
// It applies AuthMiddleware and AdminMiddleware internally so it can be
// mounted anywhere (we mount it under /admin/jobs in api.go).
func JobsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
    r := chi.NewRouter()

    r.Use(libhttp.AuthMiddleware(db, logger))
    r.Use(libhttp.AdminMiddleware)

    r.Post("/scheduler/start", func(res http.ResponseWriter, req *http.Request) {
        if err := jobs.Start(); err != nil {
            render.Status(req, http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to start scheduler"})
            return
        }
        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.MessageResponse{Message: "started"})
    })

    r.Post("/scheduler/shutdown", func(res http.ResponseWriter, req *http.Request) {
        if err := jobs.Shutdown(); err != nil {
            render.Status(req, http.StatusInternalServerError)
            render.JSON(res, req, dto.ErrorResponse{Error: "failed to shutdown scheduler"})
            return
        }
        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.MessageResponse{Message: "shutdown"})
    })

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
            "active":         active,
            "running_by_topic": stats.RunningByTopic,
            "queued_by_topic":  stats.QueuedByTopic,
            // nextCursor intentionally omitted due to scope isolation
        }
        render.Status(req, http.StatusOK)
        render.JSON(res, req, snap)
    })

    r.Get("/", func(res http.ResponseWriter, req *http.Request) {
        items := []dto.JobInfo{
            {Id: "thumbnailGeneration", Topic: "Thumbnail Generation", Status: "idle"},
            {Id: "xmpGeneration", Topic: "XMP Sidecar Generation", Status: "idle"},
        }
        render.Status(req, http.StatusOK)
        render.JSON(res, req, dto.JobListResponse{Items: items})
    })

    r.Post("/", func(res http.ResponseWriter, req *http.Request) {
        var body dto.JobCreateRequest
        if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "invalid body"})
            return
        }

        switch body.Type {
        case "thumbnailGeneration":
            handleThumbnailGeneration(db, logger, body, res, req)
            return

        case "xmpGeneration":
            handleXMPGeneration(db, logger, body, res, req)
            return

        default:
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: fmt.Sprintf("unsupported job type: %s", body.Type)})
            return
        }
    })

    r.Get("/{id}", func(res http.ResponseWriter, req *http.Request) {
        id := chi.URLParam(req, "id")
        all := jobs.GetAllJobs()
        if j, ok := all[id]; ok {
            render.Status(req, http.StatusOK)
            render.JSON(res, req, dto.JobInfo{Id: id, Topic: j.Topic(), Status: j.GetStatus()})
            return
        }
        render.Status(req, http.StatusNotFound)
        render.JSON(res, req, dto.ErrorResponse{Error: "job not found"})
    })

    r.Delete("/{id}", func(res http.ResponseWriter, req *http.Request) {
        id := chi.URLParam(req, "id")
        all := jobs.GetAllJobs()
        if j, ok := all[id]; ok {
            j.SetStatus("cancelled")
            // delete locally only (snapshot) — job handler will remove itself upon completion
            delete(all, id)
            render.Status(req, http.StatusOK)
            render.JSON(res, req, dto.MessageResponse{Message: "cancelled"})
            return
        }

        render.Status(req, http.StatusNotFound)
        render.JSON(res, req, dto.ErrorResponse{Error: "job not found"})
    })

    r.Post("/{id}", func(res http.ResponseWriter, req *http.Request) {
        render.Status(req, http.StatusNotImplemented)
        render.JSON(res, req, dto.ErrorResponse{Error: "retry not implemented"})
    })

    r.Post("/types/{type}/stop", func(res http.ResponseWriter, req *http.Request) {
        jobType := chi.URLParam(req, "type")
        if jobType == "" {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "job type required"})
            return
        }

        topic := jobType
        switch jobType {
        case "thumbnailGeneration":
            topic = workers.TopicImageProcess
        }

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
            render.JSON(res, req, dto.ErrorResponse{Error: "job type required"})
            return
        }

        var body struct {
            Concurrency int `json:"concurrency"`
        }
        if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "invalid body"})
            return
        }

        if body.Concurrency < 1 || body.Concurrency > 100 {
            render.Status(req, http.StatusBadRequest)
            render.JSON(res, req, dto.ErrorResponse{Error: "concurrency must be between 1 and 100"})
            return
        }

        topic := jobType
        switch jobType {
        case "thumbnailGeneration":
            topic = workers.TopicImageProcess
        }

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
