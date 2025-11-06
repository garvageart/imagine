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
		return "Xyz"
	} else if image.Interpretation() == libvips.InterpretationLab {
		return "Lab"
	} else if image.Interpretation() == libvips.InterpretationCmyk {
		return "Cmyk"
	} else if image.Interpretation() == libvips.InterpretationLabq {
		return "Labq"
	} else if image.Interpretation() == libvips.InterpretationRgb {
		return "Rgb"
	} else if image.Interpretation() == libvips.InterpretationCmc {
		return "Cmc"
	} else if image.Interpretation() == libvips.InterpretationLch {
		return "Lch"
	} else if image.Interpretation() == libvips.InterpretationLabs {
		return "Labs"
	} else if image.Interpretation() == libvips.InterpretationSrgb {
		return "Srgb"
	} else if image.Interpretation() == libvips.InterpretationYxy {
		return "Yxy"
	} else if image.Interpretation() == libvips.InterpretationFourier {
		return "Fourier"
	} else if image.Interpretation() == libvips.InterpretationRgb16 {
		return "Rgb16"
	} else if image.Interpretation() == libvips.InterpretationGrey16 {
		return "Grey16"
	} else if image.Interpretation() == libvips.InterpretationMatrix {
		return "Matrix"
	} else if image.Interpretation() == libvips.InterpretationScrgb {
		return "Scrgb"
	} else if image.Interpretation() == libvips.InterpretationHsv {
		return "Hsv"
	} else if image.Interpretation() == libvips.InterpretationLast {
		return "Last"
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
		// Without timezone designator
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
func WarmupAllOps() {
	// ensure vips is started
	libvips.Startup(nil)

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