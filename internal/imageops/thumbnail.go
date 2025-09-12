package imageops

import (
	"bytes"
	"fmt"
	"image"

	"github.com/kovidgoyal/imaging"
)

// CreateThumbnailWithSize creates a thumbnail of the specified width and height.
func CreateThumbnailWithSize(img image.Image, width, height int) ([]byte, error) {
	imgData := imaging.Resize(img, width, height, imaging.Lanczos)
	imageWriter := bytes.NewBuffer([]byte{})

	err := imaging.Encode(imageWriter, imgData, imaging.PNG)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode image: %w", err)
	}
	return imageWriter.Bytes(), nil
}
