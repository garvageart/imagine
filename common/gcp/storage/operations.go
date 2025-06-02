package gcp

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type ImagineStorageOperations interface {
	CreateObject(name string, data []byte) (int, error)
	ReadObject(name string) ([]byte, error)
	DeleteObject(name string) error
	MoveObject(srcName string, destName string) error
}

type ImagineStorage struct {
	Bucket  *storage.BucketHandle
	Context context.Context
	ImagineStorageOperations
}

func (s *ImagineStorage) CreateObject(name string, data []byte) (int, error) {
	object := s.Bucket.Object(name)
	writer := object.NewWriter(s.Context)

	numOfBytesWritten, err := writer.Write(data) // Use writer.Write for []byte
	if err != nil {
		return numOfBytesWritten, fmt.Errorf("error writing data. Wrote %d bytes %w", numOfBytesWritten, err)
	}

	err = writer.Close(
	)
	if err != nil {
		return numOfBytesWritten, fmt.Errorf("error closing object writer %w", err)
	}
	return numOfBytesWritten, nil
}

func (s *ImagineStorage) ReadObject(name string) ([]byte, error) {
	object := s.Bucket.Object(name)
	emptyBytes := make([]byte, 0)

	reader, err := object.NewReader(s.Context)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return emptyBytes, fmt.Errorf("object does not exist %w", err)
		}
		return emptyBytes, fmt.Errorf("error reading object %w", err)
	}
	defer reader.Close()

	gcsRes, err := io.ReadAll(reader)
	if err != nil {
		return emptyBytes, fmt.Errorf("error reading object data %w", err)
	}
	return gcsRes, nil
}

func (s *ImagineStorage) DeleteObject(name string) error {
	object := s.Bucket.Object(name)
	err := object.Delete(s.Context)
	if err != nil {
		return fmt.Errorf("error deleting object %w", err)
	}
	return nil
}
