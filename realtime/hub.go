// Package realtime provides real-time event broadcasting, channels, and Server-Sent Events (SSE).
package realtime

import "context"

// Message defines the payload sent across real-time channels and rooms.
type Message struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

// Hub defines the contract for broadcasting messages to subscribers in named channels.
type Hub interface {
	// Broadcast publishes a message to all active clients subscribed to the channel.
	Broadcast(ctx context.Context, channel string, message Message) error

	// Subscribe attaches a client to a channel and returns an unsubscription function.
	Subscribe(ctx context.Context, channel string, client Client) (unsubscribe func(), err error)

	// ClientCount returns the number of active subscribers on a channel.
	ClientCount(channel string) int
}
