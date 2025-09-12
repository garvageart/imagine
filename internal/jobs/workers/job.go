package workers

import (
	"context"
)

type Job struct {
	ctx context.Context
}

func (j *Job) SetContext(ctx context.Context) {
	j.ctx = ctx
}

func (j *Job) Context() context.Context {
	return j.ctx
}
