package imageops

import (
	"fmt"
	"os"

	exiftool "github.com/barasher/go-exiftool"
	libvips "github.com/davidbyttow/govips/v2/vips"

	libos "imagine/common/os"
)

type LibvipsImage struct {
	Height float64
	Width  float64
	Ref    *libvips.ImageRef
}

var (
	DefaultWriteFileOptions = &libos.OsPerm{
		DirPerm:  os.ModePerm,
		FilePerm: os.ModePerm,
	}
)

func (lv LibvipsImage) ScaleProportionally() error {
	image := lv.Ref

	originalWidth := image.Width()
	originalHeight := image.Height()
	scale := 1.0

	outputHeightScale := lv.Height / float64(originalHeight)
	outputWidthScale := lv.Width / float64(originalWidth)

	// This is probably unnecessary but whatever
	if originalWidth > originalHeight {
		scale = float64(outputHeightScale)
	} else {
		scale = float64(outputWidthScale)
	}

	return image.Resize(scale, libvips.KernelAuto)
}

func ExtractEXIFData(path string) map[string]string {
	exifData := make(map[string]string)

	exif, err := exiftool.NewExiftool()
	defer exif.Close()

	if err != nil {
		panic("something went wrong with exiftool " + err.Error())
	}

	metadata := exif.ExtractMetadata(path)

	for _, fileMetadata := range metadata {
		if fileMetadata.Err != nil {
			for key, value := range fileMetadata.Fields {
				if str, ok := value.(string); ok {
					exifData[key] = str
				}

				if intValue, ok := value.(int); ok {
					exifData[key] = fmt.Sprint(intValue)
				}
			}
		}
	}

	return exifData
}
