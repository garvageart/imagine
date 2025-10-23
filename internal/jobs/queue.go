package jobs

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"

	imalog "imagine/internal/logger"
)

var (
	allJobs = make(map[string]*Job)
)

var (
	PubSub *gochannel.GoChannel
	Router *message.Router
	Logger = watermill.NewSlogLogger(imalog.CreateLogger(imalog.SetupDefaultLogHandlers()))
)

// GetRunningJobs returns the current number of running jobs.
func GetRunningJobs() int {
	return len(allJobs)
}

func GetAllJobs() map[string]*Job {
	return allJobs
}

// RegisterWorkers registers all JobWorkers with the router.
// Call this after initializing Router and PubSub, but before Router.Run().
func RegisterWorkers(workers ...*Worker) {

	for _, worker := range workers {
		handler := worker.Handler
		topic := worker.Topic

		Router.AddConsumerHandler(
			worker.Name,
			topic,
			PubSub,
			func(msg *message.Message) error {
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

				allJobs[job.ID] = job

				defer func() {
					worker.Stop()
					job.SetStatus(JobStatusSuccess)
					delete(allJobs, job.ID)
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

		NewConcurrencyManager().Middleware,

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
