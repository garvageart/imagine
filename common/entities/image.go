package entities

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	*gorm.Model
	ImageMetadata
	ImagePaths
	UID         string    `json:"uid"`
	Name        string    `json:"name"`
	UploadedBy  string    `json:"uploaded_by"`
	Description string    `json:"description,omitempty"`
	EXIF        ImageEXIF `json:"exif" gorm:"type:JSONB"`
	Private     bool      `json:"private"`
	Width       uint32    `json:"width"`
	Height      uint32    `json:"height"`
	Processed   bool      `json:"processed"`
}

type ImageMetadata struct {
	FileName         string    `json:"file_name"`
	FileSize         int64     `json:"file_size"`
	OriginalFileName string    `json:"original_file_name,omitempty"`
	FileType         string    `json:"file_type"`
	Keywords         []string  `json:"keywords,omitempty" gorm:"type:text[]"`
	ColorSpace       string    `json:"color_space"`
	FileModifiedAt   time.Time `json:"file_modified_at"`
	FileCreatedAt    time.Time `json:"file_created_at"`
	Thumbhash        string    `json:"thumbhash"`
	Label            string    `json:"label,omitempty"`
}

type ImagePaths struct {
	OriginalPath  string `json:"original_path"`      // Paths are relative, not absolute. An instance can store files anywhere (eg. GCP, AWS, Local Server)
	ThumbnailPath string `json:"thumbnail_path"`     // Paths are relative, not absolute. An instance can store files anywhere (eg. GCP, AWS, Local Server)
	PreviewPath   string `json:"preview_path"`       // Paths are relative, not absolute. An instance can store files anywhere (eg. GCP, AWS, Local Server)
	RawPath       string `json:"raw_path,omitempty"` // Paths are relative, not absolute. An instance can store files anywhere (eg. GCP, AWS, Local Server)
}

type ImageEXIF struct { // NOTE: These are just the most important(?) EXIF fields, the rest will be stored in the original file upload
	ExifVersion      string `json:"exif_version"`
	Make             string `json:"make"`
	Model            string `json:"model"`
	DateTime         string `json:"date_time"`
	DateTimeOriginal string `json:"date_time_original"`
	ISO              string `json:"iso"`
	FocalLength      string `json:"focal_length"`
	ExposureTime     string `json:"exposure_time"`
	Aperture         string `json:"aperture"`
	Flash            string `json:"flash"`
	WhiteBalance     string `json:"white_balance"`
	LensModel        string `json:"lens_model"`
	ModifyDate       string `json:"modify_date"`
	Rating           string `json:"rating"`
	Orientation      string `json:"orientation"`
	Resolution       string `json:"resolution"`
	Software         string `json:"software,omitempty"`
	Longitude        string `json:"longitude"`
	Latitude         string `json:"latitude"`
}

type ImageDupes struct {
	Image
	UID              string `json:"uid"`
	OriginalImageUID string `json:"original_image_uid"`
}
