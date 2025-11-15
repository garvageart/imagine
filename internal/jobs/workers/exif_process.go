package workers

import (
    "context"
    "encoding/json"
    "fmt"
    "imagine/internal/dto"

    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill/message"

    "imagine/internal/entities"
    libhttp "imagine/internal/http"
    "imagine/internal/imageops"
    "imagine/internal/images"
    "imagine/internal/jobs"
    libvips "imagine/internal/imageops/vips"

    "gorm.io/gorm"
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
    return &jobs.Worker{
        Name:  JobTypeExifProcess,
        Topic: TopicExifProcess,
        Handler: func(msg *message.Message) error {
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
                return err
            }

            if wsBroker != nil {
                wsBroker.Broadcast("job-completed", map[string]any{
                    "jobId":   msg.UUID,
                    "type":    JobTypeExifProcess,
                    "imageId": job.Image.Uid,
                })
            }

            return nil
        },
    }
}

// EnqueueExifProcessJob publishes a new exif processing job to the queue.
func EnqueueExifProcessJob(job *ExifProcessJob) error {
    payload, err := json.Marshal(job)
    if err != nil {
        return fmt.Errorf("%s: %w", JobTypeExifProcess, err)
    }

    msg := message.NewMessage(watermill.NewUUID(), payload)
    return jobs.Publish(TopicExifProcess, msg)
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

	imgEnt.ImageMetadata.ColorSpace = imageops.GetColourSpaceString(libvipsImg)

    takenAt := imageops.GetTakenAt(imgEnt)

    if onProgress != nil {
        onProgress("Updating database", 90)
    }

    if err := db.Model(&entities.Image{}).
        Where("uid = ?", imgEnt.Uid).
        Update("exif", imgEnt.Exif).
        Update("taken_at", takenAt).
		Update("image_metadata->>'color_space'", imgEnt.ImageMetadata.ColorSpace).
        Error; err != nil {
        return fmt.Errorf("failed to update db image exif: %w", err)
    }

    return nil
}
