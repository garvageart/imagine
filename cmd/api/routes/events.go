package routes

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
	"log/slog"

	libhttp "imagine/internal/http"
)

// EventHistory stores recent events for replay
type EventHistory struct {
	mu     sync.RWMutex
	events []EventRecord
	maxLen int
}

type EventRecord struct {
	Timestamp time.Time              `json:"timestamp"`
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
}

var eventHistory = &EventHistory{
	events: make([]EventRecord, 0),
	maxLen: 100, // Keep last 100 events
}

func (h *EventHistory) Add(event string, data map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	record := EventRecord{
		Timestamp: time.Now(),
		Event:     event,
		Data:      data,
	}

	h.events = append(h.events, record)
	if len(h.events) > h.maxLen {
		h.events = h.events[1:]
	}
}

func (h *EventHistory) GetRecent(limit int) []EventRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if limit <= 0 || limit > len(h.events) {
		limit = len(h.events)
	}

	// Return most recent events
	start := len(h.events) - limit
	result := make([]EventRecord, limit)
	copy(result, h.events[start:])
	return result
}

func (h *EventHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = make([]EventRecord, 0)
}

// EventsRouter creates routes for SSE event management
func EventsRouter(db *gorm.DB, logger *slog.Logger, sseBroker *libhttp.SSEBroker) http.Handler {
	router := chi.NewRouter()

	router.Get("/", sseBroker.ServeHTTP)

	// Get SSE connection statistics
	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"connectedClients": sseBroker.GetClientCount(),
			"clientIds":        sseBroker.GetClientIDs(),
			"timestamp":        time.Now(),
		}
		render.JSON(w, r, stats)
	})

	// Get event history
	router.Get("/history", func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
				limit = parsed
				if limit > 100 {
					limit = 100
				}
			}
		}

		history := eventHistory.GetRecent(limit)
		render.JSON(w, r, map[string]interface{}{
			"events": history,
			"count":  len(history),
		})
	})

	// Clear event history (admin only)
	router.Delete("/history", func(w http.ResponseWriter, r *http.Request) {
		eventHistory.Clear()
		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"message": "Event history cleared",
		})
	})

	// Test endpoint to broadcast a message
	router.Post("/broadcast", func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Event string                 `json:"event"`
			Data  map[string]interface{} `json:"data"`
		}

		if err := render.DecodeJSON(r.Body, &payload); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid JSON payload"})
			return
		}

		if payload.Event == "" {
			payload.Event = "message"
		}

		// Add to history
		eventHistory.Add(payload.Event, payload.Data)

		err := sseBroker.Broadcast(payload.Event, payload.Data)
		if err != nil {
			logger.Error("Failed to broadcast message", slog.String("error", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to broadcast message"})
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"message": "Message broadcasted successfully",
			"clients": sseBroker.GetClientCount(),
		})
	})

	// Test endpoint to send message to specific client
	router.Post("/send/{clientId}", func(w http.ResponseWriter, r *http.Request) {
		clientID := chi.URLParam(r, "clientId")

		var payload struct {
			Event string                 `json:"event"`
			Data  map[string]interface{} `json:"data"`
		}

		if err := render.DecodeJSON(r.Body, &payload); err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "Invalid JSON payload"})
			return
		}

		if payload.Event == "" {
			payload.Event = "message"
		}

		err := sseBroker.SendToClient(clientID, payload.Event, payload.Data)
		if err != nil {
			logger.Error("Failed to send message to client",
				slog.String("clientId", clientID),
				slog.String("error", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "Failed to send message"})
			return
		}

		render.JSON(w, r, map[string]interface{}{
			"success":  true,
			"message":  "Message sent successfully",
			"clientId": clientID,
		})
	})

	// Metrics endpoint for monitoring
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		history := eventHistory.GetRecent(100)
		
		// Count events by type
		eventCounts := make(map[string]int)
		for _, record := range history {
			eventCounts[record.Event]++
		}

		render.JSON(w, r, map[string]interface{}{
			"connectedClients": sseBroker.GetClientCount(),
			"totalEvents":      len(history),
			"eventsByType":     eventCounts,
			"timestamp":        time.Now(),
		})
	})

	return router
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}
