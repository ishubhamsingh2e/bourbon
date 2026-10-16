package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/ishubhamsingh2e/bourbon/realtime"
)

// Hub implements realtime.Hub backed by Redis Pub/Sub.
type Hub struct {
	client Client
	prefix string
	mu     sync.RWMutex
	counts map[string]int
}

// NewHub constructs a Redis Pub/Sub Hub adapter.
func NewHub(client Client, prefix ...string) *Hub {
	p := "bourbon:realtime:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &Hub{
		client: client,
		prefix: p,
		counts: make(map[string]int),
	}
}

func (h *Hub) channelKey(ch string) string {
	return h.prefix + ch
}

// Broadcast serializes and publishes a message to Redis Pub/Sub.
func (h *Hub) Broadcast(ctx context.Context, channel string, message realtime.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return h.client.Publish(ctx, h.channelKey(channel), data)
}

// Subscribe listens on Redis channel and forwards messages to client.
func (h *Hub) Subscribe(ctx context.Context, channel string, client realtime.Client) (func(), error) {
	h.mu.Lock()
	h.counts[channel]++
	h.mu.Unlock()

	unsubRedis, err := h.client.Subscribe(ctx, h.channelKey(channel), func(msgBytes []byte) {
		var msg realtime.Message
		if err := json.Unmarshal(msgBytes, &msg); err == nil {
			_ = client.Send(msg)
		}
	})
	if err != nil {
		return nil, err
	}

	return func() {
		h.mu.Lock()
		h.counts[channel]--
		if h.counts[channel] <= 0 {
			delete(h.counts, channel)
		}
		h.mu.Unlock()
		unsubRedis()
	}, nil
}

// ClientCount returns subscriber count for the channel.
func (h *Hub) ClientCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.counts[channel]
}
