package entities

import "time"

// JobRun represents a persisted record for an enqueued job run.
type JobRun struct {
    ID          uint      `gorm:"primarykey" json:"-"`
    Uid         string    `gorm:"uniqueIndex" json:"uid"`
    Type        string    `json:"type"`
    Topic       string    `json:"topic"`
    ImageUid    *string   `json:"image_uid,omitempty"`
    Status      string    `json:"status"`
    Payload     *string   `gorm:"type:JSONB" json:"payload,omitempty"`
    EnqueuedAt  time.Time `json:"enqueued_at"`
    StartedAt   *time.Time `json:"started_at,omitempty"`
    CompletedAt *time.Time `json:"completed_at,omitempty"`
    Error       *string   `json:"error,omitempty"`
}
