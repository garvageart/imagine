package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"imagine/internal/dto"

	"github.com/ThreeDotsLabs/watermill/message"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/utils"

	"gorm.io/gorm"
	"time"
)

const (
	JobTypeExifProcess = "exif_process"
	TopicExifProcess   = JobTypeExifProcess
)

type ExifProcessJob struct {
	Image entities.Image
}

// NewExifWorker creates a worker that extracts EXIF and updates the DB
func NewExifWorker(db *gorm.DB, wsBroker *libhttp.WSBroker) *jobs.Worker {
	return jobs.NewWorker(JobTypeExifProcess, TopicExifProcess, "EXIF Processing", 2, func(msg *message.Message) error {
		var job ExifProcessJob
		err := json.Unmarshal(msg.Payload, &job)
		if err != nil {
			return fmt.Errorf("%s: %w", JobTypeExifProcess, err)
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-started", map[string]any{
				"jobId":    msg.UUID,
				"type":     JobTypeExifProcess,
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
			JobTypeExifProcess,
			job.Image.Uid,
			job.Image.ImageMetadata.FileName,
		)

		err = ExifProcess(msg.Context(), db, job.Image, onProgress)

		if err != nil {
			if wsBroker != nil {
				wsBroker.Broadcast("job-failed", map[string]any{
					"jobId":   msg.UUID,
					"type":    JobTypeExifProcess,
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
				"type":    JobTypeExifProcess,
				"imageId": job.Image.Uid,
			})
		}

		completedAt := time.Now().UTC()
		_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusSuccess, nil, nil, nil, &completedAt)

		return nil
	},
	)
}

// ExifProcess extracts EXIF and updates the DB (exif + taken_at + optional metadata)
func ExifProcess(ctx context.Context, db *gorm.DB, imgEnt entities.Image, onProgress func(step string, progress int)) error {
	originalData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
	if err != nil {
		return fmt.Errorf("failed to read image for exif: %w", err)
	}

	if onProgress != nil {
		onProgress("Processing EXIF data", 30)
	}

	libvipsImg, err := libvips.NewImageFromBuffer(originalData, libvips.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to create vips image from buffer: %w", err)
	}
	defer libvipsImg.Close()

	exifData, fileCreatedAt, fileModifiedAt := imageops.BuildImageEXIF(libvipsImg.Exif())
	imgEnt.Exif = &exifData

	if imgEnt.ImageMetadata == nil {
		imgEnt.ImageMetadata = &dto.ImageMetadata{}
	}

	if !fileCreatedAt.IsZero() {
		imgEnt.ImageMetadata.FileCreatedAt = fileCreatedAt
	}

	if !fileModifiedAt.IsZero() {
		imgEnt.ImageMetadata.FileModifiedAt = fileModifiedAt
	}

	hasIcc := libvipsImg.HasICCProfile()
	imgEnt.ImageMetadata.ColorSpace = imageops.GetColourSpaceString(libvipsImg)
	imgEnt.ImageMetadata.HasIccProfile = &hasIcc
	takenAt := imageops.GetTakenAt(imgEnt)

	if onProgress != nil {
		onProgress("Updating database", 90)
	}

	// Fetch latest image from DB to ensure we don't overwrite concurrent metadata updates (e.g. dimensions)
	var dbImage entities.Image
	if err := db.Where("uid = ?", imgEnt.Uid).First(&dbImage).Error; err != nil {
		return fmt.Errorf("failed to fetch image for update: %w", err)
	}

	// Ensure metadata struct exists
	if dbImage.ImageMetadata == nil {
		dbImage.ImageMetadata = &dto.ImageMetadata{}
	}

	// Merge extracted data
	dbImage.ImageMetadata.ColorSpace = imgEnt.ImageMetadata.ColorSpace
	dbImage.ImageMetadata.HasIccProfile = imgEnt.ImageMetadata.HasIccProfile
	dbImage.ImageMetadata.FileCreatedAt = imgEnt.ImageMetadata.FileCreatedAt
	dbImage.ImageMetadata.FileModifiedAt = imgEnt.ImageMetadata.FileModifiedAt

	if err := db.Model(&entities.Image{}).
		Where("uid = ?", imgEnt.Uid).
		Update("exif", imgEnt.Exif).
		Update("taken_at", takenAt).
		Update("image_metadata", dbImage.ImageMetadata).
		Error; err != nil {
		return fmt.Errorf("failed to update db image exif: %w", err)
	}

	return nil
}
