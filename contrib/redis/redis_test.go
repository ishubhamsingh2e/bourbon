package redis

import (
	"context"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/jobs"
	"github.com/ishubhamsingh2e/bourbon/realtime"
	"github.com/ishubhamsingh2e/bourbon/session"
)

type testClient struct {
	id       string
	messages []realtime.Message
}

func (c *testClient) ID() string { return c.id }
func (c *testClient) Send(msg realtime.Message) error {
	c.messages = append(c.messages, msg)
	return nil
}
func (c *testClient) Close() error { return nil }

func TestRedisAdapters(t *testing.T) {
	client := NewMemoryClient()
	ctx := context.Background()

	// 1. Cache
	cacheAdapter := NewCache(client)
	err := cacheAdapter.Set(ctx, "user:1", map[string]string{"name": "Alice"}, time.Minute)
	if err != nil {
		t.Fatalf("Cache.Set failed: %v", err)
	}

	var user map[string]string
	err = cacheAdapter.Get(ctx, "user:1", &user)
	if err != nil || user["name"] != "Alice" {
		t.Fatalf("Cache.Get mismatch: %v, %v", err, user)
	}

	// 2. Session
	sessionStore := NewSessionStore(client)
	sess := session.New("sess_123")
	sess.Set("role", "admin")
	err = sessionStore.Save(ctx, sess, time.Hour)
	if err != nil {
		t.Fatalf("Session.Save failed: %v", err)
	}

	loaded, err := sessionStore.Load(ctx, "sess_123")
	val, ok := loaded.Get("role")
	if err != nil || !ok || val != "admin" {
		t.Fatalf("Session.Load mismatch: %v, %v, %v", ok, err, loaded)
	}

	// 3. Queue
	queueAdapter := NewQueue(client)
	job := jobs.NewJob("send_mail", []byte("hello"), nil)
	err = queueAdapter.Enqueue(ctx, job)
	if err != nil {
		t.Fatalf("Queue.Enqueue failed: %v", err)
	}

	dequeued, err := queueAdapter.Dequeue(ctx, "default")
	if err != nil || dequeued.Name() != "send_mail" {
		t.Fatalf("Queue.Dequeue mismatch: %v, %v", err, dequeued)
	}

	// 4. Realtime Hub
	hubAdapter := NewHub(client)
	mockClient := &testClient{id: "client_1"}
	unsub, err := hubAdapter.Subscribe(ctx, "notifications", mockClient)
	if err != nil {
		t.Fatalf("Hub.Subscribe failed: %v", err)
	}

	if count := hubAdapter.ClientCount("notifications"); count != 1 {
		t.Errorf("ClientCount = %d, want 1", count)
	}

	err = hubAdapter.Broadcast(ctx, "notifications", realtime.Message{
		Event:   "alert",
		Channel: "notifications",
		Data:    "server alert",
	})
	if err != nil {
		t.Fatalf("Hub.Broadcast failed: %v", err)
	}

	unsub()
	if count := hubAdapter.ClientCount("notifications"); count != 0 {
		t.Errorf("ClientCount after unsub = %d, want 0", count)
	}
}
