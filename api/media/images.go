package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	libvips "github.com/davidbyttow/govips/v2/vips"
	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	"imagine/common/entities"
	gcp "imagine/common/gcp/storage"
	libhttp "imagine/common/http"
	"imagine/common/uid"
	"imagine/imageops"
	"imagine/utils"
)

type ImageUpload struct {
	Name    string `json:"name,omitempty"`
	Private bool   `json:"private"`
}

type ImageUploadURL struct {
	URLs []string `json:"urls"`
}

type ImageUploadError struct {
	ID      string `json:"name"`
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
}

func ImagesRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	ctx := context.Background()

	// TODO: create a writer interface that any writer can
	// implement besides Google Cloud Storage. This is just my
	// personal writer and choice right now
	storageClient, err := gcp.SetupClient(ctx)
	imageBucket := storageClient.Bucket("imagine-test-dev")

	if err != nil {
		panic("Failed to setup GCP Storage client" + err.Error())
	}

	// Hack to stop govips from logging any messages. Requires editing and exporting
	// libvips.DisableLogging(from the original package source code.)
	// May need a check and change every once in a while
	if utils.IsProduction || os.Getenv("LIBVIPS_DISABLE_LOGGING") == "true" {
		libvips.DisableLogging()
	}

	router.Post("/urls", func(res http.ResponseWriter, req *http.Request) {
		ctx, gcsContextCancel := context.WithCancel(ctx)
		defer gcsContextCancel()
		var imageUpload ImageUploadURL

		err := render.DecodeJSON(req.Body, &imageUpload)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		logger.Info("Received image upload request", slog.Group("data",
			slog.Any("count", len(imageUpload.URLs)),
			slog.String("request_id", libhttp.GetRequestID(req)),
		))

		failedImages := make([]ImageUploadError, 0)
		dbSlice := make([]entities.Image, 0)

		// Any failed files should be added to a failed files array/slice
		// and returned to the user after all files have been uploaded
		// at this point in the process, if something has gone so
		// catastrophically wrong to crash the server or whatever, then
		// return a 5XX error but this shouldn't happen
		for i, imageUrl := range imageUpload.URLs {
			// Download image first:
			logger.Info("Downloading image", slog.String("url", imageUrl), slog.Int("index", i))
			fileBytes, err := libhttp.DownloadFile(imageUrl)
			if err != nil {
				logger.Error("Failed to download file", slog.Any("error", err), slog.String("url", imageUrl))
				continue
			}

			urlParsed, err := url.Parse(imageUrl)

			if err != nil {
				logger.Error("Failed to parse url", slog.Any("error", err), slog.String("url", imageUrl))
				continue
			}

			fileName, _ := strings.CutPrefix(urlParsed.Path, "/")
			logger.Info("Generating ID", slog.String("file", fileName))

			id, err := uid.Generate()
			fileLogger := logger.With(
				slog.String("name", fileName),
				slog.String("id", id),
			)

			if err != nil {
				failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to generate ID", Retryable: true})
				fileLogger.Warn("Failed to generate ID", slog.String("file", fileName), slog.Any("error", err))
				continue
			}

			libvipsImg, libvipsErr := libvips.NewImageFromBuffer(fileBytes)

			if libvipsErr != nil {
				failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to create libvips image", Retryable: false})
				fileLogger.Warn("Failed to create libvips image", slog.String("file", fileName), slog.Any("error", err))
				continue
			}

			defer libvipsImg.Close() // free memory right here

			fileLogger.Info("reading exif data")
			exifData := libvipsImg.GetExif()

			if len(exifData) == 0 {
				logger.Warn("No exif data found. Blank fields", slog.String("file", fileName))
			} else {
				logger.Debug("exif data", slog.Any("data", exifData), slog.Int("length", len(exifData)))
			}

			exif := entities.ImageEXIF{
				Model:            exifData["Model"],
				Make:             exifData["Make"],
				ExifVersion:      exifData["ExifVersion"],
				DateTime:         exifData["DateTime"],
				DateTimeOriginal: exifData["DateTimeOriginal"],
				ModifyDate:       exifData["ModifyDate"],
				ISO:              exifData["ISO"],
				FocalLength:      exifData["FocalLength"],
				ExposureTime:     exifData["ExposureTime"],
				Aperture:         exifData["Aperture"],
				Flash:            exifData["Flash"],
				WhiteBalance:     exifData["WhiteBalance"],
				LensModel:        exifData["LensModel"],
				Rating:           exifData["Rating"],
				Orientation:      exifData["Orientation"],
				Resolution:       exifData["Resolution"],
				Software:         exifData["Software"],
				Longitude:        exifData["Longitude"],
				Latitude:         exifData["Latitude"],
			}

			createdDate := exif.DateTimeOriginal
			modDate := exif.ModifyDate

			if createdDate == "" || modDate == "" {
				createdDate = time.Now().UTC().String()
				modDate = time.Now().UTC().String()
			}

			fileLogger.Info("generating file name for saving")
			fileNameHash := md5.New()
			fileNameHash.Write([]byte(fileName))
			hexEncodedString := hex.EncodeToString(fileNameHash.Sum(nil))
			fileNameForSaving := id + "-" + hexEncodedString

			metadata := entities.ImageMetadata{
				FileName:         fileName,
				FileSize:         int64(len(fileBytes)),
				OriginalFileName: fileName,
				FileType:         "image/jpeg",
				ColorSpace:       exifData["ColorSpace"],
				FileModifiedAt:   carbon.Parse(modDate).StdTime(),
				FileCreatedAt:    carbon.Parse(createdDate).StdTime(),
				Keywords:         []string{},
				Label:            "",
			}

			fileLogger.Debug("dates", slog.Group("values",
				slog.String("mod from exif", exifData["ModifyDate"]),
				slog.String("create from exif", exifData["DateTimeOriginal"]),
				slog.String("created var", createdDate),
				slog.String("mod var", modDate),
				slog.Any("created", metadata.FileCreatedAt),
				slog.Any("modified", metadata.FileModifiedAt),
			))

			paths := entities.ImagePaths{
				OriginalPath:  fmt.Sprintf("/images/%s/%s", id, fileNameForSaving),
				ThumbnailPath: fmt.Sprintf("/images/%s/%s", id, fileNameForSaving+"-thumbnail"),
				PreviewPath:   fmt.Sprintf("/images/%s/%s", id, fileNameForSaving+"-preview"),
				RawPath:       fmt.Sprintf("/images/%s/%s", id, fileNameForSaving+"-raw"),
			}

			allImageData := entities.Image{
				UID:           id,
				Name:          fileName,
				Private:       false,
				Processed:     false,
				EXIF:          exif,
				ImageMetadata: metadata,
				ImagePaths:    paths,
				Width:         uint32(libvipsImg.Width()),
				Height:        uint32(libvipsImg.Height()),
				Description:   "", //TODO: evaluate if necessary, blank for now
			}

			fileLogger.Info("processing image and generating, thumbnail, preview and original images")

			for key, value := range utils.StructToMap(paths) {
				var imageData []byte
				var metadata *libvips.ImageMetadata

				keyLower := strings.ToLower(key)
				if strings.Contains(keyLower, "raw") {
					continue
				}

				if strings.Contains(keyLower, "thumbnail") {
					libvipsImg, err = imageops.ScaleProportionally(libvipsImg, 512, 512)
					if err != nil {
						failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to scale image", Retryable: false})
						fileLogger.Warn("Failed to scale image", slog.String("file", fileName), slog.Any("error", err))
						continue
					}

					imageData, metadata, libvipsErr = libvipsImg.ExportPng(&libvips.PngExportParams{
						StripMetadata: true,
						Compression:   9,
						Quality:       70,
					})
				} else if strings.Contains(keyLower, "preview") {
					libvipsImg, err = imageops.ScaleProportionally(libvipsImg, 1024, 1024)
					if err != nil {
						failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to scale image", Retryable: false})
						fileLogger.Warn("Failed to scale image", slog.String("file", fileName), slog.Any("error", err))
						continue
					}

					imageData, metadata, libvipsErr = libvipsImg.ExportPng(&libvips.PngExportParams{
						StripMetadata: true,
						Compression:   3,
						Quality:       100,
					})
				} else if strings.Contains(keyLower, "original") {
					imageData = fileBytes
					metadata = libvipsImg.Metadata()
				}

				if libvipsErr != nil {
					failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to process image", Retryable: false})
					fileLogger.Warn("Failed to process image", slog.String("file", fileName), slog.Any("error", err))
					continue
				}

				fileLogger.Info("uploading image to storage", slog.String("path", value.(string)))

				imageObject := imageBucket.Object(fmt.Sprintf("%s.%s", value, metadata.Format.FileExt()))
				objWriter := imageObject.NewWriter(ctx)
				_, _ = objWriter.Write(imageData)

				err = objWriter.Close()
				if err != nil {
					failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to close image file writer", Retryable: true})
					fileLogger.Warn("Failed to close image file writer", slog.String("file", fileName), slog.Any("error", err))
					continue
				}
			}

			fileLogger.Info("generating thumbhash")
			imgThumbhash, err := imageops.GenerateThumbhash(fileBytes)
			if err != nil {
				failedImages = append(failedImages, ImageUploadError{ID: id, Reason: "Failed to generate thumbhash", Retryable: true})
				fileLogger.Warn("Failed to generate thumbhash", slog.String("file", fileName), slog.Any("error", err))
				continue
			}

			thumbhashStrB64 := base64.StdEncoding.EncodeToString(imgThumbhash)
			allImageData.Thumbhash = thumbhashStrB64
			allImageData.Processed = true

			dbSlice = append(dbSlice, allImageData)
		}
		var dbImageIDs = make([]string, 0)
		for _, img := range dbSlice {
			dbImageIDs = append(dbImageIDs, img.UID)
		}

		// remove failed images
		// not final logic
		for _, failedImage := range failedImages {
			if slices.Contains(dbImageIDs, failedImage.ID) {
				dbImageIDs = slices.Delete(dbImageIDs, slices.Index(dbImageIDs, failedImage.ID), 1)
			}
		}

		logger.Info("adding images to database", slog.Int("count", len(dbSlice)))
		dbCreateTx := db.Create(&dbSlice)

		if dbCreateTx.Error != nil {
			libhttp.ServerError(res, req, dbCreateTx.Error, logger, nil,
				"",
				"Failed to create images",
			)
			return
		}

		logger.Info("upload images success", slog.Int("count", len(dbImageIDs)))

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, map[string]any{"images": dbImageIDs})
	})

	return router
}
