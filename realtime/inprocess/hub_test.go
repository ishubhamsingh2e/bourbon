package inprocess

import (
	"context"
	"sync"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/realtime"
)

type mockClient struct {
	id       string
	received []realtime.Message
	mu       sync.Mutex
}

func (m *mockClient) ID() string {
	return m.id
}

func (m *mockClient) Send(msg realtime.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.received = append(m.received, msg)
	return nil
}

func (m *mockClient) Close() error {
	return nil
}

func TestInProcessHubBroadcasting(t *testing.T) {
	hub := New()
	ctx := context.Background()

	client1 := &mockClient{id: "client-1"}
	client2 := &mockClient{id: "client-2"}

	unsub1, err := hub.Subscribe(ctx, "chat-room", client1)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	unsub2, _ := hub.Subscribe(ctx, "chat-room", client2)

	if count := hub.ClientCount("chat-room"); count != 2 {
		t.Errorf("ClientCount = %d, want 2", count)
	}

	msg := realtime.Message{
		Event:   "new_message",
		Channel: "chat-room",
		Data:    map[string]string{"text": "hello"},
	}

	// Broadcast to room
	err = hub.Broadcast(ctx, "chat-room", msg)
	if err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}

	if len(client1.received) != 1 || len(client2.received) != 1 {
		t.Errorf("expected both clients to receive message, got %d and %d", len(client1.received), len(client2.received))
	}

	// Unsubscribe client1
	unsub1()
	if count := hub.ClientCount("chat-room"); count != 1 {
		t.Errorf("ClientCount after unsub = %d, want 1", count)
	}

	unsub2()
	if count := hub.ClientCount("chat-room"); count != 0 {
		t.Errorf("ClientCount after all unsub = %d, want 0", count)
	}
}
