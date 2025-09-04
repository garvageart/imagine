package jobs

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

var (
	Jobs      map[string]gocron.Job
	Scheduler gocron.Scheduler
)

func CreateJob(name, schedule string, handler func(), handlerParams ...any) error {
	if schedule == "" {
		return fmt.Errorf("schedule cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	job, err := Scheduler.NewJob(
		gocron.CronJob(
			schedule,
			false,
		),
		gocron.NewTask(
			handler,
			handlerParams...,
		),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)

	if err != nil {
		return err
	}

	Jobs[name] = job

	return nil
}
