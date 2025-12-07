package images

import (
	"context"
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
func (holder *StorageStatsHolder) StartStorageStatsWorker(ctx context.Context, logger *slog.Logger, interval time.Duration) {
	logger.Info("starting storage stats worker", slog.String("path", holder.path), slog.Duration("interval", interval))

	calculate(logger, holder)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping storage stats worker")
			return
		case <-ticker.C:
			calculate(logger, holder)
		}
	}
}

func calculate(logger *slog.Logger, holder *StorageStatsHolder) {
	start := time.Now()
	var size int64

	err := filepath.WalkDir(holder.path, func(path string, d fs.DirEntry, err error) error {
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

	atomic.StoreInt64(&holder.totalSizeBytes, size)
	logger.Debug("storage stats updated", slog.Int64("size_bytes", size), slog.Duration("time_taken", time.Since(start)))
}
