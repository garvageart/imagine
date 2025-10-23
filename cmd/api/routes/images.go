package routes

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kovidgoyal/imaging"
	"gorm.io/gorm"

	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/imageops"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/jobs/workers"
	"imagine/internal/uid"
)

type ImageUpload struct {
	Name    string `json:"name,omitempty"`
	Private bool   `json:"private"`
}

type ImageUploadFile struct {
	Data     []byte `json:"data"`
	FileName string `json:"filename"`
	Checksum string `json:"checksum,omitempty"`
}

type ImageUploadError struct {
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Retryable bool   `json:"retryable"`
	Error     string `json:"error"`
}

func createNewImageEntity(logger *slog.Logger, fileName string, libvipsImg *libvips.Image) (*entities.Image, error) {
	logger.Info("Generating ID", slog.String("file", fileName))
	id, err := uid.Generate()

	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	if strings.Trim(fileName, " ") == "" {
		fileName = id
	}

	logger = logger.With(
		slog.String("name", fileName),
		slog.String("id", id),
	)

	logger.Info("reading exif data")
	exifData := libvipsImg.Exif()

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

	logger.Info("generating file name for saving")
	fileNameHash := md5.New()
	fileNameHash.Write([]byte(fileName))
	hexEncodedString := hex.EncodeToString(fileNameHash.Sum(nil))
	fileNameForSaving := id + "-" + hexEncodedString

	metadata := entities.ImageMetadata{
		FileName:         fileName,
		OriginalFileName: fileName,
		FileType:         string(libvipsImg.Format()),
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
		ImageMetadata: &metadata,
		ImagePaths:    &paths,
		Width:         uint32(libvipsImg.Width()),
		Height:        uint32(libvipsImg.Height()),
		Description:   "", //TODO: evaluate if necessary, blank for now
	}

	return &allImageData, nil
}

func ImagesRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

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

		tx := db.Model(&entities.Image{}).Where("uid = ?", uid)
		var imgEnt entities.Image
		result := tx.First(&imgEnt)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				res.WriteHeader(http.StatusNotFound)
				render.JSON(res, req, map[string]string{"error": "image not found"})
				return
			}
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": "failed to fetch image from database"})
			return
		}

		fileBytes := []byte{}

		goimg, _, err := images.ReadFileAsGoImage(imgEnt.UID, imgEnt.FileName, imgEnt.FileType)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}

		newImage := imaging.Resize(goimg, int(width), int(height), imaging.Lanczos)
		fileBytes = newImage.Pix
		libvipsImg, err := libvips.NewImageFromBuffer(fileBytes, libvips.DefaultLoadOptions())

		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}

		defer libvipsImg.Close()

		var imageData []byte

		switch format {
		case "webp":
			imageData, err = libvipsImg.WebpsaveBuffer(&libvips.WebpsaveBufferOptions{
				Q: int(quality),
			})
		case "png":
			imageData, err = libvipsImg.PngsaveBuffer(&libvips.PngsaveBufferOptions{
				Filter:      libvips.PngFilterNone,
				Interlace:   false,
				Palette:     false,
				Compression: int(quality),
			})
		case "jpg":
		case "jpeg":
			imageData, err = libvipsImg.JpegsaveBuffer(&libvips.JpegsaveBufferOptions{
				Q:         int(quality),
				Interlace: true,
			})
		case "avif":
		case "heif":
			imageData, err = libvipsImg.HeifsaveBuffer(&libvips.HeifsaveBufferOptions{
				Q:        int(quality),
				Bitdepth: 8,
				Effort:   5,
				Lossless: false,
			})
		default:
			imageData, err = libvipsImg.RawsaveBuffer(&libvips.RawsaveBufferOptions{
				Keep: libvips.KeepAll,
			})

			format = string(libvipsImg.Format())
		}

		res.Header().Set("Content-Type", "image/"+format)

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
		var fileImageUpload ImageUploadFile

		// Parse the multipart form in the request
		err := req.ParseMultipartForm(10 << 20) // limit your max input length!
		if err != nil {
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		// Get the form fields
		fileImageUpload.FileName = req.FormValue("filename")

		// Get the file from the request
		file, _, err := req.FormFile("data")
		if err != nil {
			render.JSON(res, req, map[string]string{"error": err.Error()})
			return
		}
		defer file.Close()
		fileImageUpload.Data, err = io.ReadAll(file)
		if err != nil {
			render.JSON(res, req, map[string]string{"error": "invalid file data"})
			return
		}

		if fileImageUpload.FileName == "" || len(fileImageUpload.Data) == 0 {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		// fileImageUpload.Checksum = req.FormValue("checksum")

		// if fileImageUpload.Checksum != "" {
		// 	hasher := sha1.New()
		// 	hasher.Write(fileImageUpload.Data)
		// 	calculatedChecksum := hex.EncodeToString(hasher.Sum(nil))
		// 	if fileImageUpload.Checksum != calculatedChecksum {
		// 		res.WriteHeader(http.StatusBadRequest)
		// 		render.JSON(res, req, map[string]string{"error": "checksum mismatch"})
		// 		return
		// 	}
		// }

		libvipsImg, err := libvips.NewImageFromBuffer(fileImageUpload.Data, libvips.DefaultLoadOptions())
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		defer libvipsImg.Close()
		imageEntity, err := createNewImageEntity(logger, fileImageUpload.FileName, libvipsImg)

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image data",
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

		imageEntity.FileSize = int64(len(fileImageUpload.Data))
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

		logger.Info("starting image processing", slog.String("id", imageEntity.UID))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		imageEntity.FileName = strings.Split(imageEntity.FileName, ".")[0] // remove extension for saving
		err = images.SaveImage(fileImageUpload.Data, imageEntity.UID, imageEntity.FileName, imageEntity.FileType)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		err = workers.EnqueueImageProcessJob(workerJob)

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.UID))

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, map[string]string{"id": imageEntity.UID})
	})

	// TODO: either this should be a config option or removed entirely.
	// it's possible that URL uploads could be a security problem.
	// cool idea in theory i guess tho
	router.Post("/url", func(res http.ResponseWriter, req *http.Request) {
		if os.Getenv("ENABLE_URL_UPLOAD") == "false" {
			res.WriteHeader(http.StatusForbidden)
			render.JSON(res, req, map[string]string{"error": "url uploads are disabled"})
			return
		}

		imageUrlBytes, err := io.ReadAll(req.Body)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		imageUrl := string(imageUrlBytes)

		// Any failed files should be added to a failed files array/slice
		// and returned to the user after all files have been uploaded
		// at this point in the process, if something has gone so
		// catastrophically wrong to crash the server or whatever, then
		// return a 5XX error but this shouldn't happen

		// Download image first:
		logger.Info("Downloading image", slog.String("url", imageUrl))
		fileBytes, err := libhttp.DownloadFile(imageUrl)

		if err != nil {
			logger.Error("Failed to download file", slog.Any("error", err), slog.String("url", imageUrl))
			return
		}

		urlParsed, err := url.Parse(imageUrl)

		if err != nil {
			logger.Error("Failed to parse url", slog.Any("error", err), slog.String("url", imageUrl))
			return
		}

		fileName, _ := strings.CutPrefix(urlParsed.Path, "/")
		libvipsImg, err := libvips.NewImageFromBuffer(fileBytes, libvips.DefaultLoadOptions())
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}
		defer libvipsImg.Close()
		imageEntity, err := createNewImageEntity(logger, fileName, libvipsImg)

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image data",
			)
			return
		}

		hasher := sha1.New()
		if _, err := io.Copy(hasher, bytes.NewReader(fileBytes)); err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		imageEntity.FileSize = int64(len(fileBytes))
		imageEntity.Checksum = hex.EncodeToString(hasher.Sum(nil))

		logger.Info("adding image to database", slog.String("id", imageEntity.UID))
		dbCreateTx := db.Create(&imageEntity)

		if dbCreateTx.Error != nil {
			libhttp.ServerError(res, req, dbCreateTx.Error, logger, nil,
				"",
				"Failed to create image",
			)
			return
		}

		logger.Info("starting image processing", slog.String("id", imageEntity.UID))
		workerJob := &workers.ImageProcessJob{
			Image: *imageEntity,
		}

		err = images.SaveImage(fileBytes, imageEntity.UID, imageEntity.FileName, imageEntity.FileType)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		err = workers.EnqueueImageProcessJob(workerJob)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"Failed to process image",
			)
			return
		}

		logger.Info("upload images success", slog.String("id", imageEntity.UID))

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, map[string]string{"id": imageEntity.UID})
	})

	return router
}
