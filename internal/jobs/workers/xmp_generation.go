package workers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/trimmer-io/go-xmp/models/dc"
	"github.com/trimmer-io/go-xmp/models/exif"
	"github.com/trimmer-io/go-xmp/models/tiff"
	xmpbase "github.com/trimmer-io/go-xmp/models/xmp_base"
	"github.com/trimmer-io/go-xmp/xmp"
	"gorm.io/gorm"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/utils"
)

const (
	JobTypeXMPGeneration = "xmp_generation"
	TopicXMPGeneration   = JobTypeXMPGeneration
)

type XMPGenerationJob struct {
	Image entities.Image
}

// NewXMPWorker creates a worker that generates XMP sidecar files
func NewXMPWorker(db *gorm.DB, logger *slog.Logger, wsBroker *libhttp.WSBroker) *jobs.Worker {
	return jobs.NewWorker(JobTypeXMPGeneration, TopicXMPGeneration, "XMP Sidecar Generation", 2, func(msg *message.Message) error {
		var job XMPGenerationJob
		err := json.Unmarshal(msg.Payload, &job)
		if err != nil {
			return fmt.Errorf("%s: %w", JobTypeXMPGeneration, err)
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-started", map[string]any{
				"jobId":    msg.UUID,
				"type":     JobTypeXMPGeneration,
				"imageId":  job.Image.Uid,
				"filename": job.Image.ImageMetadata.FileName,
			})
		}

		// mark running
		startedAt := time.Now().UTC()
		_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusRunning, nil, nil, &startedAt, nil)

		onProgress := jobs.NewProgressCallback(
			wsBroker,
			msg.UUID,
			JobTypeXMPGeneration,
			job.Image.Uid,
			job.Image.ImageMetadata.FileName,
		)

		err = generateXMPSidecar(job.Image, logger, onProgress)

		if err != nil {
			if wsBroker != nil {
				wsBroker.Broadcast("job-failed", map[string]any{
					"jobId":   msg.UUID,
					"type":    JobTypeXMPGeneration,
					"imageId": job.Image.Uid,
					"error":   err.Error(),
				})
			}
			_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusFailed, utils.StringPtr("worker_error"), utils.StringPtr(jobs.Truncate(err.Error(), 1024)), nil, nil)
			return err
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-completed", map[string]any{
				"jobId":   msg.UUID,
				"type":    JobTypeXMPGeneration,
				"imageId": job.Image.Uid,
			})
		}

		completedAt := time.Now().UTC()
		_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusSuccess, nil, nil, nil, &completedAt)

		return nil
	},
	)
}


func generateXMPSidecar(img entities.Image, logger *slog.Logger, onProgress func(step string, progress int)) error {
	originalPath := images.GetImagePath(img.Uid, img.ImageMetadata.FileName)

	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return fmt.Errorf("original image file not found: %s", originalPath)
	}

	if onProgress != nil {
		onProgress("Validating input", 5)
	}

	xmpPath := strings.TrimSuffix(originalPath, filepath.Ext(originalPath)) + ".xmp"
	doc := xmp.NewDocument()
	xmpBase := &xmpbase.XmpBase{
		CreatorTool: "Imagine Image Management System",
	}

	dcModel := &dc.DublinCore{}
	exifModel := &exif.ExifInfo{}
	tiffModel := &tiff.TiffInfo{}

	if onProgress != nil {
		onProgress("Building XMP models", 20)
	}

	if img.Exif != nil {
		if img.Exif.DateTimeOriginal != nil {
			if t, err := xmp.ParseDate(*img.Exif.DateTimeOriginal); err == nil {
				xmpBase.CreateDate = t
				exifModel.DateTimeOriginalXMP = t
			}
		}

		if img.Exif.ModifyDate != nil {
			if t, err := xmp.ParseDate(*img.Exif.ModifyDate); err == nil {
				xmpBase.ModifyDate = t
			}
		}

		if img.Exif.Make != nil {
			tiffModel.Make = *img.Exif.Make
		}

		if img.Exif.Model != nil {
			tiffModel.Model = *img.Exif.Model
		}

		if img.Exif.Software != nil {
			tiffModel.Software = *img.Exif.Software
		}

		if img.Exif.Orientation != nil {
			orientation, err := imageops.ConvertOrientation(*img.Exif.Orientation)
			if err != nil {
				logger.Error("Failed to convert orientation", "error", err)
			} else {
				tiffModel.Orientation = orientation
			}
		}
	}

	if onProgress != nil {
		onProgress("Populating metadata", 50)
	}

	if img.ImageMetadata != nil {
		if img.ImageMetadata.Label != nil {
			xmpBase.Label = *img.ImageMetadata.Label
		}

		if img.ImageMetadata.Keywords != nil && len(*img.ImageMetadata.Keywords) > 0 {
			dcModel.Subject = *img.ImageMetadata.Keywords
		}
	}

	if img.Description != nil {
		dcModel.Description = xmp.NewAltString(*img.Description)
	}

	doc.AddModel(xmpBase)
	doc.AddModel(dcModel)
	doc.AddModel(exifModel)
	doc.AddModel(tiffModel)

	if onProgress != nil {
		onProgress("Marshalling XMP", 80)
	}

	xmpData, err := xmp.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XMP data: %w", err)
	}

	if onProgress != nil {
		onProgress("Writing XMP file", 90)
	}

	if err := os.WriteFile(xmpPath, xmpData, 0644); err != nil {
		return fmt.Errorf("failed to write XMP file: %w", err)
	}

	doc.Close()

	logger.Info("generated XMP sidecar",
		slog.String("image_uid", img.Uid),
		slog.String("xmp_path", xmpPath),
	)

	if onProgress != nil {
		onProgress("Complete", 100)
	}
	return nil
}
