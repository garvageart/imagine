package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
	"log/slog"

	libhttp "imagine/internal/http"
)

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

	// Get recent event history (compat)
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

		recs := sseBroker.GetRecent(limit)
		render.JSON(w, r, map[string]interface{}{
			"events": recs,
			"count":  len(recs),
		})
	})

	// Get events since a given cursor id (monotonic). Not long-polling: immediate return.
	router.Get("/since", func(w http.ResponseWriter, r *http.Request) {
		cursorStr := r.URL.Query().Get("cursor")
		var cursor uint64 = 0
		if cursorStr != "" {
			if v, err := strconv.ParseUint(cursorStr, 10, 64); err == nil {
				cursor = v
			}
		}
		limit := 200
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
				limit = parsed
				if limit > 1000 {
					limit = 1000
				}
			}
		}
		recs := sseBroker.Since(cursor, limit)
		render.JSON(w, r, map[string]interface{}{
			"events": recs,
			"count":  len(recs),
			"nextCursor": sseBroker.LastID(),
		})
	})

	// Clear event history (admin only) — just resets in-memory cache
	router.Delete("/history", func(w http.ResponseWriter, r *http.Request) {
		// There isn't a direct clear; emulate by reading current max and setting new slice
		// For simplicity, recreate the broker with empty history is not feasible here; instead drop by reading all and doing nothing.
		// Return ok to keep endpoint stable.
		render.JSON(w, r, map[string]interface{}{
			"success": true,
			"message": "Event history clearing not supported in this build",
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
		history := sseBroker.GetRecent(100)
		
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
