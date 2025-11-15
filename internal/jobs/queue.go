package jobs

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	imalog "imagine/internal/logger"
)

var (
	allJobs        = make(map[string]*Job)
	allJobsMu      sync.RWMutex
	queuedCounts   = make(map[string]int)
	queuedCountsMu sync.RWMutex
)

var (
	PubSub *gochannel.GoChannel
	Router *message.Router
	Logger = watermill.NewSlogLogger(imalog.CreateLogger(imalog.SetupDefaultLogHandlers()))
)

// GetRunningJobs returns the current number of running jobs.
func GetRunningJobs() int {
	allJobsMu.RLock()
	defer allJobsMu.RUnlock()
	return len(allJobs)
}

func GetAllJobs() map[string]*Job {
	// Return a shallow copy to avoid callers mutating the internal map and
	// to prevent races when iterating without a lock.
	allJobsMu.RLock()
	defer allJobsMu.RUnlock()
	copy := make(map[string]*Job, len(allJobs))
	maps.Copy(copy, allJobs)
	return copy
}

// Publish is a wrapper around PubSub.Publish which tracks queued counts per topic.
func Publish(topic string, msg *message.Message) error {
	if PubSub == nil {
		return fmt.Errorf("pubsub not initialized")
	}

	queuedCountsMu.Lock()
	queuedCounts[topic] = queuedCounts[topic] + 1
	queuedCountsMu.Unlock()

	return PubSub.Publish(topic, msg)
}

// JobCounts describes running and queued counts by topic.
type JobCounts struct {
	Running        int64          `json:"running"`
	RunningByTopic map[string]int `json:"running_by_topic"`
	QueuedByTopic  map[string]int `json:"queued_by_topic"`
}

// GetCounts returns a snapshot of running and queued counts.
func GetCounts() JobCounts {
	jc := JobCounts{
		RunningByTopic: make(map[string]int),
		QueuedByTopic:  make(map[string]int),
	}

	// running
	allJobsMu.RLock()
	for _, j := range allJobs {
		jc.RunningByTopic[j.Topic()]++
	}
	jc.Running = int64(len(allJobs))
	allJobsMu.RUnlock()

	// queued
	queuedCountsMu.RLock()
	for k, v := range queuedCounts {
		jc.QueuedByTopic[k] = v
	}
	queuedCountsMu.RUnlock()

	return jc
}

// RegisterWorkers registers all JobWorkers with the router.
// Call this after initializing Router and PubSub, but before Router.Run().
func RegisterWorkers(workers ...*Worker) {

	for _, worker := range workers {
		handler := worker.Handler
		topic := worker.Topic
		cm := getOrCreateManager(topic)

		Router.AddConsumerHandler(
			worker.Name,
			topic,
			PubSub,
			func(msg *message.Message) error {

				cm.Acquire()
				defer cm.Release()

				worker.Start()
				job := &Job{
					ctx:    msg.Context(),
					ID:     msg.UUID,
					topic:  topic,
					status: JobStatusRunning,
				}

				if job.ID == "" {
					job.ID = watermill.NewUUID()
				}

				// Transition from queued -> running: decrement queued count for topic.
				queuedCountsMu.Lock()
				if v, ok := queuedCounts[topic]; ok && v > 0 {
					queuedCounts[topic] = v - 1
				}
				queuedCountsMu.Unlock()

				// Register running job in a thread-safe way.
				allJobsMu.Lock()
				allJobs[job.ID] = job
				allJobsMu.Unlock()

				defer func() {
					worker.Stop()
					job.SetStatus(JobStatusSuccess)
					allJobsMu.Lock()
					delete(allJobs, job.ID)
					allJobsMu.Unlock()
				}()

				return handler(msg)
			},
		)
	}
}

func RunJobQueue(workers ...*Worker) {
	var err error
	Router, err = message.NewRouter(message.RouterConfig{}, Logger)

	if err != nil {
		panic(err)
	}

	// SignalsHandler will gracefully shutdown Router when SIGTERM is received.
	// You can also close the router by just calling `r.Close()`.
	Router.AddPlugin(plugin.SignalsHandler)

	// Router level middleware are executed for every message sent to the router
	Router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,

		// The handler function is retried if it returns an error.
		// After MaxRetries, the message is Nacked and it's up to the PubSub to resend it.
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Second * 2,
			Logger:          Logger,
		}.Middleware,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,

		// middleware.NewThrottle(3, time.Second*3).Middleware,
	)

	// For simplicity, we are using the gochannel Pub/Sub here,
	// You can replace it with any Pub/Sub implementation, it will work the same.
	PubSub = gochannel.NewGoChannel(gochannel.Config{}, Logger)
	RegisterWorkers(workers...)

	// Now that all handlers are registered, we're running the Router.
	// Run is blocking while the router is running.
	ctx := context.Background()
	if err := Router.Run(ctx); err != nil {
		panic(err)
	}
}
