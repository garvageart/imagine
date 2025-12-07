package routes_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"

	"imagine/api/routes"
	"imagine/internal/dto"
	"imagine/internal/entities"
	"imagine/internal/images"
)

// Helper function to create a new test logger
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// Helper function to create a new in-memory SQLite database
func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	// Auto-migrate all entities
	err = db.AutoMigrate(
		&entities.User{},
		&entities.Image{},
		&entities.Collection{},
		&entities.Session{},
		&entities.APIKey{},
		&entities.DownloadToken{},
		&entities.WorkerJob{},
		&entities.UserWithPassword{},
		&entities.SettingDefault{},
		&entities.SettingOverride{},
	)
	assert.NoError(t, err)
	return db
}

func TestAdminSystemStats(t *testing.T) {
	db := newTestDB(t)
	logger := newTestLogger()
	storageStats := images.NewStorageStatsHolder(os.TempDir()) // Use temp dir for stats

	r := chi.NewRouter()
	r.Mount("/admin", routes.AdminRouter(db, logger, storageStats))

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make request to the system stats endpoint
	resp, err := ts.Client().Get(ts.URL + "/admin/system/stats")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats dto.SystemStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&stats)
	assert.NoError(t, err)

	// Assertions for SystemStatsResponse
	assert.GreaterOrEqual(t, stats.UptimeSeconds, int64(0))
	assert.GreaterOrEqual(t, stats.NumGoroutine, 1) // At least one goroutine should be running
	assert.GreaterOrEqual(t, stats.AllocMemory, int64(0))
	assert.GreaterOrEqual(t, stats.SysMemory, int64(0))
	assert.GreaterOrEqual(t, stats.StorageUsedBytes, int64(0))
	assert.NotEmpty(t, stats.StoragePath)
	assert.NotNil(t, stats.TotalSystemSpaceBytes)
	assert.GreaterOrEqual(t, *stats.TotalSystemSpaceBytes, uint64(0))
}

func TestAdminDatabaseStats(t *testing.T) {
	db := newTestDB(t)
	logger := newTestLogger()
	storageStats := images.NewStorageStatsHolder(os.TempDir())

	r := chi.NewRouter()
	r.Mount("/admin", routes.AdminRouter(db, logger, storageStats))

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make request to the database stats endpoint
	resp, err := ts.Client().Get(ts.URL + "/admin/db/stats")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats dto.DatabaseStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&stats)
	assert.NoError(t, err)

	// Assertions for DatabaseStatsResponse
	assert.GreaterOrEqual(t, stats.UserCount, int64(0))
	assert.GreaterOrEqual(t, stats.ImageCount, int64(0))
	assert.GreaterOrEqual(t, stats.CollectionCount, int64(0))

	// DbSizeBytes and ActiveConnections are Postgres-specific and can be nil for SQLite
	if db.Dialector.Name() == "postgres" {
		assert.NotNil(t, stats.DbSizeBytes)
		assert.GreaterOrEqual(t, *stats.DbSizeBytes, int64(0))
		assert.NotNil(t, stats.ActiveConnections)
		assert.GreaterOrEqual(t, *stats.ActiveConnections, int64(0))
	} else {
		assert.Nil(t, stats.DbSizeBytes)
		assert.Nil(t, stats.ActiveConnections)
	}
}
