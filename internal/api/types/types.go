package types

import "time"

// Collection represents a collection resource returned by the API
// Mirrors the current wire format used by clients.
type Collection struct {
	CreatedAt   time.Time          `json:"created_at"`
	CreatedBy   *string            `json:"created_by,omitempty"`
	Description *string            `json:"description,omitempty"`
	ImageCount  int                `json:"image_count"`
	Images      *[]CollectionImage `json:"images,omitempty"`
	Name        string             `json:"name"`
	Private     *bool              `json:"private"`
	Uid         string             `json:"uid"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// CollectionCreate is the request body for creating a collection
// Keep minimal fields required by the handler.
type CollectionCreate struct {
	Description *string `json:"description,omitempty"`
	Name        string  `json:"name"`
	Private     *bool   `json:"private"`
}

// CollectionDetailResponse represents a detailed collection view
// with a paginated list of images.
type CollectionDetailResponse struct {
	CreatedBy   *string    `json:"created_by,omitempty"`
	Description *string    `json:"description,omitempty"`
	ImageCount  *int       `json:"image_count,omitempty"`
	Images      ImagesPage `json:"images"`
	Name        string     `json:"name"`
	Private     *bool      `json:"private"`
	Uid         string     `json:"uid"`
}

// CollectionImage represents an image summary within a collection.
type CollectionImage struct {
	AddedAt time.Time `json:"added_at"`
	AddedBy *string   `json:"added_by,omitempty"`
	Uid     string    `json:"uid"`
}

// CollectionListResponse represents a paginated list of collections.
type CollectionListResponse struct {
	Count  *int         `json:"count,omitempty"`
	Href   *string      `json:"href,omitempty"`
	Items  []Collection `json:"items"`
	Limit  int          `json:"limit"`
	Next   *string      `json:"next,omitempty"`
	Offset int          `json:"offset"`
	Prev   *string      `json:"prev,omitempty"`
}

// Image represents an image resource returned by the API.
type Image struct {
	CreatedAt     time.Time      `json:"created_at"`
	Description   *string        `json:"description,omitempty"`
	Exif          *ImageEXIF     `json:"exif,omitempty"`
	Height        int32          `json:"height"`
	ImageMetadata *ImageMetadata `json:"image_metadata,omitempty"`
	ImagePaths    *ImagePaths    `json:"image_paths,omitempty"`
	Name          string         `json:"name"`
	Private       bool           `json:"private"`
	Processed     bool           `json:"processed"`
	Uid           string         `json:"uid"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UploadedBy    *string        `json:"uploaded_by,omitempty"`
	Width         int32          `json:"width"`
}

// ImageEXIF represents selected EXIF metadata.
type ImageEXIF struct {
	Aperture         *string `json:"aperture,omitempty"`
	DateTime         *string `json:"date_time,omitempty"`
	DateTimeOriginal *string `json:"date_time_original,omitempty"`
	ExifVersion      *string `json:"exif_version,omitempty"`
	ExposureTime     *string `json:"exposure_time,omitempty"`
	Flash            *string `json:"flash,omitempty"`
	FocalLength      *string `json:"focal_length,omitempty"`
	Iso              *string `json:"iso,omitempty"`
	Latitude         *string `json:"latitude,omitempty"`
	LensModel        *string `json:"lens_model,omitempty"`
	Longitude        *string `json:"longitude,omitempty"`
	Make             *string `json:"make,omitempty"`
	Model            *string `json:"model,omitempty"`
	ModifyDate       *string `json:"modify_date,omitempty"`
	Orientation      *string `json:"orientation,omitempty"`
	Rating           *string `json:"rating,omitempty"`
	Resolution       *string `json:"resolution,omitempty"`
	Software         *string `json:"software,omitempty"`
	WhiteBalance     *string `json:"white_balance,omitempty"`
}

// ImageMetadata holds file-level metadata.
type ImageMetadata struct {
	Checksum         string    `json:"checksum"`
	ColorSpace       string    `json:"color_space"`
	FileCreatedAt    time.Time `json:"file_created_at"`
	FileModifiedAt   time.Time `json:"file_modified_at"`
	FileName         string    `json:"file_name"`
	FileSize         *int64    `json:"file_size,omitempty"`
	FileType         string    `json:"file_type"`
	Keywords         *[]string `json:"keywords,omitempty"`
	Label            *string   `json:"label,omitempty"`
	OriginalFileName *string   `json:"original_file_name,omitempty"`
	Thumbhash        *string   `json:"thumbhash,omitempty"`
}

// ImagePaths are the canonical locations for different image variants.
type ImagePaths struct {
	OriginalPath  string  `json:"original_path"`
	PreviewPath   string  `json:"preview_path"`
	RawPath       *string `json:"raw_path,omitempty"`
	ThumbnailPath string  `json:"thumbnail_path"`
}

// ImagesPage is a generic pagination wrapper for images.
type ImagesPage struct {
	Count  *int             `json:"count,omitempty"`
	Href   *string          `json:"href,omitempty"`
	Items  []ImagesResponse `json:"items"`
	Limit  int              `json:"limit"`
	Next   *string          `json:"next,omitempty"`
	Offset int              `json:"offset"`
	Prev   *string          `json:"prev,omitempty"`
}

// ImagesResponse is an item in an ImagesPage.
type ImagesResponse struct {
	AddedAt time.Time `json:"added_at"`
	AddedBy *string   `json:"added_by,omitempty"`
	Image   Image     `json:"image"`
}
