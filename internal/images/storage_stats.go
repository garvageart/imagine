package images

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// StorageStatsHolder holds the current calculated storage size.
// It uses atomic for thread-safe access without locks for reading.
type StorageStatsHolder struct {
	totalSizeBytes int64
	path           string
}

func NewStorageStatsHolder(path string) *StorageStatsHolder {
	return &StorageStatsHolder{
		path: path,
	}
}

func (s *StorageStatsHolder) GetTotalSize() int64 {
	return atomic.LoadInt64(&s.totalSizeBytes)
}

func (s *StorageStatsHolder) GetPath() string {
	return s.path
}

// StartStorageStatsWorker starts a background goroutine to calculate storage size.
func (s *StorageStatsHolder) StartStorageStatsWorker(ctx context.Context, logger *slog.Logger, interval time.Duration) {
	logger.Info("starting storage stats worker", slog.String("path", s.path), slog.Duration("interval", interval))

	s.calculate(logger)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping storage stats worker")
			return
		case <-ticker.C:
			s.calculate(logger)
		}
	}
}

func (s *StorageStatsHolder) calculate(logger *slog.Logger) {
	start := time.Now()
	var size int64

	err := filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})

	if err != nil {
		logger.Error("failed to calculate storage size", slog.Any("error", err))
		return
	}

	atomic.StoreInt64(&s.totalSizeBytes, size)
	logger.Debug("storage stats updated",
		slog.Int64("size_bytes", size),
		slog.Duration("time_taken", time.Since(start)),
		slog.String("time_taken_seconds", fmt.Sprintf("%.2fs", time.Since(start).Seconds())),
	)
}
