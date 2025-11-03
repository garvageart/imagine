package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
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
}

// SSEMessage represents a message to be sent via SSE
type SSEMessage struct {
	Event    string
	Data     any
	ClientID string // If empty, broadcast to all
}

// NewSSEBroker creates and starts a new SSE broker
func NewSSEBroker() *SSEBroker {
	broker := &SSEBroker{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan *SSEMessage, 100),
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
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	message := formatSSEMessage(eventType, jsonData)

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

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	message := formatSSEMessage(eventType, jsonData)

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
	fmt.Fprintf(w, "%s", formatSSEMessage("connected", jsonData))
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
			w.Write(formatSSEMessage("ping", []byte("{}")))
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
func formatSSEMessage(eventType string, data []byte) []byte {
	return fmt.Appendf(nil, "event: %s\ndata: %s\n\n", eventType, data)
}
