// Package inprocess provides an in-memory topic and room broadcasting Hub.
package inprocess

import (
	"context"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/realtime"
)

// Hub implements realtime.Hub in-memory using synchronized client maps.
type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[string]realtime.Client
}

// New constructs an in-memory real-time Hub.
func New() *Hub {
	return &Hub{
		channels: make(map[string]map[string]realtime.Client),
	}
}

// Broadcast transmits a message to all active clients in the specified channel.
func (h *Hub) Broadcast(ctx context.Context, channel string, message realtime.Message) error {
	h.mu.RLock()
	clientsMap, exists := h.channels[channel]
	if !exists || len(clientsMap) == 0 {
		h.mu.RUnlock()
		return nil
	}

	clients := make([]realtime.Client, 0, len(clientsMap))
	for _, c := range clientsMap {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		_ = client.Send(message)
	}

	return nil
}

// Subscribe attaches a client to a channel and returns an unsubscription function.
func (h *Hub) Subscribe(ctx context.Context, channel string, client realtime.Client) (func(), error) {
	h.mu.Lock()
	if _, exists := h.channels[channel]; !exists {
		h.channels[channel] = make(map[string]realtime.Client)
	}
	h.channels[channel][client.ID()] = client
	h.mu.Unlock()

	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if clients, exists := h.channels[channel]; exists {
			delete(clients, client.ID())
			if len(clients) == 0 {
				delete(h.channels, channel)
			}
		}
	}, nil
}

// ClientCount returns the number of active clients in the channel.
func (h *Hub) ClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, exists := h.channels[channel]; exists {
		return len(clients)
	}
	return 0
}
