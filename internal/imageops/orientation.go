package imageops

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/trimmer-io/go-xmp/models/tiff"
)

// ConvertOrientation converts an EXIF orientation string value to tiff.OrientationType
// EXIF orientation values range from 1-8:
// 1 = Normal (0° rotation) / Top-left
// 2 = Mirrored horizontally / Top-right
// 3 = Rotated 180° / Bottom-right
// 4 = Mirrored vertically / Bottom-left
// 5 = Mirrored horizontally + Rotated 270° CCW / Left-top
// 6 = Rotated 90° CW / Right-top
// 7 = Mirrored horizontally + Rotated 90° CW / Right-bottom
// 8 = Rotated 270° CW / Left-bottom
func ConvertOrientation(orientationStr string) (tiff.OrientationType, error) {
	if orientationStr == "" {
		return 1, nil // Default to normal orientation
	}

	// Try numeric value first
	value, err := strconv.Atoi(orientationStr)
	if err == nil {
		if value < 1 || value > 8 {
			return 0, fmt.Errorf("orientation value out of range (1-8): %d", value)
		}
		return tiff.OrientationType(value), nil
	}

	// Handle text-based orientation values
	orientationStr = strings.ToLower(strings.TrimSpace(orientationStr))
	switch orientationStr {
	case "top-left", "horizontal (normal)":
		return 1, nil
	case "top-right", "mirror horizontal":
		return 2, nil
	case "bottom-right", "rotate 180":
		return 3, nil
	case "bottom-left", "mirror vertical":
		return 4, nil
	case "left-top", "mirror horizontal and rotate 270 cw":
		return 5, nil
	case "right-top", "rotate 90 cw":
		return 6, nil
	case "right-bottom", "mirror horizontal and rotate 90 cw":
		return 7, nil
	case "left-bottom", "rotate 270 cw":
		return 8, nil
	default:
		return 0, fmt.Errorf("unknown orientation value: %s", orientationStr)
	}
}
