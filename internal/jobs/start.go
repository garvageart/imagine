package jobs

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

func Start() error {
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.Now().Location()))
	if err != nil {
		return fmt.Errorf("error creating scheduler: %w", err)
	}

	Jobs = make(map[string]gocron.Job)
	Scheduler = scheduler

	return nil
}

func Shutdown() error {
	return Scheduler.Shutdown()
}
