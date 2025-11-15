package jobs

import (
	"fmt"
	"sync"
)

var (
	workersMu sync.RWMutex
	workers   = map[string]*Worker{}
	persisted = map[string]map[string]any{}
)

// RegisterWorker registers a worker in the in-memory registry. Safe to call multiple times.
func RegisterWorker(w *Worker) {
	if w == nil || w.Name == "" {
		return
	}
	workersMu.Lock()
	defer workersMu.Unlock()
	workers[w.Name] = w
}

// GetAllWorkers returns a slice copy of all registered workers.
func GetAllWorkers() []*Worker {
	workersMu.RLock()
	defer workersMu.RUnlock()

	out := make([]*Worker, 0, len(workers))
	for _, w := range workers {
		out = append(out, w)
	}
	return out
}

// FindWorker returns a registered worker by id, or nil if not found.
func FindWorker(id string) *Worker {
	workersMu.RLock()
	defer workersMu.RUnlock()
	if w, ok := workers[id]; ok {
		return w
	}
	return nil
}

// SetPersisted sets a key/value pair for a worker id in the in-memory persisted map.
func SetPersisted(workerID string, key string, value any) error {
	if persisted == nil {
		persisted = map[string]map[string]any{}
	}
	if _, ok := persisted[workerID]; !ok {
		persisted[workerID] = map[string]any{}
	}
	persisted[workerID][key] = value
	return nil
}

// GetPersisted returns a previously persisted value for a worker id/key.
func GetPersisted(workerID string, key string) any {
	if v, ok := persisted[workerID]; ok {
		return v[key]
	}
	return nil
}

// DumpRegistry returns a debug string representation of the registry.
func DumpRegistry() string {
	workersMu.RLock()
	defer workersMu.RUnlock()
	return fmt.Sprintf("%v", workers)
}
