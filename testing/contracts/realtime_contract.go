package contracts

import (
	"context"
	"sync"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/realtime"
)

type contractMockClient struct {
	id       string
	received []realtime.Message
	mu       sync.Mutex
}

func (c *contractMockClient) ID() string {
	return c.id
}

func (c *contractMockClient) Send(msg realtime.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.received = append(c.received, msg)
	return nil
}

func (c *contractMockClient) Close() error {
	return nil
}

// VerifyRealtimeHubContract executes a standard compliance test suite against any realtime.Hub implementation.
func VerifyRealtimeHubContract(t *testing.T, hub realtime.Hub) {
	ctx := context.Background()

	t.Run("Subscribe_and_Broadcast", func(t *testing.T) {
		client := &contractMockClient{id: "contract-client"}
		unsub, err := hub.Subscribe(ctx, "contract-room", client)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}
		defer unsub()

		if count := hub.ClientCount("contract-room"); count != 1 {
			t.Errorf("ClientCount = %d, want 1", count)
		}

		msg := realtime.Message{
			Event:   "test.event",
			Channel: "contract-room",
			Data:    "payload",
		}

		err = hub.Broadcast(ctx, "contract-room", msg)
		if err != nil {
			t.Fatalf("Broadcast failed: %v", err)
		}

		if len(client.received) != 1 {
			t.Errorf("expected client to receive 1 message, got: %d", len(client.received))
		}
	})
}
