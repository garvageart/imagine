package imageops

import (
	"fmt"
	"strings"
    "time"

    "imagine/internal/dto"
	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
)

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

	mapData := make(map[string]any)
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

// Helpers to normalize EXIF keys/values coming from libvips (exif-ifdX-*)
func CleanExifVal(s string) string {
	// Prefer human-friendly token: if value is like "10/12500 (1/1250 sec., Rational, ...)"
	// pick the first token inside parentheses before the comma. Otherwise take prefix before " ("
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	pIdx := strings.Index(s, " (")
	if pIdx > 0 {
		prefix := strings.TrimSpace(s[:pIdx])
		inner := s[pIdx+2:]
		end := strings.Index(inner, ")")
		if end > 0 {
			inner = inner[:end]
		}
		comma := strings.Index(inner, ",")
		if comma > 0 {
			inner = inner[:comma]
		}
		inner = strings.TrimSpace(inner)
		// If prefix looks like a fraction and inner looks like a nicer fraction/number, prefer inner
		if strings.Contains(prefix, "/") && inner != "" {
			return inner
		}
		// If inner is a wordy label (e.g., Left-bottom), prefer inner for Orientation
		if strings.HasPrefix(inner, "Top") || strings.HasPrefix(inner, "Bottom") || strings.Contains(inner, "left") || strings.Contains(inner, "right") || strings.Contains(inner, "sec") {
			return inner
		}
		return prefix
	}
	return s
}

func FindExif(exifData map[string]string, keys ...string) *string {
	if len(exifData) == 0 {
		return nil
	}
	var raw string
	search := func(k string) (string, bool) {
		if v, ok := exifData[k]; ok {
			return v, true
		}
		// Try common libvips prefixes and IFD groups
		if v, ok := exifData["exif-"+k]; ok { // rarely present but cheap to check
			return v, true
		}
		for _, ifd := range []string{"ifd0", "ifd1", "ifd2", "ifd3", "ifd4"} {
			key := "exif-" + ifd + "-" + k
			if v, ok := exifData[key]; ok {
				return v, true
			}
		}
		return "", false
	}
	for _, k := range keys {
		if v, ok := search(k); ok {
			raw = v
			break
		}
	}
	if raw == "" {
		return nil
	}
	val := CleanExifVal(raw)
	if val == "" {
		return nil
	}
	return &val
}


// BuildImageEXIF normalizes libvips EXIF map into a dto.ImageEXIF and returns
// parsed created/modified times (with sensible fallbacks).
func BuildImageEXIF(exifData map[string]string) (dto.ImageEXIF, time.Time, time.Time) {
    var out dto.ImageEXIF
    if len(exifData) == 0 {
        return out, time.Time{}, time.Time{}
    }

    out = dto.ImageEXIF{
        Model:            FindExif(exifData, "Model"),
        Make:             FindExif(exifData, "Make"),
        ExifVersion:      FindExif(exifData, "ExifVersion"),
        DateTime:         FindExif(exifData, "DateTime", "ModifyDate"),
        DateTimeOriginal: FindExif(exifData, "DateTimeOriginal"),
        ModifyDate:       FindExif(exifData, "ModifyDate", "DateTime"),
        Iso:              FindExif(exifData, "ISO", "ISOSpeedRatings"),
        FocalLength:      FindExif(exifData, "FocalLength"),
        ExposureTime:     FindExif(exifData, "ExposureTime"),
        Aperture:         FindExif(exifData, "ApertureValue", "FNumber", "Aperture"),
        Flash:            FindExif(exifData, "Flash"),
        WhiteBalance:     FindExif(exifData, "WhiteBalance"),
        LensModel:        FindExif(exifData, "LensModel"),
        Rating:           FindExif(exifData, "Rating"),
        Orientation:      FindExif(exifData, "Orientation"),
        Software:         FindExif(exifData, "Software"),
        Longitude:        FindExif(exifData, "GPSLongitude", "Longitude"),
        Latitude:         FindExif(exifData, "GPSLatitude", "Latitude"),
    }

    // Derive resolution from X/Y if present
    xRes := FindExif(exifData, "XResolution")
    yRes := FindExif(exifData, "YResolution")
    if xRes != nil && yRes != nil {
        resStr := fmt.Sprintf("%sx%s DPI", *xRes, *yRes)
        out.Resolution = &resStr
    }

    // Parse dates with fallback logic
    var fileCreatedAt time.Time
    var fileModifiedAt time.Time

    if cd := FindExif(exifData, "DateTimeOriginal"); cd != nil {
        if t := ConvertEXIFDateTime(*cd); t != nil {
            fileCreatedAt = *t
        }
    }

    if md := FindExif(exifData, "ModifyDate"); md != nil {
        if t := ConvertEXIFDateTime(*md); t != nil {
            fileModifiedAt = *t
        }
    }

    now := time.Now()
    if fileCreatedAt.IsZero() && fileModifiedAt.IsZero() {
        fileCreatedAt = now
        fileModifiedAt = now
    } else if fileCreatedAt.IsZero() && !fileModifiedAt.IsZero() {
        fileCreatedAt = fileModifiedAt
    } else if !fileCreatedAt.IsZero() && fileModifiedAt.IsZero() {
        fileModifiedAt = fileCreatedAt
    }

    return out, fileCreatedAt, fileModifiedAt
}
