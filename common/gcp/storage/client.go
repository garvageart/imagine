package gcp

***REMOVED***
***REMOVED***
	"encoding/json"
***REMOVED***
***REMOVED***
	"path"

***REMOVED***

	"cloud.google.com/go/storage"
	"github.com/fullstorydev/emulators/storage/gcsemu"
	"google.golang.org/api/option"

	liberrors "imagine/common/errors"
	libos "imagine/common/os"
***REMOVED***

var (
	CredentialsPath  = constructCredentialsPath(***REMOVED***
	CredentialsBytes = func(***REMOVED*** []byte {
		credsBytes, err := ReadCredentials(CredentialsPath***REMOVED***
	***REMOVED***
			panic(fmt.Sprint("Error reading credentials file", err***REMOVED******REMOVED***
	***REMOVED***

		return credsBytes
***REMOVED***(***REMOVED***
***REMOVED***

func constructCredentialsPath(***REMOVED*** string {
	wd, err := os.Getwd(***REMOVED***

***REMOVED***
		panic(fmt.Sprint("Error getting working directory", err***REMOVED******REMOVED***
***REMOVED***

	pathConcat := path.Join(wd, os.Getenv("AUTH_PATH"***REMOVED***, os.Getenv("GCP_AUTH_FILE"***REMOVED******REMOVED***
	finalPath := libos.StandardisePaths(pathConcat***REMOVED***

	return finalPath
***REMOVED***

func ReadCredentials(path string***REMOVED*** ([]byte, error***REMOVED*** {
	credentials, err := os.ReadFile(path***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

	return credentials, nil
***REMOVED***

func CredentialsJSON(***REMOVED*** (CredentialsFileJSON, error***REMOVED*** {
	var credentialsJSON CredentialsFileJSON
	credsJsonMarBytes, err := json.RawMessage.MarshalJSON(CredentialsBytes***REMOVED***

***REMOVED***
		return CredentialsFileJSON{***REMOVED***, liberrors.NewErrorf("Failed to marshal raw JSON message: %w", err***REMOVED***
***REMOVED***

	err = json.Unmarshal(credsJsonMarBytes, &credentialsJSON***REMOVED***

***REMOVED***
		return CredentialsFileJSON{***REMOVED***, liberrors.NewErrorf("Failed to unmarshal raw JSON message: %w", err***REMOVED***
***REMOVED***

	return credentialsJSON, nil
***REMOVED***

func setupGCSEmuClient(ctx context.Context, addr string***REMOVED*** (*storage.Client, error***REMOVED*** {
	_ = os.Setenv("GCS_EMULATOR_HOST", addr***REMOVED***

	client, err := gcsemu.NewClient(ctx***REMOVED***
***REMOVED***
	***REMOVED***, fmt.Errorf("failed to setup GCS emulator client: %w", err***REMOVED***
***REMOVED***
	defer client.Close(***REMOVED***

	return client, nil
***REMOVED***
func setupGCSClient(ctx context.Context***REMOVED*** (*storage.Client, error***REMOVED*** {
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(CredentialsBytes***REMOVED******REMOVED***

***REMOVED***
	***REMOVED***, fmt.Errorf("failed to create storage client: %w", err***REMOVED***
***REMOVED***

	return storageClient, nil
***REMOVED***

func SetupClient(ctx context.Context***REMOVED*** (*storage.Client, error***REMOVED*** {
	var client *storage.Client
	var err error

***REMOVED***
		client, err = setupGCSEmuClient(ctx, "127.0.0.1:9000"***REMOVED***
***REMOVED*** else {
		client, err = setupGCSClient(ctx***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
