package images

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"imagine/internal/config"
	"imagine/internal/dto"
	"imagine/internal/entities"
	"imagine/internal/transform"
	"imagine/internal/uid"
)

func TestPerformTransformCacheCleanup(t *testing.T) {
	// Setup temp directory
	rootDir := t.TempDir()
	uid := uid.MustGenerate()
	transformsDir := filepath.Join(rootDir, uid, "transforms")
	if err := os.MkdirAll(transformsDir, 0o755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Helper to create dummy files
	createFile := func(name string, age time.Duration, size int64) string {
		path := filepath.Join(transformsDir, name)
		// Create file with specific size
		data := make([]byte, size)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
		// Set mod time
		oldTime := time.Now().Add(-age)
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("failed to chtimes for %s: %v", name, err)
		}
		return path
	}

	// Calculate a hash for our "permanent" file
	permKey := "permanent-transform-key"
	h := sha1.Sum([]byte(permKey))
	permHash := fmt.Sprintf("%x", h)
	permFileName := fmt.Sprintf("%s.webp", permHash)

	// Mock Hash Getter
	mockHashGetter := func(db *gorm.DB) (map[string]bool, error) {
		return map[string]bool{
			permHash: true,
		}, nil
	}

	// Logger for tests
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name                 string
		clearPermanent       bool
		maxSizeBytes         int64
		files                map[string]time.Duration // filename -> age
		fileSizes            map[string]int64         // filename -> size (default 10 bytes)
		expectExists         []string
		expectDeleted        []string
	}{
		{
			name:           "Preserve Permanent (Default)",
			clearPermanent: false,
			maxSizeBytes:   1000,
			files: map[string]time.Duration{
				permFileName:    40 * 24 * time.Hour, // Old permanent
				"old-temp.webp": 40 * 24 * time.Hour, // Old temp
				"new-temp.webp": 1 * time.Hour,       // New temp
			},
			expectExists:  []string{permFileName, "new-temp.webp"},
			expectDeleted: []string{"old-temp.webp"},
		},
		{
			name:           "Clear Permanent (Old)",
			clearPermanent: true,
			maxSizeBytes:   1000,
			files: map[string]time.Duration{
				permFileName:    40 * 24 * time.Hour,
				"old-temp.webp": 40 * 24 * time.Hour,
				"new-temp.webp": 1 * time.Hour,
			},
			expectExists:  []string{"new-temp.webp"},
			expectDeleted: []string{permFileName, "old-temp.webp"},
		},
		{
			name:           "Clear Permanent (New - Should Keep if space)",
			clearPermanent: true,
			maxSizeBytes:   1000,
			files: map[string]time.Duration{
				permFileName:    1 * time.Hour, // New permanent
				"old-temp.webp": 40 * 24 * time.Hour,
			},
			expectExists:  []string{permFileName},
			expectDeleted: []string{"old-temp.webp"},
		},
		{
			name:           "Clear Permanent (Size Pressure)",
			clearPermanent: true,
			maxSizeBytes:   50, // Small limit
			files: map[string]time.Duration{
				permFileName:    2 * time.Hour, // older than new-temp
				"new-temp.webp": 1 * time.Hour, // newest
				"big-file.webp": 1 * time.Hour, // same age
			},
			fileSizes: map[string]int64{
				permFileName:    40,
				"new-temp.webp": 40,
			},
			// Total size = 80. Limit = 50.
			// Both are "new" (not expired by days).
			// Eviction sorts by modification time (oldest first).
			// permFileName is 2h old, new-temp is 1h old.
			// permFileName should be evicted first.
			expectExists:  []string{"new-temp.webp"},
			expectDeleted: []string{permFileName},
		},
		{
			name:           "Preserve Permanent (Size Pressure - Should still keep)",
			clearPermanent: false,
			maxSizeBytes:   50,
			files: map[string]time.Duration{
				permFileName:    2 * time.Hour,
				"new-temp.webp": 1 * time.Hour,
			},
			fileSizes: map[string]int64{
				permFileName:    40,
				"new-temp.webp": 40,
			},
			expectExists:  []string{permFileName},
			expectDeleted: []string{"new-temp.webp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear directory
			files, _ := os.ReadDir(transformsDir)
			for _, f := range files {
				os.Remove(filepath.Join(transformsDir, f.Name()))
			}

			// Create files
			paths := make(map[string]string)
			for name, age := range tt.files {
				size := int64(10) // default size
				if s, ok := tt.fileSizes[name]; ok {
					size = s
				}
				paths[name] = createFile(name, age, size)
			}

			cfg := config.CacheConfig{
				GCEnabled:                true,
				ClearPermanentTransforms: tt.clearPermanent,
				MaxSizeBytes:             tt.maxSizeBytes,
				MaxAgeDays:               30,
				CleanupIntervalMinutes:   60,
			}

			// Run Cleanup
			// Pass a dummy DB since the mockHashGetter doesn't use it but PerformTransformCacheCleanup checks for it
			PerformTransformCacheCleanup(rootDir, logger, &gorm.DB{}, cfg, mockHashGetter)

			// Verify Exists
			for _, name := range tt.expectExists {
				if _, err := os.Stat(paths[name]); os.IsNotExist(err) {
					t.Errorf("expected %s to exist, but it was deleted", name)
				}
			}

			// Verify Deleted
			for _, name := range tt.expectDeleted {
				if _, err := os.Stat(paths[name]); err == nil {
					t.Errorf("expected %s to be deleted, but it exists", name)
				}
			}
		})
	}
}

func TestCacheFileName(t *testing.T) {
	key := "test-key"
	ext := "webp"

	// Expected: sha1("test-key") + ".webp"
	h := sha1.Sum([]byte(key))
	expected := fmt.Sprintf("%x.%s", h, ext)

	got := cacheFileName(key, ext)
	if got != expected {
		t.Errorf("cacheFileName(%q, %q) = %q, want %q", key, ext, got, expected)
	}
}

// Ensure eTags can be reconstructed and are stable
func TestETagReconstruction(t *testing.T) {
	// Dummy Image
	img := entities.Image{
		ImageMetadata: &dto.ImageMetadata{
			Checksum: "checksum123",
		},
	}

	params := transform.TransformParams{
		Width:   100,
		Height:  200,
		Quality: 80,
		Format:  "webp",
	}

	etag1 := transform.CreateTransformEtag(img, &params)
	if etag1 == nil {
		t.Fatal("CreateTransformEtag returned nil")
	}

	// 1. Stability check
	etag2 := transform.CreateTransformEtag(img, &params)
	if *etag1 != *etag2 {
		t.Errorf("ETags are not stable. Got %s and %s", *etag1, *etag2)
	}

	// 2. Sensitivity check
	params.Width = 101
	etag3 := transform.CreateTransformEtag(img, &params)
	if *etag1 == *etag3 {
		t.Errorf("ETag should change when params change")
	}
}