package jobs

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"gorm.io/gorm"

	"imagine/internal/entities"
	"imagine/internal/utils"
)

const (
	WorkerJobStatusQueued  = "queued"
	WorkerJobStatusRunning = "running"
	WorkerJobStatusFailed  = "failed"
	WorkerJobStatusSuccess = "completed"
)

// Enqueue creates a persisted WorkerJob and publishes the message to the
// router. It returns the created WorkerJob UID which is also set as the
// Watermill message UUID.
func Enqueue(db *gorm.DB, topic string, payload any, cmd *JobCommand, imageUid *string) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	uid := watermill.NewUUID()

	var cmdStr *string
	if cmd != nil {
		s := string(*cmd)
		cmdStr = &s
	}

	payloadStr := string(payloadBytes)
	// Truncate payload for DB storage if too large
	if len(payloadStr) > 10_000 {
		payloadStr = payloadStr[:10_000]
	}

	wj := entities.WorkerJob{
		Uid:        uid,
		Type:       topic,
		Topic:      topic,
		Command:    cmdStr,
		ImageUid:   imageUid,
		Status:     WorkerJobStatusQueued,
		Payload:    &payloadStr,
		EnqueuedAt: time.Now().UTC(),
	}

	if err := db.Create(&wj).Error; err != nil {
		return "", fmt.Errorf("failed to persist worker job: %w", err)
	}

	msg := message.NewMessage(uid, payloadBytes)
	msg.Metadata.Set("X-Worker-Job-Uid", uid)

	if err := Publish(topic, msg); err != nil {
		_ = UpdateWorkerJobStatus(db, uid, WorkerJobStatusFailed, utils.StringPtr("publish_failed"), utils.StringPtr("failed to publish message"), nil, nil)
		return uid, fmt.Errorf("publish: %w", err)
	}

	return uid, nil
}

// UpdateWorkerJobStatus updates WorkerJob status and optional timestamps and error info.
func UpdateWorkerJobStatus(db *gorm.DB, uid string, status string, errorCode *string, errorMsg *string, startedAt *time.Time, completedAt *time.Time) error {
	updates := map[string]any{"status": status}
	if errorCode != nil {
		updates["error_code"] = *errorCode
	}

	if errorMsg != nil {
		// truncate to reasonable length
		msg := *errorMsg
		if len(msg) > 1024 {
			msg = msg[:1024]
		}
		updates["error_msg"] = msg
	}

	if startedAt != nil {
		updates["started_at"] = *startedAt
	}

	if completedAt != nil {
		updates["completed_at"] = *completedAt
	}

	if err := db.Model(&entities.WorkerJob{}).Where("uid = ?", uid).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update worker job: %w", err)
	}
	return nil
}
