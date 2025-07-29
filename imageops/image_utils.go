package imageops

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"

	libos "imagine/common/os"

	libvips "github.com/davidbyttow/govips/v2/vips"
	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/galdor/go-thumbhash"
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

func ScaleProportionally(lv *libvips.ImageRef, width int, height int) (*libvips.ImageRef, error) {
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

	err := image.Resize(scale, libvips.KernelAuto)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func ReadExif(bytes []byte) (data map[string]any, err error) {
	exifData, err := exif.SearchAndExtractExif(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read exif data: %w", err)
	}

	exifMap, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to create exif map: %w", err)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(exifMap, ti, exifData)
	if err != nil {
		return nil, fmt.Errorf("failed to collect exif data: %w", err)
	}

	var mapData map[string]any
	cb := func(ifd *exif.Ifd, ite *exif.IfdTagEntry) error {
		mapData[ite.String()] = ite.Value
		return nil
	}

	err = index.RootIfd.EnumerateTagsRecursively(cb)

	if err != nil {
		return nil, fmt.Errorf("failed to enumerate exif data: %w", err)
	}

	return mapData, nil
}

func ReadToImage(reader io.Reader) (image.Image, string, error) {
	img, str, err := image.Decode(reader)
	return img, str, err
}

func GenerateThumbhash(imgData []byte) (hash []byte, err error) {
	ioRead := io.NewSectionReader(bytes.NewReader(imgData), 0, int64(len(imgData)))

	imgDecoded, _, err := ReadToImage(ioRead)

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	hashBytes := thumbhash.EncodeImage(imgDecoded)

	return hashBytes, nil
}
