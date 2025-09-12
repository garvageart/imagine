package imageops

import (
	"bytes"
	"image"
	"os"

	libos "imagine/internal/os"

	libvips "github.com/cshum/vipsgen/vips"
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
