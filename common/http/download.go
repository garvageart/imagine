package http

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var (
	ErrGetFile = errors.New("error getting file")
	ErrReadingBody = errors.New("error reading body")
	ErrCreatingFile = errors.New("error creating file")
	ErrWritingToFile = errors.New("error writing to file")
	
)

func DownloadFile(URL string) ([]byte, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func SaveToFile(path string, fileBytes []byte) (file *os.File, err error) {
	file, err = os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %s", err.Error())
	}

	defer file.Close()

	if _, err := file.Write(fileBytes); err != nil {
		return nil, fmt.Errorf("error writing to file: %s", err.Error())
	}

	return file, nil
}

func DownloadMultipleFiles(urls []string) (allFiles [][]byte, failedUrls []string, err error) {
	var wg sync.WaitGroup
	allFiles = make([][]byte, 0, len(urls))
	failedUrls = make([]string, 0, len(urls))

	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			fileBytes, err := DownloadFile(url)
			if err != nil {
				failedUrls = append(failedUrls, url)
				return
			}

			allFiles = append(allFiles, fileBytes)
		}(url)
	}
	wg.Wait()

	return allFiles, failedUrls, nil
}
