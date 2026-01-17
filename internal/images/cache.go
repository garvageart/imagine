package images

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"imagine/internal/config"
	"imagine/internal/dto"
	"imagine/internal/entities"
	"imagine/internal/transform"
)

const (
	TempTransformPrefix = "tmp-transform-"
)

const (
	CacheErrTransformNotFound = "cache: transform not found"
	CacheErrTransformExists   = "cache: transform already exists"
	CacheErrTransformFailed   = "cache: transform failed"
)

// cacheFileName returns the filename for a given key and extension (sha1 hex)
func cacheFileName(key string, ext string) string {
	h := sha1.Sum([]byte(key))
	return fmt.Sprintf("%x.%s", h, ext)
}

// CacheDirForUID returns the transforms dir for a given UID, creating it if necessary
func CacheDirForUID(uid string) (string, error) {
	if err := CreateImageDir(uid); err != nil {
		return "", err
	}

	dir := filepath.Join(GetImageDir(uid), "transforms")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}

	return dir, nil
}

// CacheFilePath returns the expected cache file path for the given uid/key/ext
func CacheFilePath(uid string, key string, ext string) (string, error) {
	dir, err := CacheDirForUID(uid)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, cacheFileName(key, ext)), nil
}

// FindCachedTransform returns the path to the cached transform if it exists.
// If not present, exists==false.
func FindCachedTransform(uid string, key string, ext string) (path string, exists bool, err error) {
	path, err = CacheFilePath(uid, key, ext)
	if err != nil {
		return "", false, err
	}

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}

		return "", false, err
	}

	return path, true, nil
}

// ReadCachedTransform reads the cached transform bytes for the given uid/key/ext.
func ReadCachedTransform(uid string, key string, ext string) (data []byte, err error) {
	path, ok, err := FindCachedTransform(uid, key, ext)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New(CacheErrTransformNotFound)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// WriteCachedTransform writes bytes to a temp file then renames into place atomically.
func WriteCachedTransform(uid string, key string, ext string, data []byte) error {
	dir, err := CacheDirForUID(uid)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, TempTransformPrefix)
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()

	// write
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	tmpFile.Close()

	finalPath := filepath.Join(dir, cacheFileName(key, ext))

	// atomic rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

// PurgeTransformsForUID removes the transforms directory for a UID
func PurgeTransformsForUID(uid string) error {
	dir := filepath.Join(GetImageDir(uid), "transforms")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(dir)
}

// GetCacheStatus calculates and returns the current status of the image transform cache.
func GetCacheStatus() (dto.CacheStatusResponse, error) {
	var totalSize int64
	var totalItems int64

	entries, err := os.ReadDir(Directory)
	if err != nil {
		return dto.CacheStatusResponse{}, fmt.Errorf("failed to read images directory: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		transformsPath := filepath.Join(Directory, e.Name(), "transforms")
		info, err := os.Stat(transformsPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return dto.CacheStatusResponse{}, fmt.Errorf("failed to stat transforms directory %s: %w", transformsPath, err)
		}
		if !info.IsDir() {
			continue
		}

		tfiles, err := os.ReadDir(transformsPath)
		if err != nil {
			return dto.CacheStatusResponse{}, fmt.Errorf("failed to read transforms directory %s: %w", transformsPath, err)
		}

		for _, tf := range tfiles {
			if tf.IsDir() {
				continue
			}

			base := tf.Name()
			if len(base) >= len(TempTransformPrefix) && base[:len(TempTransformPrefix)] == TempTransformPrefix {
				continue
			}

			finfo, err := tf.Info()
			if err != nil {
				return dto.CacheStatusResponse{}, fmt.Errorf("failed to get file info for %s: %w", filepath.Join(transformsPath, base), err)
			}

			totalSize += finfo.Size()
			totalItems++
		}
	}

	// For now, hits and misses are not tracked.
	// If tracking is implemented, update these values.
	var hitRatio float64
	if totalItems > 0 {
		hitRatio = float64(0) / float64(totalItems) // Placeholder
	}

	return dto.CacheStatusResponse{
		Size:     int(totalSize),
		Items:    int(totalItems),
		Hits:     0,
		Misses:   0,
		HitRatio: float32(hitRatio),
	}, nil
}

// ClearCache removes all cached transform files.
func ClearCache(logger *slog.Logger) error {
	entries, err := os.ReadDir(Directory)
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		uid := e.Name()
		if err := PurgeTransformsForUID(uid); err != nil {
			return fmt.Errorf("failed to purge transforms for UID %s: %w", uid, err)
		}

		logger.Debug("cleared transform cache for UID", slog.String("uid", uid))
	}
	return nil
}

// PermanentHashGetter is a function that returns a set of hashes for permanent transforms that should be preserved.
type PermanentHashGetter func(db *gorm.DB) (map[string]bool, error)

// GetPermanentTransformHashes builds a map of hashes for all permanent transforms of existing images.
func GetPermanentTransformHashes(db *gorm.DB) (map[string]bool, error) {
	permanentHashes := make(map[string]bool)
	var images []entities.Image

	// Process images in batches to avoid loading everything into memory
	err := db.Model(&entities.Image{}).Where("deleted_at IS NULL").FindInBatches(&images, 1000, func(tx *gorm.DB, batch int) error {
		for _, img := range images {
			if img.ImageMetadata == nil {
				continue
			}
			// Recalculate etag for each permanent transform and add its hash to the set
			for _, params := range GetAllPermanentTransforms() {
				etag := *transform.CreateTransformEtag(img, &params)
				// The filename is the SHA1 hash of the etag
				fname := cacheFileName(etag, params.Format)
				hash := strings.TrimSuffix(fname, filepath.Ext(fname))
				permanentHashes[hash] = true
			}
		}
		return nil
	}).Error

	return permanentHashes, err
}

// PerformTransformCacheCleanup executes the cache cleanup logic.
func PerformTransformCacheCleanup(rootDir string, logger *slog.Logger, db *gorm.DB, cfg config.CacheConfig, hashGetter PermanentHashGetter) {
	var maxSizeBytes int64 = 10 * 1000 * 1000 * 1000 // 10 GB
	var maxAgeDays int = 30
	var cleanupIntervalMinutes int = 60 * 24 // daily
	var shouldPreservePermanent bool = true

	if cfg.GCEnabled {
		maxSizeBytes = cfg.MaxSizeBytes
		maxAgeDays = cfg.MaxAgeDays
		cleanupIntervalMinutes = cfg.CleanupIntervalMinutes
		shouldPreservePermanent = !cfg.ClearPermanentTransforms
	}

	if cleanupIntervalMinutes <= 0 {
		cleanupIntervalMinutes = 1440 // 24 hours
	}

	if logger != nil {
		logger.Debug("transform cache gc: starting", slog.Int64("max_size_bytes", maxSizeBytes), slog.Int("max_age_days", maxAgeDays), slog.Bool("preserve_permanent", shouldPreservePermanent))
	}

	permanentHashes := make(map[string]bool)
	if shouldPreservePermanent && db != nil && hashGetter != nil {
		logger.Debug("transform cache gc: building list of permanent transforms to preserve")
		var err error
		permanentHashes, err = hashGetter(db)
		if err != nil {
			logger.Error("transform cache gc: failed to build permanent transform list", slog.Any("error", err))
		} else {
			logger.Debug("transform cache gc: finished building permanent transform list", slog.Int("count", len(permanentHashes)))
		}
	}

	type fileInfo struct {
		path string
		size int64
		mod  time.Time
	}

	var files []fileInfo
	var total int64

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		if logger != nil {
			logger.Warn("cache gc: failed to read images directory", slog.Any("error", err))
		}
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		transformsPath := filepath.Join(rootDir, e.Name(), "transforms")
		info, err := os.Stat(transformsPath)
		if err != nil || !info.IsDir() {
			continue
		}

		tfiles, err := os.ReadDir(transformsPath)
		if err != nil {
			continue
		}

		for _, tf := range tfiles {
			if tf.IsDir() {
				continue
			}

			base := tf.Name()
			if strings.HasPrefix(base, TempTransformPrefix) {
				continue
			}

			finfo, err := tf.Info()
			if err != nil {
				continue
			}

			p := filepath.Join(transformsPath, base)
			files = append(files, fileInfo{path: p, size: finfo.Size(), mod: finfo.ModTime()})
			total += finfo.Size()
		}
	}

	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	var remaining []fileInfo
	for _, f := range files {
		if shouldPreservePermanent {
			hash := strings.TrimSuffix(filepath.Base(f.path), filepath.Ext(f.path))
			if permanentHashes[hash] {
				if logger != nil {
					logger.Debug("transform cache gc: preserving permanent file", slog.String("path", f.path))
				}
				// Do not add to 'remaining' so it is not subject to size-based eviction

				continue // Skip to next file
			}
		}

		if f.mod.Before(cutoff) {
			if err := os.Remove(f.path); err == nil {
				if logger != nil {
					logger.Debug("transform cache gc: removed old file", slog.String("path", f.path), slog.Time("mod", f.mod))
				}

				total -= f.size
				continue
			} else if logger != nil {
				logger.Warn("transform cache gc: failed to remove old file", slog.String("path", f.path), slog.Any("error", err))
			}
		}

		remaining = append(remaining, f)
	}

	if total > maxSizeBytes {
		sort.Slice(remaining, func(i, j int) bool { return remaining[i].mod.Before(remaining[j].mod) })
		for _, f := range remaining {
			if total <= maxSizeBytes {
				break
			}
			if err := os.Remove(f.path); err == nil {
				total -= f.size
				if logger != nil {
					logger.Debug("transform cache gc: evicted file", slog.String("path", f.path), slog.Int64("size", f.size))
				}
			} else if logger != nil {
				logger.Warn("transform cache gc: failed to evict file", slog.String("path", f.path), slog.Any("error", err))
			}
		}
	}

	if logger != nil {
		logger.Debug("transform cache gc: finished", slog.Int64("remaining_total_bytes", total))
	}
}

// StartTransformCacheGC starts a background goroutine that periodically
// enforces the transform cache eviction policy based on config values.
// The goroutine will stop when ctx is canceled.
func StartTransformCacheGC(ctx context.Context, logger *slog.Logger, db *gorm.DB) {
	go func() {
		cfg := config.AppConfig
		cleanupIntervalMinutes := cfg.Cache.CleanupIntervalMinutes
		if cleanupIntervalMinutes <= 0 {
			cleanupIntervalMinutes = 1440 // 24 hours
			if logger != nil {
				logger.Warn("transform cache gc: invalid cleanup interval, using default", slog.Int("interval_minutes", cleanupIntervalMinutes))
			}
		}

		interval := time.Duration(cleanupIntervalMinutes) * time.Minute
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		doCleanup := func() {
			PerformTransformCacheCleanup(Directory, logger, db, config.AppConfig.Cache, GetPermanentTransformHashes)
		}

		// do startup run
		doCleanup()

		for {
			select {
			case <-ctx.Done():
				if logger != nil {
					logger.Debug("transform cache gc: stopping")
				}
				return
			case <-ticker.C:
				doCleanup()
			}
		}
	}()
}
