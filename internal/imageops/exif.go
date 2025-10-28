package imageops

import (
	"fmt"
	"strings"

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
