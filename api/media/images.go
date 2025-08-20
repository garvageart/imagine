package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
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

type ImageUploadFile struct {
	Data     []byte `json:"data"`
	FileName string `json:"file_name"`
}

type ImageUploadError struct {
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
	Error     string `json:"error"`
}

func createImageTypes(fileLogger *slog.Logger, img *entities.Image, libvipsImg *libvips.ImageRef, bucket *storage.BucketHandle, context context.Context) *ImageUploadError {
	fileLogger.Info("processing image and generating, thumbnail, preview and original images")
	var libvipsErr error

	for key, value := range utils.StructToMap(img.ImagePaths) {
		var imageData []byte
		var metadata *libvips.ImageMetadata

		keyLower := strings.ToLower(key)
		if strings.Contains(keyLower, "raw") {
			continue
		}

		if strings.Contains(keyLower, "thumbnail") {
			libvipsImgScaled, err := imageops.ScaleProportionally(libvipsImg, 200, 200)
			if err != nil {
				return &ImageUploadError{ID: img.UID, Reason: "Failed to scale image", Retryable: false}
			}

			libvipsImgScaled.ThumbnailWithSize(libvipsImgScaled.Width(), libvipsImgScaled.Height(), libvips.InterestingNone, libvips.SizeDown)
			imageData, metadata, libvipsErr = libvipsImg.ExportWebp(&libvips.WebpExportParams{
				StripMetadata:   true,
				Quality:         30,
				Lossless:        false,
				NearLossless:    false,
				ReductionEffort: 6,
			})

			goimg, err := libvipsImgScaled.ToImage(libvips.NewDefaultExportParams())
			if err != nil {
				fileLogger.Warn("Failed to convert golang image", slog.String("file", img.FileName), slog.Any("error", err))
				return &ImageUploadError{ID: img.UID, Reason: "Failed to convert image to bytes", Retryable: false}
			}

			// NOTE: Use the thumbnail data to generate the thumbhash instead of
			// decoding the image again in a seperate place and using resources uncecessarily
			fileLogger.Info("generating thumbhash")
			imgThumbhash, err := imageops.GenerateThumbhash(fileLogger, goimg)
			if err != nil {
				fileLogger.Warn("Failed to generate thumbhash", slog.String("file", img.FileName), slog.Any("error", err))
				return &ImageUploadError{ID: img.UID, Reason: "Failed to generate thumbhash", Retryable: true}
			}

			thumbhashStrB64 := base64.StdEncoding.EncodeToString(imgThumbhash)
			img.Thumbhash = thumbhashStrB64
		} else if strings.Contains(keyLower, "preview") {
			libvipsImgScaled, err := imageops.ScaleProportionally(libvipsImg, 1080, 1080)
			if err != nil {
				fileLogger.Warn("Failed to scale image", slog.String("file", img.FileName), slog.Any("error", err))
				return &ImageUploadError{ID: img.UID, Reason: "Failed to scale image", Retryable: false}
			}

			libvipsImgScaled.ThumbnailWithSize(libvipsImgScaled.Width(), libvipsImgScaled.Height(), libvips.InterestingNone, libvips.SizeDown)
			imageData, metadata, libvipsErr = libvipsImg.ExportWebp(&libvips.WebpExportParams{
				StripMetadata:   true,
				Quality:         80,
				Lossless:        false,
				NearLossless:    false,
				ReductionEffort: 4,
			})
		} else if strings.Contains(keyLower, "original") {
			imageData, _ = libvipsImg.ToBytes()
			metadata = libvipsImg.Metadata()
		}

		if libvipsErr != nil {
			fileLogger.Warn("Failed to process image", slog.String("file", img.FileName), slog.Any("error", libvipsErr))
			return &ImageUploadError{ID: img.UID, Reason: "Failed to process image", Retryable: false}
		}

		fileLogger.Info("uploading image to storage", slog.String("path", value.(string)))

		imageObject := bucket.Object(fmt.Sprintf("%s%s", value, metadata.Format.FileExt()))
		objWriter := imageObject.NewWriter(context)
		_, _ = objWriter.Write(imageData)

		err := objWriter.Close()
		if err != nil {

			fileLogger.Warn("Failed to close image file writer", slog.String("file", img.FileName), slog.Any("error", err))
			return &ImageUploadError{ID: img.UID, Reason: "Failed to close image file writer", Retryable: true}
		}
	}

	return nil
}

func createNewImageEntity(logger *slog.Logger, fileName string, libvipsImg *libvips.ImageRef) (*entities.Image, *ImageUploadError) {
	logger.Info("Generating ID", slog.String("file", fileName))
	id, err := uid.Generate()

	fileLogger := logger.With(
		slog.String("name", fileName),
		slog.String("id", id),
	)

	if err != nil {
		fileLogger.Warn("Failed to generate ID", slog.String("file", fileName), slog.Any("error", err))
		return nil, &ImageUploadError{ID: id, Reason: "Failed to generate ID", Retryable: true}
	}

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
		OriginalFileName: fileName,
		FileType:         strings.Replace(libvipsImg.Format().FileExt(), ".", "", 1),
		ColorSpace:       imageops.GetColourSpaceString(libvipsImg),
		FileModifiedAt:   carbon.Parse(modDate).StdTime(),
		FileCreatedAt:    carbon.Parse(createdDate).StdTime(),
		Keywords:         []string{},
		Label:            "",
	}

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

	return &allImageData, nil
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

	// TODO: Get param values to serve the original file in different
	// formats, resolutions and sizes
	// TODO TODO: Finalize route name. "/file/" isn't exactly honest in my opinion
	router.Get("/{uid}/file", func(res http.ResponseWriter, req *http.Request) {
		uid := chi.URLParam(req, "uid")
		format := chi.URLParam(req, "format")
		width, wErr := strconv.ParseInt(chi.URLParam(req, "w"), 10, 64)
		height, hErr := strconv.ParseInt(chi.URLParam(req, "h"), 10, 64)
		quality, qErr := strconv.ParseInt(chi.URLParam(req, "quality"), 10, 64)

		// Convert quality to 0-10 instead of 0-100 for the Compression option
		// NOTE: This is not final, depending on if this idea is understandable
		// and accepted by the community/users on release
		if format == "png" {
			quality = int64(math.Round(float64(quality/100) * 10))
		}

		if wErr != nil || hErr != nil || qErr != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request parameters"})
			return
		}

		if width < 0 || height < 0 {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "width/height cannot be negative"})
			return
		}

		imageFile := imageBucket.Object(uid)
		fileDataReader, err := imageFile.NewReader(ctx)

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}

		defer fileDataReader.Close()

		libvipsImg, err := libvips.NewImageFromReader(fileDataReader)
		if height == 0 && width != 0 {
			libvipsImg, err = imageops.ScaleProportionally(libvipsImg, int(width), 0)
		} else if width == 0 && height != 0 {
			libvipsImg, err = imageops.ScaleProportionally(libvipsImg, 0, int(height))
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}

		defer libvipsImg.Close()

		var imageData []byte

		switch format {
		case "webp":
			res.Header().Set("Content-Type", "image/webp")
			imageData, _, err = libvipsImg.ExportWebp(&libvips.WebpExportParams{
				StripMetadata: true,
				Quality:       int(quality),
			})
		case "png":
			res.Header().Set("Content-Type", "image/png")
			imageData, _, err = libvipsImg.ExportPng(&libvips.PngExportParams{
				Filter:      libvips.PngFilterNone,
				Interlace:   false,
				Palette:     false,
				Compression: int(quality),
			})
		case "jpg":
		case "jpeg":
			res.Header().Set("Content-Type", "image/jpeg")
			imageData, _, err = libvipsImg.ExportJpeg(&libvips.JpegExportParams{
				StripMetadata: true,
				Quality:       int(quality),
				Interlace:     true,
			})
		case "avif":
			res.Header().Set("Content-Type", "image/avif")
			imageData, _, err = libvipsImg.ExportAvif(&libvips.AvifExportParams{
				StripMetadata: true,
				Quality:       int(quality),
				Bitdepth:      8,
				Effort:        5,
				Lossless:      false,
			})
		case "heif":
			res.Header().Set("Content-Type", "image/heic")
			imageData, _, err = libvipsImg.ExportHeif(&libvips.HeifExportParams{
				Quality:  int(quality),
				Bitdepth: 8,
				Effort:   5,
				Lossless: false,
			})
		default:
			imageData, err = libvipsImg.ToBytes()
			metadata := libvipsImg.Metadata()

			res.Header().Set("Content-Type", "image/"+strings.ToLower(strings.TrimPrefix(metadata.Format.FileExt(), ".")))
		}

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
		res.WriteHeader(http.StatusOK)
		res.Write(imageData)
	})

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		ctx, gcsContextCancel := context.WithCancel(ctx)
		defer gcsContextCancel()
		var fileImageUpload ImageUploadFile

		err := render.DecodeJSON(req.Body, &fileImageUpload)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		libvipsImg, err := libvips.NewImageFromBuffer(fileImageUpload.Data)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		defer libvipsImg.Close()
		imageEntity, imageErr := createNewImageEntity(logger, "", libvipsImg)

		if imageErr != nil {
			libhttp.ServerError(res, req, errors.New(imageErr.Error), logger, nil,
				"",
				"Failed to process image data",
			)
			return
		}

		imageEntity.FileSize = req.ContentLength
		imageErr = createImageTypes(logger, imageEntity, libvipsImg, imageBucket, ctx)

		if imageErr != nil {
			libhttp.ServerError(res, req, errors.New(imageErr.Error), logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		hasher := md5.New()
		if _, err := io.Copy(hasher, req.Body); err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		imageEntity.Checksum = hex.EncodeToString(hasher.Sum(nil))

		logger.Info("adding images to database", slog.String("id", imageEntity.UID))
		dbCreateTx := db.Create(&imageEntity)

		if dbCreateTx.Error != nil {
			libhttp.ServerError(res, req, dbCreateTx.Error, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.UID))

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, map[string]string{"id": imageEntity.UID})
	})

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

		logger.Info("image upload request", slog.Group("data",
			slog.Any("count", len(imageUpload.URLs)),
			slog.String("request_id", libhttp.GetRequestID(req)),
		))

		failedImages := make([]*ImageUploadError, 0)
		dbSlice := make([]*entities.Image, 0)

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
				failedImages = append(failedImages, &ImageUploadError{ID: id, Reason: "Failed to generate ID", Retryable: true})
				fileLogger.Warn("Failed to generate ID", slog.String("file", fileName), slog.Any("error", err))
				continue
			}

			libvipsImg, libvipsErr := libvips.NewImageFromBuffer(fileBytes)

			if libvipsErr != nil {
				failedImages = append(failedImages, &ImageUploadError{ID: id, Reason: "Failed to create image", Retryable: true})
				fileLogger.Warn("Failed to create image", slog.String("file", fileName), slog.Any("error", libvipsErr))
				continue
			}

			defer libvipsImg.Close()

			allImageData, uploadErr := createNewImageEntity(fileLogger, fileName, libvipsImg)
			if uploadErr != nil {
				failedImages = append(failedImages, uploadErr)
				logger.Warn("Failed to create image entity", slog.String("file", fileName), slog.Any("error", uploadErr))
				continue
			}

			uploadErr = createImageTypes(logger, allImageData, libvipsImg, imageBucket, ctx)
			if uploadErr != nil {
				failedImages = append(failedImages, uploadErr)
				logger.Warn("Failed to create image types", slog.String("file", fileName), slog.Any("error", uploadErr))

				allImageData.Processed = true
				dbSlice = append(dbSlice, allImageData)
			}

			hasher := md5.New()
			if _, err := io.Copy(hasher, bytes.NewReader(fileBytes)); err != nil {
				libhttp.ServerError(res, req, err, logger, nil,
					"",
					"Failed to create image",
				)
				return
			}

			allImageData.Checksum = hex.EncodeToString(hasher.Sum(nil))
		}

		var dbImageIDs = make([]string, 0)
		for _, img := range dbSlice {
			dbImageIDs = append(dbImageIDs, img.UID)
		}

		// remove failed images
		// not final logic
		for _, img := range failedImages {
			if slices.Contains(dbImageIDs, img.ID) {
				logger.Warn("failed to upload image, removing", slog.String("id", img.ID), slog.String("reason", img.Reason))
				dbImageIDs = slices.Delete(dbImageIDs, slices.Index(dbImageIDs, img.ID), 1)
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
