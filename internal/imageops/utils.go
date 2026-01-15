package imageops

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	"imagine/internal/config"
	"imagine/internal/entities"
	libvips "imagine/internal/imageops/vips"
	libos "imagine/internal/os"

	"github.com/galdor/go-thumbhash"
)

type LibvipsImage struct {
	Height float64
	Width  float64
	Ref    *libvips.Image
}

var (
	DefaultWriteFileOptions = &libos.OsPerm{
		DirPerm:  os.ModePerm,
		FilePerm: os.ModePerm,
	}
)

func GetColourSpaceString(image *libvips.Image) string {
	if image.Interpretation() == libvips.InterpretationError {
		return "Error"
	} else if image.Interpretation() == libvips.InterpretationMultiband {
		return "Multiband"
	} else if image.Interpretation() == libvips.InterpretationBW {
		return "BW"
	} else if image.Interpretation() == libvips.InterpretationHistogram {
		return "Histogram"
	} else if image.Interpretation() == libvips.InterpretationXyz {
		return "XYZ"
	} else if image.Interpretation() == libvips.InterpretationLab {
		return "LAB"
	} else if image.Interpretation() == libvips.InterpretationCmyk {
		return "CMYK"
	} else if image.Interpretation() == libvips.InterpretationLabq {
		return "LabQ"
	} else if image.Interpretation() == libvips.InterpretationRgb {
		return "RGB"
	} else if image.Interpretation() == libvips.InterpretationCmc {
		return "CMC"
	} else if image.Interpretation() == libvips.InterpretationLch {
		return "LCH"
	} else if image.Interpretation() == libvips.InterpretationLabs {
		return "LABS"
	} else if image.Interpretation() == libvips.InterpretationSrgb {
		return "sRGB"
	} else if image.Interpretation() == libvips.InterpretationYxy {
		return "YXY"
	} else if image.Interpretation() == libvips.InterpretationFourier {
		return "Fourier"
	} else if image.Interpretation() == libvips.InterpretationRgb16 {
		return "RGB16"
	} else if image.Interpretation() == libvips.InterpretationGrey16 {
		return "Grey16"
	} else if image.Interpretation() == libvips.InterpretationMatrix {
		return "Matrix"
	} else if image.Interpretation() == libvips.InterpretationScrgb {
		return "scRGB"
	} else if image.Interpretation() == libvips.InterpretationHsv {
		return "HSV"
	}

	return "Unknown"
}

func ScaleProportionally(lv *libvips.Image, width int, height int) (*libvips.Image, error) {
	image := lv

	originalWidth := image.Width()
	originalHeight := image.Height()
	scale := 1.0

	outputHeightScale := float64(height) / float64(originalHeight)
	outputWidthScale := float64(width) / float64(originalWidth)

	// This is probably unnecessary but whatever
	if originalWidth > originalHeight {
		scale = float64(outputHeightScale)
	} else {
		scale = float64(outputWidthScale)
	}

	err := image.Resize(scale, libvips.DefaultResizeOptions())
	if err != nil {
		return nil, err
	}

	return image, nil
}

func ReadToImage(imgBytes []byte) (image.Image, string, error) {
	img, str, err := image.Decode(bytes.NewReader(imgBytes))
	return img, str, err
}

func GenerateThumbhash(img image.Image) (hash []byte, err error) {
	hashBytes := thumbhash.EncodeImage(img)
	return hashBytes, nil
}

func ConvertEXIFDateTime(exifDateTime string) *time.Time {
	if exifDateTime == "" {
		return nil
	}

	// Trim whitespace and common surrounding tokens
	s := strings.TrimSpace(exifDateTime)

	// Common EXIF/date formats we want to accept. Try them in order.
	layouts := []string{
		// Standard EXIF layout
		"2006:01:02 15:04:05",
		// Some tools use dash separators
		"2006-01-02 15:04:05",
		// RFC3339 / ISO formats
		time.RFC3339,
		"2006:01:02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
		// With/without timezone designator
		"2006:01:02 15:04:05-07:00",
		"2006-01-02 15:04:05-07:00",
	}

	// Some EXIF values might include extra annotations like "2006:01:02 15:04:05 (some text)".
	// Strip trailing parenthetical content if present.
	if idx := strings.Index(s, " ("); idx > 0 {
		s = strings.TrimSpace(s[:idx])
	}

	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return &t
		}
	}

	// As a last resort, try to parse only the date portion if time.Parse fails
	// e.g., "2006:01:02" or "2006-01-02"
	dateOnlyLayouts := []string{"2006:01:02", "2006-01-02"}
	for _, l := range dateOnlyLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return &t
		}
	}

	return nil
}

// WarmupAllOps forces libvips to instantiate a set of commonly-used
// operations so the corresponding modules/plugins are loaded eagerly.
// Call this after SetLogging (so libvips logs are routed) and after
// Startup (Startup is called internally by HasOperation). This reduces
// first-request latency by avoiding lazy loading when the operation is
// first used.
func WarmupAllOps(cfg config.LibvipsConfig) {
	// Configure VIPS for optimal performance
	vipsConfig := &libvips.Config{
		ConcurrencyLevel: cfg.Concurrency,
		MaxCacheFiles:    cfg.CacheMaxFiles,
		MaxCacheMem:      cfg.CacheMaxMemoryMB * 1024 * 1024,
		MaxCacheSize:     cfg.CacheMaxOperations,
		ReportLeaks:      false,
		CacheTrace:       false,
		VectorEnabled:    cfg.VectorEnabled,
	}

	if vipsConfig.MaxCacheFiles == 0 {
		vipsConfig.MaxCacheFiles = 100
	}
	if vipsConfig.MaxCacheMem == 0 {
		vipsConfig.MaxCacheMem = 50 * 1024 * 1024 // 50MB
	}
	if vipsConfig.MaxCacheSize == 0 {
		vipsConfig.MaxCacheSize = 500
	}

	// ensure vips is started with optimized config
	libvips.Startup(vipsConfig)

	ops := []string{
		// common loaders/savers and ops that trigger plugin loading
		"jpegload", "jpegload_buffer", "jpegsave", "jpegsave_buffer",
		"pngload", "pngsave", "pngsave_buffer",
		"webpload", "webpsave",
		"gifload", "gifload_buffer", "gifsave", "gifsave_buffer",
		"heifload", "heifload_buffer", "heifsave", "heifsave_buffer",
		"jxlload", "jxlload_buffer", "jxlsave", "jxlsave_buffer",
		"jp2kload", "jp2kload_buffer", "jp2ksave", "jp2ksave_buffer",
		"magickload", "magickload_buffer", "magicksave",
		"pdfload", "pdfload_buffer",
		"openslideload", "openslideload_source",
		"tiffload", "tiffsave",
		"fitsload",
		// a few generic/core ops to exercise vips core
		"resize", "thumbnail", "dzsave", "dzsave_buffer",
	}

	for _, name := range ops {
		// HasOperation will call Startup if needed and will invoke
		// vips_operation_new which causes the plugin that provides
		// the operation to be loaded.
		_ = libvips.HasOperation(name)
	}
}

// ParseExifDate tries several common EXIF/ISO formats and returns the parsed time if successful.
// It accepts optional offset strings (e.g. "+02:00") which are tried in order by appending
// them to the primary date string before parsing.
func ParseExifDate(s *string, offsets ...*string) (time.Time, bool) {
	if s == nil {
		return time.Time{}, false
	}
	str := strings.TrimSpace(*s)
	if str == "" {
		return time.Time{}, false
	}

	// Common EXIF: "2006:01:02 15:04:05"
	layouts := []string{
		"2006:01:02 15:04:05",
		"2006:01:02 15:04:05-07:00",
		"2006:01:02 15:04",
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	// Try with offsets first
	for _, offsetPtr := range offsets {
		if offsetPtr == nil {
			continue
		}
		offset := strings.TrimSpace(*offsetPtr)
		if offset == "" {
			continue
		}
		combined := str + offset
		for _, l := range layouts {
			if t, err := time.Parse(l, combined); err == nil {
				return t, true
			}
		}
	}

	// Fallback to primary string without extra offsets
	for _, l := range layouts {
		if t, err := time.Parse(l, str); err == nil {
			return t, true
		}
	}

	// Try a best-effort replacement: convert first two ':' into '-' for date part
	// e.g. 2020:01:02 12:00:00 -> 2020-01-02 12:00:00
	if len(str) >= 10 {
		datePart := str[:10]
		replaced := strings.Replace(datePart, ":", "-", 2)
		candidate := replaced
		if len(str) > 10 {
			candidate += str[10:]
		}
		for _, l := range layouts {
			if t, err := time.Parse(l, candidate); err == nil {
				return t, true
			}
		}
	}

	return time.Time{}, false
}

// getTakenAt returns the most appropriate taken/creation timestamp for an image,
// Priority: EXIF Original -> EXIF Modify -> metadata file_created_at -> image.created_at
func GetTakenAt(img entities.Image) time.Time {
	// Try EXIF fields first
	if img.Exif != nil {
		if img.Exif.DateTimeOriginal != nil {
			effectiveOffset := GetEffectiveExifOffset(img.Exif)
			if t, ok := ParseExifDate(img.Exif.DateTimeOriginal, effectiveOffset); ok {
				return t
			}
		}

		if t, ok := ParseExifDate(img.Exif.DateTime, img.Exif.OffsetTime); ok {
			return t
		}
		if t, ok := ParseExifDate(img.Exif.ModifyDate, img.Exif.OffsetTime); ok {
			return t
		}
	}

	if img.ImageMetadata != nil {
		if !img.ImageMetadata.FileCreatedAt.IsZero() {
			return img.ImageMetadata.FileCreatedAt
		}
		if !img.ImageMetadata.FileModifiedAt.IsZero() {
			return img.ImageMetadata.FileModifiedAt
		}
	}

	return img.CreatedAt
}

// LessByTakenAtDesc returns true if image `a` should sort before image `b` when
// ordering by taken-at descending (newest first). If taken-at timestamps are
// equal it falls back to Name descending to provide a stable ordering that
// matches the client-side comparator.
func LessByTakenAtDesc(a, b entities.Image) bool {
	ta := GetTakenAt(a)
	tb := GetTakenAt(b)

	if ta.Equal(tb) {
		return a.Name > b.Name
	}

	return ta.After(tb)
}