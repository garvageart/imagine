package imgops

***REMOVED***
***REMOVED***

	libvips "github.com/davidbyttow/govips/v2/vips"

	liberrors "imagine/common/errors"
***REMOVED***
***REMOVED***

func ImageProcess(buffer []byte***REMOVED*** error {
	// Hack to stop govips from logging any messages. Requires editing and exporting
	// libvips.DisableLogging(***REMOVED*** from the original package source code.
	// May need a check and change every once in a while
	if utils.IsProduction || os.Getenv("LIBVIPS_DISABLE_LOGGING"***REMOVED*** == "true" {
		libvips.DisableLogging(***REMOVED***
***REMOVED***

	libvips.Startup(&libvips.Config{***REMOVED******REMOVED***
	defer libvips.Shutdown(***REMOVED***

	image, err := libvips.NewImageFromBuffer(buffer***REMOVED***

***REMOVED***
		return liberrors.NewErrorf(err.Error(***REMOVED******REMOVED***
***REMOVED***

***REMOVED***
		return liberrors.NewErrorf(err.Error(***REMOVED******REMOVED***
***REMOVED***

	ExportParams := libvips.NewDefaultJPEGExportParams(***REMOVED***
	newImage, metadata, err := image.Export(ExportParams***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
