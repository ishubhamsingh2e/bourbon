// Package redis provides Redis-backed adapters for Cache, Session, Jobs, and Realtime.
package redis

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// Client defines the common key-value operations required by Redis adapters.
type Client interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	FlushAll(ctx context.Context) error
	IncrBy(ctx context.Context, key string, val int64) (int64, error)
	RPush(ctx context.Context, key string, val []byte) error
	LPop(ctx context.Context, key string) ([]byte, bool, error)
	LLen(ctx context.Context, key string) (int64, error)
	Publish(ctx context.Context, channel string, msg []byte) error
	Subscribe(ctx context.Context, channel string, handler func(msg []byte)) (unsubscribe func(), err error)
}

type memoryEntry struct {
	val       []byte
	expiresAt time.Time
}

// MemoryClient implements Client in-memory for testing and development.
type MemoryClient struct {
	mu          sync.RWMutex
	store       map[string]memoryEntry
	lists       map[string][][]byte
	subscribers map[string][]func(msg []byte)
}

// NewMemoryClient constructs an in-memory Redis client mock.
func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		store:       make(map[string]memoryEntry),
		lists:       make(map[string][][]byte),
		subscribers: make(map[string][]func(msg []byte)),
	}
}

func (m *MemoryClient) Get(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.store[key]
	if !ok {
		return nil, false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return nil, false, nil
	}
	return entry.val, true, nil
}

func (m *MemoryClient) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	m.store[key] = memoryEntry{val: val, expiresAt: exp}
	return nil
}

func (m *MemoryClient) Delete(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.store, k)
		delete(m.lists, k)
	}
	return nil
}

func (m *MemoryClient) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]memoryEntry)
	m.lists = make(map[string][][]byte)
	return nil
}

func (m *MemoryClient) IncrBy(ctx context.Context, key string, val int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var current int64
	if entry, ok := m.store[key]; ok {
		parsed, err := strconv.ParseInt(string(entry.val), 10, 64)
		if err == nil {
			current = parsed
		}
	}
	current += val
	m.store[key] = memoryEntry{val: []byte(strconv.FormatInt(current, 10))}
	return current, nil
}

func (m *MemoryClient) RPush(ctx context.Context, key string, val []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lists[key] = append(m.lists[key], val)
	return nil
}

func (m *MemoryClient) LPop(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	list := m.lists[key]
	if len(list) == 0 {
		return nil, false, nil
	}

	val := list[0]
	m.lists[key] = list[1:]
	return val, true, nil
}

func (m *MemoryClient) LLen(ctx context.Context, key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.lists[key])), nil
}

func (m *MemoryClient) Publish(ctx context.Context, channel string, msg []byte) error {
	m.mu.RLock()
	subs := append([]func(msg []byte){}, m.subscribers[channel]...)
	m.mu.RUnlock()

	for _, sub := range subs {
		sub(msg)
	}
	return nil
}

func (m *MemoryClient) Subscribe(ctx context.Context, channel string, handler func(msg []byte)) (func(), error) {
	m.mu.Lock()
	m.subscribers[channel] = append(m.subscribers[channel], handler)
	m.mu.Unlock()

	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		subs := m.subscribers[channel]
		for i := range subs {
			m.subscribers[channel] = append(subs[:i], subs[i+1:]...)
			break
		}
	}, nil
}
