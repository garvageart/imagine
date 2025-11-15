package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	"imagine/internal/images"
	"imagine/internal/jobs"

	"gorm.io/gorm"
)

const (
	JobTypeImageProcess = "image_process"
	TopicImageProcess   = JobTypeImageProcess
)

type ImageProcessJob struct {
	Image entities.Image
}

// NewImageWorker creates a worker that processes images and sends WebSocket updates
func NewImageWorker(db *gorm.DB, wsBroker *libhttp.WSBroker) *jobs.Worker {
	return &jobs.Worker{
		Name:  JobTypeImageProcess,
		Topic: TopicImageProcess,
		Handler: func(msg *message.Message) error {
			var job ImageProcessJob
			err := json.Unmarshal(msg.Payload, &job)
			if err != nil {
				return fmt.Errorf("%s: %w", JobTypeImageProcess, err)
			}

			if wsBroker != nil {
				wsBroker.Broadcast("job-started", map[string]any{
					"jobId":    msg.UUID,
					"type":     JobTypeImageProcess,
					"imageId":  job.Image.Uid,
					"filename": job.Image.ImageMetadata.FileName,
				})
			}

			// Create reusable progress reporter and process
			onProgress := jobs.NewProgressCallback(
				wsBroker,
				msg.UUID,
				JobTypeImageProcess,
				job.Image.Uid,
				job.Image.ImageMetadata.FileName,
			)

			err = ImageProcess(msg.Context(), db, job.Image, onProgress)

			if err != nil {
				if wsBroker != nil {
					wsBroker.Broadcast("job-failed", map[string]any{
						"jobId":   msg.UUID,
						"type":    JobTypeImageProcess,
						"imageId": job.Image.Uid,
						"error":   err.Error(),
					})
				}
				return err
			}

			if wsBroker != nil {
				wsBroker.Broadcast("job-completed", map[string]any{
					"jobId":   msg.UUID,
					"type":    JobTypeImageProcess,
					"imageId": job.Image.Uid,
				})
			}

			return nil
		},
	}
}

// EnqueueImageProcessJob publishes a new image processing job to the queue.
func EnqueueImageProcessJob(job *ImageProcessJob) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("%s: %w", JobTypeImageProcess, err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	return jobs.Publish(TopicImageProcess, msg)
}

func ImageProcess(ctx context.Context, db *gorm.DB, imgEnt entities.Image, onProgress func(step string, progress int)) error {
	originalData, err := images.ReadImage(imgEnt.Uid, imgEnt.ImageMetadata.FileName)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	if onProgress != nil {
		onProgress("Creating display thumbnail", 25)
	}

	// Create a display thumbnail from the image
	thumbData, err := imageops.CreateThumbnailWithSize(originalData, 200, 0)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail: %w", err)
	}

	if onProgress != nil {
		onProgress("Creating thumbnail for thumbhash", 40)
	}

	// Create a very small thumbnail for thumbhash (e.g., 32x32)
	smallThumbData, err := imageops.CreateThumbnailWithSize(originalData, 32, 32)
	if err != nil {
		return fmt.Errorf("failed to create small thumbnail for thumbhash: %w", err)
	}

	loggerFields := watermill.LogFields{
		"uid":      imgEnt.Uid,
		"filename": imgEnt.ImageMetadata.FileName,
	}

	jobs.Logger.Info("saving thumbnail to disk", loggerFields)

	if onProgress != nil {
		onProgress("Saving thumbnail to disk", 55)
	}

	// Save the thumbnail to disk
	err = images.SaveImage(thumbData, imgEnt.Uid, fmt.Sprintf("%s-thumbnail", imgEnt.Uid)+".jpeg")
	if err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	// Decode the thumbnail bytes to an image and generate the thumbhash from it
	jobs.Logger.Info("generating thumbhash", loggerFields)

	if onProgress != nil {
		onProgress("Generating thumbhash", 70)
	}

	thumbhashTimeStart := time.Now()
	// Generate a thumbhash from the small thumbnail
	smallThumbImg, _, err := imageops.ReadToImage(smallThumbData)
	if err != nil {
		return fmt.Errorf("failed to decode thumbnail for thumbhash: %w", err)
	}
	thumbhash, err := imageops.GenerateThumbhash(smallThumbImg)
	if err != nil {
		return fmt.Errorf("failed to generate thumbhash: %w", err)
	}

	jobs.Logger.Debug("finished generating thumbhash", loggerFields.Add(watermill.LogFields{
		"duration": time.Since(thumbhashTimeStart).Milliseconds(),
	}))

	// Update the database with the generated thumbhash (stored inside image_metadata JSON)
	encoded := images.EncodeThumbhashToString(thumbhash)
	imgEnt.ImageMetadata.Thumbhash = &encoded

	if onProgress != nil {
		onProgress("Updating database", 90)
	}

	if err := db.Model(&entities.Image{}).
		Where("uid = ?", imgEnt.Uid).
		Update("image_metadata", imgEnt.ImageMetadata).
		Error; err != nil {
		return fmt.Errorf("failed to update db image thumbhash: %w", err)
	}

	return nil
}
