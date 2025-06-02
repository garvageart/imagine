package imgops

import (
	"os"
	
	libvips "github.com/davidbyttow/govips/v2/vips"

	liberrors "imagine/common/errors"
	"imagine/utils" 
)

func ImageProcess(buffer []byte) error {
	// Hack to stop govips from logging any messages. Requires editing and exporting
	// libvips.DisableLogging( from the original package source code.
	// May need a check and change every once in a while
	if utils.IsProduction || os.Getenv("LIBVIPS_DISABLE_LOGGING") == "true" {
		libvips.DisableLogging()
	}

	libvips.Startup(&libvips.Config{})
	defer libvips.Shutdown()

	image, err := libvips.NewImageFromBuffer(buffer)
	if err != nil {
		return liberrors.NewErrorf(err.Error())
	}
	defer image.Close()

	exportParams := libvips.NewDefaultJPEGExportParams()
	_, _, err = image.Export(exportParams)
	if err != nil {
		return liberrors.NewErrorf("failed to export image: %w", err)
	}
	return nil
}
