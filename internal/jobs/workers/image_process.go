package workers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"imagine/internal/entities"
	"imagine/internal/imageops"
	"imagine/internal/jobs"

	"cloud.google.com/go/storage"
	"gorm.io/gorm"
)

const (
	JobTypeImageProcess = "image_process"
	TopicImageProcess   = JobTypeImageProcess
)

type ImageProcessJob struct {
	Image entities.Image
}

// NewImageWorker now requires bucket and db injection
func NewImageWorker(bucket *storage.BucketHandle, db *gorm.DB) *jobs.Worker {
	return &jobs.Worker{
		Name:  JobTypeImageProcess,
		Topic: TopicImageProcess,
		Handler: func(msg *message.Message) error {
			var job ImageProcessJob
			err := json.Unmarshal(msg.Payload, &job)
			if err != nil {
				return fmt.Errorf("%s: %w", JobTypeImageProcess, err)
			}

			return ImageProcess(msg.Context(), db, job.Image, bucket)
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
	return jobs.PubSub.Publish(TopicImageProcess, msg)
}

func ImageProcess(ctx context.Context, db *gorm.DB, imgEnt entities.Image, bucket *storage.BucketHandle) error {
	// Validate bucket handle
	if bucket == nil {
		return fmt.Errorf("bucket handle is nil")
	}

	// Download the original image file from the bucket
	imageFile := bucket.Object(fmt.Sprintf("images/%s/%s.%s", imgEnt.UID, imgEnt.FileName, imgEnt.FileType))
	fileDataReader, err := imageFile.NewReader(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get object reader: %w", err)
	}

	fileBytes, err := io.ReadAll(fileDataReader)
	if err != nil {
		return fmt.Errorf("failed to get object reader: %w", err)
	}
	defer fileDataReader.Close()

	// Decode the image bytes into an image object
	img, _, err := imageops.ReadToImage(fileBytes)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// Create a display thumbnail from the image
	thumbData, err := imageops.CreateThumbnailWithSize(img, 200, 0)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail: %w", err)
	}

	// Create a very small thumbnail for thumbhash (e.g., 32x32)
	smallThumbData, err := imageops.CreateThumbnailWithSize(img, 32, 32)
	if err != nil {
		return fmt.Errorf("failed to create small thumbnail for thumbhash: %w", err)
	}

	loggerFields := watermill.LogFields{
		"uid":      imgEnt.UID,
		"filename": imgEnt.FileName,
	}

	jobs.Logger.Info("saving thumbnail to disk", loggerFields)

	// Save the thumbnail back to the bucket
	imageObject := bucket.Object(fmt.Sprintf("images/%s/%s-thumb.%s", imgEnt.UID, imgEnt.FileName, imgEnt.FileType))
	objWriter := imageObject.NewWriter(ctx)
	_, _ = objWriter.Write(thumbData)

	err = objWriter.Close()
	if err != nil {
		return fmt.Errorf("failed to close object writer: %w", err)
	}

	// Decode the thumbnail bytes to an image and generate the thumbhash from it
	jobs.Logger.Info("generating thumbhash", loggerFields)
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

	// Update the database with the generated thumbhash
	err = db.Model(&entities.Image{}).Where("uid = ?", imgEnt.UID).Update("thumbhash", base64.StdEncoding.EncodeToString(thumbhash)).Error
	if err != nil {
		return fmt.Errorf("failed to update db image thumbhash: %w", err)
	}

	return nil
}
