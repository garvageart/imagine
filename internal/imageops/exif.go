package imageops

import (
	"fmt"

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
