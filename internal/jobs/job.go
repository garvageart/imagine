package jobs

import "context"

type Job struct {
	ctx    context.Context
	ID     string
	topic  string
	status string
}

func (j *Job) SetStatus(status string) {
	j.status = status
}

func (j *Job) GetStatus() string {
	return j.status
}

func (j *Job) Topic() string {
	return j.topic
}

func (j *Job) SetContext(ctx context.Context) {
	j.ctx = ctx
}

func (j *Job) Context() context.Context {
	return j.ctx
}

func (j *Job) SetID(id string) {
	j.ID = id
}

func (j *Job) GetID() string {
	return j.ID
}
