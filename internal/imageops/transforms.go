package imageops

import (
	"errors"
	"fmt"
	"imagine/internal/entities"
	libvips "imagine/internal/imageops/vips"
	"imagine/internal/images"
	"imagine/internal/utils"
	"net/url"
	"strconv"
)

type TransformResult struct {
	ImageData     []byte
	TransformHash *string
	Ext           string
}

type TransformParams struct {
	Format   string
	Width    int64
	Height   int64
	Quality  int64
	Rotate   int
	Flip     string
	Kernel   string
}

func CreateTransformEtag(imgEnt entities.Image, params *TransformParams) *string {
	return utils.StringPtr(fmt.Sprintf("%s-%dx%d-%s-%d-%d-%s-%s", imgEnt.ImageMetadata.Checksum, params.Width, params.Height, params.Format, params.Quality, params.Rotate, params.Flip, params.Kernel))
}

func ParseTransformParams(pathStr string) (*TransformParams, error) {
	u, err := url.Parse(pathStr)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	params := &TransformParams{}
	params.Format = q.Get("format")
	params.Flip = q.Get("flip")
	params.Kernel = q.Get("kernel")

	// Check for 'w' (short for width) first, then 'width'
	if widthParam := q.Get("w"); widthParam != "" {
		if w, err := strconv.ParseInt(widthParam, 10, 64); err == nil {
			params.Width = w
		}
	} else if widthParam := q.Get("width"); widthParam != "" {
		if w, err := strconv.ParseInt(widthParam, 10, 64); err == nil {
			params.Width = w
		}
	}

	// Check for 'h' (short for height) first, then 'height'
	if heightParam := q.Get("h"); heightParam != "" {
		if h, err := strconv.ParseInt(heightParam, 10, 64); err == nil {
			params.Height = h
		}
	} else if heightParam := q.Get("height"); heightParam != "" {
		if h, err := strconv.ParseInt(heightParam, 10, 64); err == nil {
			params.Height = h
		}
	}

	if qualityParam := q.Get("quality"); qualityParam != "" {
		if qn, err := strconv.ParseInt(qualityParam, 10, 64); err == nil {
			params.Quality = qn
		}
	}

	if rotateParam := q.Get("rotate"); rotateParam != "" {
		if r, err := strconv.Atoi(rotateParam); err == nil {
			params.Rotate = r
		}
	}

	return params, nil
}

// GenerateTransform generates permanent cached transforms for thumbnail/preview paths if present.
// These are the URLs stored in ImagePaths (e.g. /images/<uid>/file?format=webp&w=400&h=400&quality=85)
func GenerateTransform(params *TransformParams, imgEnt entities.Image, originalData []byte) (result *TransformResult, err error) {
	ext := params.Format
	if ext == "" {
		ext = imgEnt.ImageMetadata.FileType
	}

	// Build transform ETag key same as route
	transformEtag := CreateTransformEtag(imgEnt, params)

	// If cached already exists, skip
	if _, ok, cerr := images.FindCachedTransform(imgEnt.Uid, *transformEtag, ext); cerr == nil && ok {
		return nil, errors.New(images.CacheErrTransformExists)
	}

	// Perform transform using libvips similarly to the route
	libvipsImg, err := libvips.NewImageFromBuffer(originalData, libvips.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create libvips image for transform: %w", err)
	}
	defer libvipsImg.Close()

	err = libvipsImg.Autorot(&libvips.AutorotOptions{}) // non-fatal
	if err != nil {
		return nil, fmt.Errorf("failed to auto-rotate image: %w", err)
	}

	// Ensure consistent color profile (sRGB) for web display
	if err := NormalizeToSRGB(libvipsImg); err != nil {
		return nil, fmt.Errorf("failed to normalize to sRGB: %w", err)
	}

	if params.Rotate > 0 {
		var angle libvips.Angle
		switch params.Rotate {
		case 90:
			angle = libvips.AngleD90
		case 180:
			angle = libvips.AngleD180
		case 270:
			angle = libvips.AngleD270
		default:
			angle = libvips.AngleD0
		}
		if err := libvipsImg.Rotate(float64(angle), &libvips.RotateOptions{}); err != nil {
			return nil, fmt.Errorf("failed to rotate image: %w", err)
		}
	}

	if params.Flip != "" {
		var direction libvips.Direction
		switch params.Flip {
		case "horizontal":
			direction = libvips.DirectionHorizontal
		case "vertical":
			direction = libvips.DirectionVertical
		case "last":
			direction = libvips.DirectionLast
		default:
			direction = libvips.DirectionLast
		}
		_ = libvipsImg.Flip(direction)
	}

	if params.Width > 0 || params.Height > 0 {
		var kernel libvips.Kernel
		if params.Kernel != "" {
			switch params.Kernel {
			case "nearest":
				kernel = libvips.KernelNearest
			case "linear":
				kernel = libvips.KernelLinear
			case "cubic":
				kernel = libvips.KernelCubic
			case "mitchell":
				kernel = libvips.KernelMitchell
			case "lanczos2":
				kernel = libvips.KernelLanczos2
			case "lanczos3":
				kernel = libvips.KernelLanczos3
			case "mks2013":
				kernel = libvips.KernelMks2013
			case "mks2021":
				kernel = libvips.KernelMks2021
			default:
				kernel = libvips.KernelLanczos3 // Default to Lanczos3
			}
		} else {
			kernel = libvips.KernelLanczos3
		}

		// Calculate scale
		scale := 1.0
		imgW := float64(libvipsImg.Width())
		imgH := float64(libvipsImg.Height())

		if params.Width > 0 && params.Height > 0 {
			// Both provided: "contain" behavior (fit within box)
			wScale := float64(params.Width) / imgW
			hScale := float64(params.Height) / imgH
			scale = min(wScale, hScale)
		} else if params.Width > 0 {
			scale = float64(params.Width) / imgW
		} else if params.Height > 0 {
			scale = float64(params.Height) / imgH
		}

		if err := libvipsImg.Resize(scale, &libvips.ResizeOptions{Kernel: kernel}); err != nil {
			return nil, fmt.Errorf("failed to resize image: %w", err)
		}
	}

	// Encode
	var imageData []byte
	switch params.Format {
	case "webp":
		imageData, err = libvipsImg.WebpsaveBuffer(&libvips.WebpsaveBufferOptions{Q: int(params.Quality)})
	case "png":
		imageData, err = libvipsImg.PngsaveBuffer(&libvips.PngsaveBufferOptions{Filter: libvips.PngFilterNone, Interlace: false, Palette: false, Compression: int(params.Quality)})
	case "jpg", "jpeg":
		imageData, err = libvipsImg.JpegsaveBuffer(&libvips.JpegsaveBufferOptions{Q: int(params.Quality), Interlace: true})
	case "avif", "heif":
		imageData, err = libvipsImg.HeifsaveBuffer(&libvips.HeifsaveBufferOptions{Q: int(params.Quality), Bitdepth: 8, Effort: 5, Lossless: false})
	default:
		imageData, err = libvipsImg.RawsaveBuffer(&libvips.RawsaveBufferOptions{Keep: libvips.KeepAll})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode transform: %w", err)
	}

	return &TransformResult{
		ImageData:     imageData,
		TransformHash: transformEtag,
		Ext:           ext,
	}, nil
}
