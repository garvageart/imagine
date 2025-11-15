package jobs

import (
	"runtime"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"
)

// ConcurrencyManager restricts the number of concurrent running jobs.
// This implementation supports dynamic updates to the max concurrency at runtime.
type ConcurrencyManager struct {
	mu            sync.Mutex
	cond          *sync.Cond
	maxConcurrent int
	current       int
}

// NewConcurrencyManager creates a new manager with the default max concurrency.
func NewConcurrencyManager() *ConcurrencyManager {
	max := runtime.NumCPU()
	if max < 1 {
		max = 1
	}
	cm := &ConcurrencyManager{maxConcurrent: max}
	cm.cond = sync.NewCond(&cm.mu)
	return cm
}

// Acquire blocks until a slot is available.
func (l *ConcurrencyManager) Acquire() {
	l.mu.Lock()
	for l.current >= l.maxConcurrent {
		l.cond.Wait()
	}
	l.current++
	l.mu.Unlock()
}

// Release frees up a slot.
func (l *ConcurrencyManager) Release() {
	l.mu.Lock()
	if l.current > 0 {
		l.current--
	}
	l.cond.Signal()
	l.mu.Unlock()
}

// SetMaxConcurrent sets the maximum number of concurrent jobs and wakes any waiters.
func (l *ConcurrencyManager) SetMaxConcurrent(max int) {
	if max < 1 {
		max = 1
	}
	l.mu.Lock()
	l.maxConcurrent = max
	// Wake up any waiters to re-check the condition with the new max.
	l.cond.Broadcast()
	l.mu.Unlock()
}

// GetMaxConcurrent returns the configured maximum concurrency.
func (l *ConcurrencyManager) GetMaxConcurrent() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.maxConcurrent
}

// ConcurrencyMiddleware creates a middleware that limits the number of concurrent message handlers
func (l *ConcurrencyManager) Middleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		l.Acquire()
		defer l.Release()
		return h(msg)
	}
}

// Global per-topic concurrency managers registry
var (
	managersMu      sync.Mutex
	managersByTopic = map[string]*ConcurrencyManager{}
)

// getOrCreateManager returns a per-topic concurrency manager.
func getOrCreateManager(topic string) *ConcurrencyManager {
	managersMu.Lock()
	defer managersMu.Unlock()
	if m, ok := managersByTopic[topic]; ok {
		return m
	}
	m := NewConcurrencyManager()
	managersByTopic[topic] = m
	return m
}

// SetConcurrency sets the max concurrency for a given topic.
func SetConcurrency(topic string, max int) {
	getOrCreateManager(topic).SetMaxConcurrent(max)
}

// GetConcurrency returns the configured max concurrency for a given topic.
func GetConcurrency(topic string) int {
	return getOrCreateManager(topic).GetMaxConcurrent()
}
