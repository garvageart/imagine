package jobs

import (
	libhttp "imagine/internal/http"
)

// NewProgressCallback creates a reusable progress reporter closure that broadcasts
// generic job-progress WebSocket events with a consistent payload shape.
// If wsBroker is nil, the returned function is a no-op.
func NewProgressCallback(
	wsBroker *libhttp.WSBroker,
	jobId string,
	jobType string,
	imageId string,
	filename string,
) func(step string, progress int) {
	if wsBroker == nil {
		return func(step string, progress int) {}
	}

	return func(step string, progress int) {
		wsBroker.Broadcast("job-progress", map[string]interface{}{
			"jobId":    jobId,
			"type":     jobType,
			"imageId":  imageId,
			"filename": filename,
			"progress": progress,
			"status":   step,
			"step":     step,
		})
	}
}
