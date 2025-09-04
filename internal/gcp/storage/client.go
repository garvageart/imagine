package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"path"

	"cloud.google.com/go/storage"
	"github.com/fullstorydev/emulators/storage/gcsemu"
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"

	liberrors "imagine/internal/errors"
	libos "imagine/internal/os"
	"imagine/internal/utils"
)

var (
	CredentialsPath  = constructCredentialsPath()
	CredentialsBytes = func() []byte {
		credsBytes, err := ReadCredentials(CredentialsPath)
		if err != nil {
			panic(fmt.Sprint("Error reading credentials file", err))
		}

		return credsBytes
	}()
)

func constructCredentialsPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprint("Error getting working directory", err))
	}

	pathConcat := path.Join(wd, os.Getenv("AUTH_PATH"), os.Getenv("GCP_AUTH_FILE"))
	finalPath := libos.StandardisePaths(pathConcat)

	return finalPath
}

func ReadCredentials(path string) ([]byte, error) {
	credentials, err := os.ReadFile(path)

	if err != nil {
		return make([]byte, 0), err
	}

	return credentials, nil
}

func CredentialsJSON() (CredentialsFileJSON, error) {
	var credentialsJSON CredentialsFileJSON
	credsJsonMarBytes, err := json.RawMessage.MarshalJSON(CredentialsBytes)

	if err != nil {
		return CredentialsFileJSON{}, liberrors.NewErrorf("Failed to marshal raw JSON message: %w", err)
	}

	err = json.Unmarshal(credsJsonMarBytes, &credentialsJSON)
	if err != nil {
		return CredentialsFileJSON{}, liberrors.NewErrorf("Failed to unmarshal raw JSON message: %w", err)
	}

	return credentialsJSON, nil
}

func setupGCSEmuClient(ctx context.Context, addr string) (*storage.Client, error) {
	_ = os.Setenv("GCS_EMULATOR_HOST", addr)

	client, err := gcsemu.NewClient(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to setup GCS emulator client: %w", err)
	}
	defer client.Close()

	return client, nil
}

func setupGCSClient(ctx context.Context) (*storage.Client, error) {
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON(CredentialsBytes))

	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return storageClient, nil
}

func SetupClient(ctx context.Context) (*storage.Client, error) {
	var client *storage.Client
	var err error

	if !utils.IsProduction {
		client, err = setupGCSEmuClient(ctx, "127.0.0.1:9000")
	} else {
		client, err = setupGCSClient(ctx)
	}

	return client, err
}
