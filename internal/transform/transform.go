package transform

import (
	"fmt"
	"imagine/internal/entities"
	"imagine/internal/utils"
	"net/url"
	"strconv"
)

// TransformParams defines the parameters for an image transformation.
type TransformParams struct {
	Format   string
	Width    int64
	Height   int64
	Quality  int64
	Rotate   int
	Flip     string
	Kernel   string
}

// ToQueryString serializes the transform parameters into a URL query string.
func (p *TransformParams) ToQueryString() string {
	q := url.Values{}
	if p.Format != "" {
		q.Set("format", p.Format)
	}
	if p.Width > 0 {
		q.Set("w", strconv.FormatInt(p.Width, 10))
	}
	if p.Height > 0 {
		q.Set("h", strconv.FormatInt(p.Height, 10))
	}
	if p.Quality > 0 {
		q.Set("quality", strconv.FormatInt(p.Quality, 10))
	}
	if p.Rotate > 0 {
		q.Set("rotate", strconv.Itoa(p.Rotate))
	}
	if p.Flip != "" {
		q.Set("flip", p.Flip)
	}
	if p.Kernel != "" {
		q.Set("kernel", p.Kernel)
	}
	return q.Encode()
}

// CreateTransformEtag creates a unique ETag for a given image and transform.
func CreateTransformEtag(imgEnt entities.Image, params *TransformParams) *string {
	checksum := "unknown"
	if imgEnt.ImageMetadata != nil {
		checksum = imgEnt.ImageMetadata.Checksum
	}
	return utils.StringPtr(fmt.Sprintf("%s-%dx%d-%s-%d-%d-%s-%s", checksum, params.Width, params.Height, params.Format, params.Quality, params.Rotate, params.Flip, params.Kernel))
}
