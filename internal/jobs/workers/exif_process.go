package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"imagine/internal/dto"
	"strings"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/trimmer-io/go-xmp/models/dc"
	xmpbase "github.com/trimmer-io/go-xmp/models/xmp_base"
	"github.com/trimmer-io/go-xmp/xmp"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs"
	"imagine/internal/utils"
	customxmp "imagine/internal/xmp"

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

		if job.Image.ImageMetadata == nil {
			err = fmt.Errorf("job %s failed: image metadata is nil for image %s", JobTypeExifProcess, job.Image.Uid)
			_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusFailed, utils.StringPtr("worker_error"), utils.StringPtr(jobs.Truncate(err.Error(), 1024)), nil, nil)
			return nil // Return nil to avoid retry loop
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-started", map[string]any{
				"uid":       msg.UUID,
				"jobId":     msg.UUID,
				"type":      JobTypeExifProcess,
				"topic":     JobTypeExifProcess,
				"image_uid": job.Image.Uid,
				"imageId":   job.Image.Uid,
				"filename":  job.Image.ImageMetadata.FileName,
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
					"uid":       msg.UUID,
					"jobId":     msg.UUID,
					"type":      JobTypeExifProcess,
					"topic":     JobTypeExifProcess,
					"image_uid": job.Image.Uid,
					"imageId":   job.Image.Uid,
					"error":     err.Error(),
				})
			}
			_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusFailed, utils.StringPtr("worker_error"), utils.StringPtr(jobs.Truncate(err.Error(), 1024)), nil, nil)
			return err
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-completed", map[string]any{
				"uid":       msg.UUID,
				"jobId":     msg.UUID,
				"type":      JobTypeExifProcess,
				"topic":     JobTypeExifProcess,
				"image_uid": job.Image.Uid,
				"imageId":   job.Image.Uid,
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

	// Extract XMP Metadata (ACR, Capture One, Standard)
	if onProgress != nil {
		onProgress("Processing XMP data", 60)
	}

	if doc, err := xmp.Scan(bytes.NewReader(originalData)); err == nil {
		defer doc.Close()

		xmpBase := &xmpbase.XmpBase{}
		dcModel := &dc.DublinCore{}
		crsModel := &customxmp.CameraRawSettings{}
		psModel := &customxmp.PhotoshopInfo{}

		// Register models on the document directly
		// This should trigger SyncFromXMP to populate the structs from the parsed DOM
		doc.AddModel(xmpBase)
		doc.AddModel(dcModel)
		doc.AddModel(crsModel)
		doc.AddModel(psModel)

		// 1. Rating
		// Prioritize existing rating
		if imgEnt.ImageMetadata.Rating == nil || *imgEnt.ImageMetadata.Rating == 0 {
			var rating int
			if crsModel.Rating != nil {
				rating = *crsModel.Rating
			} else if xmpBase.Rating > 0 {
				rating = int(xmpBase.Rating)
			}

			if rating > 0 {
				imgEnt.ImageMetadata.Rating = &rating
			}
		}

		// 2. Label
		// Prioritize existing label
		if imgEnt.ImageMetadata.Label == nil || *imgEnt.ImageMetadata.Label == "" || *imgEnt.ImageMetadata.Label == dto.ImageMetadataLabelNone {
			var label string
			if crsModel.Label != nil && *crsModel.Label != "" {
				label = *crsModel.Label
			} else if xmpBase.Label != "" {
				label = xmpBase.Label
			} else if psModel.Urgency > 0 {
				// Map urgency to color label
				switch psModel.Urgency {
				case 1:
					label = "Red"
				case 2:
					label = "Orange"
				case 3:
					label = "Yellow"
				case 4:
					label = "Green"
				case 5:
					label = "Blue"
				case 6:
					label = "Purple"
				case 7:
					label = "Grey"
				}
			}

			if label != "" {
				// Normalize label to match enum if possible
				normalizedLabel := utils.Capitalize(strings.ToLower(label))
				// Check if it matches valid labels
				switch normalizedLabel {
				case "Red", "Orange", "Yellow", "Green", "Blue", "Purple", "Pink", "Grey", "Gray":
					l := dto.ImageMetadataLabel(normalizedLabel)
					imgEnt.ImageMetadata.Label = &l
				}
			}
		}

		// 3. Keywords / Subjects
		if imgEnt.ImageMetadata.Keywords == nil || len(*imgEnt.ImageMetadata.Keywords) == 0 {
			if len(dcModel.Subject) > 0 {
				// Convert xmp.StringArray to []string
				keywords := []string(dcModel.Subject)
				imgEnt.ImageMetadata.Keywords = &keywords
			}
		}
	}

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

	// Only merge extracted metadata if DB version is missing it, to respect user edits
	if dbImage.ImageMetadata.Rating == nil && imgEnt.ImageMetadata.Rating != nil {
		dbImage.ImageMetadata.Rating = imgEnt.ImageMetadata.Rating
	}
	if (dbImage.ImageMetadata.Label == nil || *dbImage.ImageMetadata.Label == dto.ImageMetadataLabelNone) && imgEnt.ImageMetadata.Label != nil {
		dbImage.ImageMetadata.Label = imgEnt.ImageMetadata.Label
	}
	if (dbImage.ImageMetadata.Keywords == nil || len(*dbImage.ImageMetadata.Keywords) == 0) && imgEnt.ImageMetadata.Keywords != nil {
		dbImage.ImageMetadata.Keywords = imgEnt.ImageMetadata.Keywords
	}

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
