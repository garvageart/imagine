package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"sync/atomic"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID      string
	Channel chan []byte
	Request *http.Request
}

// SSEBroker manages SSE connections and message broadcasting
type SSEBroker struct {
	clients    map[string]*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan *SSEMessage
	mu         sync.RWMutex

	lastID     uint64
	history    []SSERecord
	historyMu  sync.RWMutex
	historyMax int
}

// SSEMessage represents a message to be sent via SSE
type SSEMessage struct {
	Event    string
	Data     any
	ClientID string // If empty, broadcast to all
	ID       uint64 // Monotonic id for SSE 'id:' field
}

// SSERecord is a stored history record
type SSERecord struct {
	ID        uint64      `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Event     string      `json:"event"`
	Data      any `json:"data"`
}

// NewSSEBroker creates and starts a new SSE broker
func NewSSEBroker() *SSEBroker {
	broker := &SSEBroker{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan *SSEMessage, 100),
		history:    make([]SSERecord, 0, 512),
		historyMax: 512,
	}
	go broker.run()
	return broker
}

// run manages the broker's event loop
func (b *SSEBroker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.ID] = client
			b.mu.Unlock()

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client.ID]; ok {
				close(client.Channel)
				delete(b.clients, client.ID)
			}
			b.mu.Unlock()

		case message := <-b.broadcast:
			if message.ClientID != "" {
				// Send to specific client
				b.sendToClient(message.ClientID, message.Event, message.Data)
			} else {
				// Broadcast to all clients
				b.broadcastToAll(message.Event, message.Data)
			}
		}
	}
}

// broadcastToAll sends a message to all connected clients
func (b *SSEBroker) broadcastToAll(eventType string, data any) {
	// Assign id and append to history once for this broadcast
	id := atomic.AddUint64(&b.lastID, 1)
	b.appendHistory(id, eventType, data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	message := formatSSEMessage(eventType, id, jsonData)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, client := range b.clients {
		select {
		case client.Channel <- message:
		default:
			// Client channel full, skip
		}
	}
}

// sendToClient sends a message to a specific client
func (b *SSEBroker) sendToClient(clientID, eventType string, data any) {
	b.mu.RLock()
	client, ok := b.clients[clientID]
	b.mu.RUnlock()

	if !ok {
		return
	}

	// Assign id and append to history for this event as well
	id := atomic.AddUint64(&b.lastID, 1)
	b.appendHistory(id, eventType, data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	message := formatSSEMessage(eventType, id, jsonData)

	select {
	case client.Channel <- message:
	default:
		// Client channel full, skip
	}
}

// Broadcast sends an event to all connected clients
func (b *SSEBroker) Broadcast(eventType string, data interface{}) error {
	select {
	case b.broadcast <- &SSEMessage{
		Event: eventType,
		Data:  data,
	}:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// SendToClient sends an event to a specific client by ID
func (b *SSEBroker) SendToClient(clientID, eventType string, data interface{}) error {
	select {
	case b.broadcast <- &SSEMessage{
		Event:    eventType,
		Data:     data,
		ClientID: clientID,
	}:
		return nil
	default:
		return fmt.Errorf("broadcast channel full")
	}
}

// GetClientCount returns the number of connected clients
func (b *SSEBroker) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// GetClientIDs returns all connected client IDs
func (b *SSEBroker) GetClientIDs() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ids := make([]string, 0, len(b.clients))
	for id := range b.clients {
		ids = append(ids, id)
	}
	return ids
}

// ServeHTTP handles SSE connections
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientID := GetRequestID(r)
	client := &SSEClient{
		ID:      clientID,
		Channel: make(chan []byte, 10),
		Request: r,
	}

	b.register <- client
	defer func() {
		b.unregister <- client
	}()

	connectionData := map[string]interface{}{
		"clientId": clientID,
		"message":  "Connected to SSE stream",
	}
	jsonData, _ := json.Marshal(connectionData)
	// Emit connected with synthetic id (doesn't increase lastID)
	fmt.Fprintf(w, "%s", formatSSEMessage("connected", 0, jsonData))
	w.Write([]byte("retry: 10000\n\n"))
	flusher.Flush()

	// Keep connection alive and send messages
	// Heartbeat ticker prevents idle timeouts by intermediaries/browsers
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			w.Write(formatSSEMessage("ping", 0, []byte("{}")))
			flusher.Flush()
		case msg, ok := <-client.Channel:
			if !ok {
				return
			}
			w.Write(msg)
			flusher.Flush()
		}
	}
}

// formatSSEMessage formats data in SSE protocol format
func formatSSEMessage(eventType string, id uint64, data []byte) []byte {
	// Include id if non-zero so clients can use Last-Event-ID semantics
	if id > 0 {
		return fmt.Appendf(nil, "id: %s\nevent: %s\ndata: %s\n\n", strconv.FormatUint(id, 10), eventType, data)
	}
	return fmt.Appendf(nil, "event: %s\ndata: %s\n\n", eventType, data)
}

// appendHistory appends to in-memory history with max length
func (b *SSEBroker) appendHistory(id uint64, eventType string, data any) {
	b.historyMu.Lock()
	b.history = append(b.history, SSERecord{ID: id, Timestamp: time.Now(), Event: eventType, Data: data})
	if len(b.history) > b.historyMax {
		// drop oldest
		drop := len(b.history) - b.historyMax
		if drop < 0 { drop = 0 }
		b.history = b.history[drop:]
	}
	b.historyMu.Unlock()
}

// GetRecent returns the most recent up to limit records
func (b *SSEBroker) GetRecent(limit int) []SSERecord {
	b.historyMu.RLock()
	defer b.historyMu.RUnlock()
	n := len(b.history)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]SSERecord, limit)
	copy(out, b.history[n-limit:])
	return out
}

// Since returns records with id > cursor up to limit
func (b *SSEBroker) Since(cursor uint64, limit int) []SSERecord {
	b.historyMu.RLock()
	defer b.historyMu.RUnlock()
	out := make([]SSERecord, 0, limit)
	for _, rec := range b.history {
		if rec.ID > cursor {
			out = append(out, rec)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out
}

// LastID returns the last assigned event id
func (b *SSEBroker) LastID() uint64 { return atomic.LoadUint64(&b.lastID) }
