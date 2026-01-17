package images

import "imagine/internal/transform"

// PermanentTransformName is a type for permanent transform names
type PermanentTransformName string

const (
	// TransformThumbnail is the name for the thumbnail permanent transform.
	TransformThumbnail PermanentTransformName = "thumbnail"
	// TransformPreview is the name for the preview permanent transform.
	TransformPreview PermanentTransformName = "preview"
)

// permanentTransforms defines the parameters for different types of permanent transforms.
var permanentTransforms = map[PermanentTransformName]transform.TransformParams{
	TransformThumbnail: {
		Format:  "webp",
		Width:   400,
		Height:  400,
		Quality: 85,
	},
	TransformPreview: {
		Format:  "webp",
		Width:   1920,
		Height:  1920,
		Quality: 90,
	},
}

// GetPermanentTransformParams returns the transform parameters for a given permanent transform name.
func GetPermanentTransformParams(name PermanentTransformName) (transform.TransformParams, bool) {
	params, ok := permanentTransforms[name]
	return params, ok
}

// GetAllPermanentTransforms returns all permanent transform definitions.
func GetAllPermanentTransforms() map[PermanentTransformName]transform.TransformParams {
	return permanentTransforms
}
