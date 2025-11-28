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
	"imagine/internal/utils"

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
	return jobs.NewWorker(JobTypeImageProcess, TopicImageProcess, "Image Processing", 5, func(msg *message.Message) error {
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

		// Mark job as running in DB
		startedAt := time.Now().UTC()
		_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusRunning, nil, nil, &startedAt, nil)

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
			// persist concise error
			_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusFailed, utils.StringPtr("worker_error"), utils.StringPtr(jobs.Truncate(err.Error(), 1024)), nil, nil)
			return err
		}

		if wsBroker != nil {
			wsBroker.Broadcast("job-completed", map[string]any{
				"jobId":   msg.UUID,
				"type":    JobTypeImageProcess,
				"imageId": job.Image.Uid,
			})
		}

		// mark completed
		completedAt := time.Now().UTC()
		_ = jobs.UpdateWorkerJobStatus(db, msg.UUID, jobs.WorkerJobStatusSuccess, nil, nil, nil, &completedAt)

		return nil
	},
	)
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
		"uid":  imgEnt.Uid,
		"name": imgEnt.Name,
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

	ext := imgEnt.ImageMetadata.FileType

	var transformParams *imageops.TransformParams
	var terr error

	// Generate thumbnail transform (permanent paths)
	tstart := time.Now()
	if imgEnt.ImagePaths.Thumbnail != "" {
		jobs.Logger.Debug("GenerateTransformFromPath: generating transform", loggerFields.Add(watermill.LogFields{
			"path": imgEnt.ImagePaths.Thumbnail,
		}))

		transformParams, terr = imageops.ParseTransformParams(imgEnt.ImagePaths.Thumbnail)
		if terr != nil {
			return terr
		}

		result, terr := imageops.GenerateTransform(transformParams, imgEnt, originalData)
		if terr != nil {
			if terr.Error() == images.CacheErrTransformExists {
				jobs.Logger.Debug("GenerateTransformFromPath: transform already exists", loggerFields.Add(watermill.LogFields{
					"path": imgEnt.ImagePaths.Thumbnail,
				}))
			} else {
				return terr
			}
		} else {
			// Write cache
			if result.Ext != "" {
				ext = result.Ext
			}

			if terr := images.WriteCachedTransform(imgEnt.Uid, *result.TransformHash, ext, result.ImageData); terr != nil {
				return fmt.Errorf("failed to write cached transform: %w", terr)
			}

			jobs.Logger.Debug("GenerateTransformFromPath: finished generating transform", watermill.LogFields{
				"uid":         imgEnt.Uid,
				"path":        imgEnt.ImagePaths.Thumbnail,
				"duration_ms": time.Since(tstart).Milliseconds(),
			})

			jobs.Logger.Debug("finished generating thumbnail transform", loggerFields)
		}
	}

	// Generate preview transform
	tstart = time.Now()
	if imgEnt.ImagePaths.Preview != "" {
		jobs.Logger.Debug("GenerateTransformFromPath: generating transform", loggerFields.Add(watermill.LogFields{
			"path": imgEnt.ImagePaths.Preview,
		}))

		transformParams, terr = imageops.ParseTransformParams(imgEnt.ImagePaths.Preview)
		if terr != nil {
			return terr
		}

		result, terr := imageops.GenerateTransform(transformParams, imgEnt, originalData)
		if terr != nil {
			if terr.Error() == images.CacheErrTransformExists {
				jobs.Logger.Debug("GenerateTransformFromPath: transform already exists", loggerFields.Add(watermill.LogFields{
					"path": imgEnt.ImagePaths.Preview,
				}))
			} else {
				return terr
			}
		} else {
			// Write cache
			if result.Ext != "" {
				ext = result.Ext
			}

			if terr := images.WriteCachedTransform(imgEnt.Uid, *result.TransformHash, ext, result.ImageData); terr != nil {
				return fmt.Errorf("failed to write cached transform: %w", terr)
			}

			jobs.Logger.Debug("GenerateTransformFromPath: finished generating transform", watermill.LogFields{
				"uid":         imgEnt.Uid,
				"path":        imgEnt.ImagePaths.Preview,
				"duration_ms": time.Since(tstart).Milliseconds(),
			})

			jobs.Logger.Debug("finished generating preview transform", loggerFields)
		}
	}

	// just for debugging purposes in case some thumbhashes take too long
	thumbhashTimeStart := time.Now()
	smallThumbImg, _, err := imageops.ReadToImage(smallThumbData)
	if err != nil {
		return fmt.Errorf("failed to decode thumbnail for thumbhash: %w", err)
	}

	thumbhash, err := imageops.GenerateThumbhash(smallThumbImg)
	if err != nil {
		return fmt.Errorf("failed to generate thumbhash: %w", err)
	}

	if onProgress != nil {
		onProgress("Generating transforms", 80)
	}

	jobs.Logger.Debug("finished generating thumbhash", loggerFields.Add(watermill.LogFields{
		"duration": time.Since(thumbhashTimeStart).Milliseconds(),
	}))

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
