package imageops

import (
	"fmt"

	libvips "imagine/internal/imageops/vips"
)

// CreateThumbnailWithSize creates a thumbnail using libvips from the input image bytes.
//
// Behaviour:
// - Respects EXIF orientation (auto-rotate) via libvips thumbnail operation.
// - If height is 0, only width is constrained and aspect ratio is preserved.
// - If both width and height are provided (>0), libvips will fit within bounds; cropping is not enabled by default.
// - Output encoding is JPEG (binary bytes), suitable for saving with .jpeg extension.
func CreateThumbnailWithSize(imgBytes []byte, width, height int) ([]byte, error) {
	if len(imgBytes) == 0 {
		return nil, fmt.Errorf("no image data provided")
	}

	opts := libvips.DefaultThumbnailBufferOptions()
	opts.NoRotate = false
	opts.Height = height
	
	// Ensure we export to sRGB, handling CMYK/ProPhoto/etc correctly
	// Note: vips_thumbnail handles simple colorspace conversion, but explicit ICC export is better
	opts.OutputProfile = "srgb" // Target sRGB for web
	// InputProfile is a FALLBACK. If the image has an embedded profile, vips uses that.
	// If it has none (e.g. untagged CMYK), it assumes this profile.
	opts.InputProfile = "srgb" 

	thumb, err := libvips.NewThumbnailBuffer(imgBytes, width, opts)
	if err != nil {
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}
	defer thumb.Close()

	jpegOpts := libvips.DefaultJpegsaveBufferOptions()
	data, err := thumb.JpegsaveBuffer(jpegOpts)
	if err != nil {
		return nil, fmt.Errorf("thumbnail encode failed: %w", err)
	}
	return data, nil
}
