package jobs

import (
    "fmt"
    "sync"

    "gorm.io/gorm"
)

type CountFunc func(db *gorm.DB, command string, payload any) (int64, error)
type EnqueueFunc func(db *gorm.DB, command string, payload any) (int, error)
type CustomHandler any

type JobDescriptor struct {
    ID                 string        `json:"id"`
    Topic              string        `json:"topic"`
    DisplayName        string        `json:"display_name"`
    Description        string        `json:"description,omitempty"`
    Concurrency        int           `json:"concurrency"`
    Count              CountFunc     `json:"-"`
    Enqueue            EnqueueFunc   `json:"-"`
    CustomHandler      CustomHandler `json:"-"`
}

var (
    descriptorsMu sync.RWMutex
    descriptors   = map[string]*JobDescriptor{}
    persisted    = map[string]map[string]any{}
)

// RegisterJob registers a job descriptor in the in-memory registry and applies
// default concurrency.
func RegisterJob(desc *JobDescriptor) {
    descriptorsMu.Lock()
    defer descriptorsMu.Unlock()
    descriptors[desc.ID] = desc
}

// GetAll returns a shallow copy of all registered descriptors.
func GetAll() []*JobDescriptor {
    descriptorsMu.RLock()
    defer descriptorsMu.RUnlock()

    out := make([]*JobDescriptor, 0, len(descriptors))
    for _, d := range descriptors {
        out = append(out, d)
    }

    return out
}

// Find returns a descriptor by id, or nil.
func Find(id string) *JobDescriptor {
    descriptorsMu.RLock()
    defer descriptorsMu.RUnlock()

    if d, ok := descriptors[id]; ok {
        return d
    }

    return nil
}

// SetPersisted sets a key/value for a given job id and persists immediately.
func SetPersisted(jobID string, key string, value any) error {
    if persisted == nil {
        persisted = map[string]map[string]any{}
    }

    if _, ok := persisted[jobID]; !ok {
        persisted[jobID] = map[string]any{}
    }
	
    persisted[jobID][key] = value
    return nil
}

// GetPersisted returns a persisted value for a job id/key, or nil.
func GetPersisted(jobID string, key string) any {
    if v, ok := persisted[jobID]; ok {
        return v[key]
    }
    return nil
}

// For debugging convenience
func DumpRegistry() string {
    descriptorsMu.RLock()
    defer descriptorsMu.RUnlock()
    return fmt.Sprintf("%v", descriptors)
}
