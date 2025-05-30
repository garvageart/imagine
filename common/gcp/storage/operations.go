package gcp

***REMOVED***
***REMOVED***
***REMOVED***
	"io"

	"cloud.google.com/go/storage"
***REMOVED***

type ImagineStorageOperations interface {
	CreateObject(***REMOVED***
	ReadObject(***REMOVED***
	DeleteObject(***REMOVED***
	MoveObject(***REMOVED***
***REMOVED***

type ImagineStorage struct {
	Bucket  *storage.BucketHandle
	Context context.Context
	ImagineStorageOperations
***REMOVED***

func (s *ImagineStorage***REMOVED*** CreateObject(name string, data []byte***REMOVED*** (int, error***REMOVED*** {
	object := s.Bucket.Object(name***REMOVED***

	writer := object.NewWriter(s.Context***REMOVED***
	numOfBytesWritten, err := fmt.Fprint(writer, data***REMOVED***

***REMOVED***
		return numOfBytesWritten, fmt.Errorf("error writing data. Wrote %d bytes %w", numOfBytesWritten, err***REMOVED***
***REMOVED***

	err = writer.Close(***REMOVED***

***REMOVED***
		return numOfBytesWritten, fmt.Errorf("error closing object writer %w", err***REMOVED***
***REMOVED***

	return numOfBytesWritten, nil
***REMOVED***

func (s *ImagineStorage***REMOVED*** ReadObject(name string***REMOVED*** ([]byte, error***REMOVED*** {
	object := s.Bucket.Object(name***REMOVED***

	reader, err := object.NewReader(s.Context***REMOVED***
	emptyBytes := make([]byte, 0***REMOVED***

***REMOVED***
		return emptyBytes, fmt.Errorf("error reading object %w", err***REMOVED***
***REMOVED***

	defer reader.Close(***REMOVED***

	reader, readerErr := object.NewReader(s.Context***REMOVED***
	if readerErr != nil {
		if readerErr == storage.ErrObjectNotExist {
			return emptyBytes, fmt.Errorf("object does not exist %w", readerErr***REMOVED***

	***REMOVED***

		return emptyBytes, fmt.Errorf("error reading object data %w", err***REMOVED***
***REMOVED***

	gcsRes, err := io.ReadAll(reader***REMOVED***
***REMOVED***
		return emptyBytes, fmt.Errorf("error reading object data %w", err***REMOVED***
***REMOVED***

	return gcsRes, nil
***REMOVED***

func (s *ImagineStorage***REMOVED*** DeleteObject(name string***REMOVED*** error {
	object := s.Bucket.Object(name***REMOVED***
	err := object.Delete(s.Context***REMOVED***

***REMOVED***
		return fmt.Errorf("error deleting object %w", err***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***