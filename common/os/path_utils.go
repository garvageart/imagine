package libos

***REMOVED***
***REMOVED***
***REMOVED***
	"strings"
***REMOVED***

var (
	CurrentWorkingDirectory = func(***REMOVED*** string {
		cwd, err := os.Getwd(***REMOVED***

	***REMOVED***
			panic(fmt.Errorf("error retrieving current working directory: %w", err***REMOVED******REMOVED***
	***REMOVED***

		return StandardisePaths(cwd***REMOVED***
***REMOVED***(***REMOVED***
***REMOVED***

// Microsoft you will pay for your crimes against standards
func StandardisePaths(path string***REMOVED*** string {
	return strings.ReplaceAll(path, "\\", "/"***REMOVED***
***REMOVED***
