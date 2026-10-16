// Package memstore provides an in-core, zero-dependency in-memory session store reference adapter.
package memstore

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/session"
)

type storedSession struct {
	Values    map[string]any `json:"values"`
	Flashes   map[string]any `json:"flashes"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// MemStore provides a thread-safe, in-memory session store implementation.
type MemStore struct {
	mu     sync.RWMutex
	data   map[string][]byte
	stopCh chan struct{}
}

// New initializes an in-memory session store and starts periodic cleanup.
func New(cleanupInterval ...time.Duration) *MemStore {
	interval := time.Minute
	if len(cleanupInterval) > 0 && cleanupInterval[0] > 0 {
		interval = cleanupInterval[0]
	}

	ms := &MemStore{
		data:   make(map[string][]byte),
		stopCh: make(chan struct{}),
	}

	go ms.evictionLoop(interval)

	return ms
}

func (m *MemStore) evictionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case now := <-ticker.C:
			m.mu.Lock()
			for id, raw := range m.data {
				var st storedSession
				if err := json.Unmarshal(raw, &st); err == nil {
					if !st.ExpiresAt.IsZero() && now.After(st.ExpiresAt) {
						delete(m.data, id)
					}
				}
			}
			m.mu.Unlock()
		}
	}
}

// Close stops background cleaner routines and clears stored session data.
func (m *MemStore) Close() error {
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
	return nil
}

// Load hydrates a Session from storage by ID, or creates a new one if missing or expired.
func (m *MemStore) Load(ctx context.Context, id string) (*session.Session, error) {
	if id == "" {
		return session.New(""), nil
	}

	m.mu.RLock()
	raw, ok := m.data[id]
	m.mu.RUnlock()

	if !ok {
		return session.New(id), nil
	}

	var st storedSession
	if err := json.Unmarshal(raw, &st); err != nil {
		return session.New(id), nil
	}

	if !st.ExpiresAt.IsZero() && time.Now().After(st.ExpiresAt) {
		_ = m.Destroy(ctx, id)
		return session.New(id), nil
	}

	s := session.New(id)
	s.Restore(st.Values, st.Flashes)
	return s, nil
}

// Save serializes session values and flash attributes to storage.
func (m *MemStore) Save(ctx context.Context, s *session.Session, ttl time.Duration) error {
	if s.IsDestroyed() {
		return m.Destroy(ctx, s.ID())
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	st := storedSession{
		Values:    s.Values(),
		Flashes:   s.Flashes(),
		ExpiresAt: expiresAt,
	}

	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[s.ID()] = raw
	return nil
}

// Destroy deletes session data for the given identifier.
func (m *MemStore) Destroy(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, id)
	return nil
}

// Provider implements core.Provider for in-memory session store integration.
type Provider struct {
	cleanupInterval []time.Duration
	store           *MemStore
}

// NewProvider returns a new MemStore Provider.
func NewProvider(cleanupInterval ...time.Duration) *Provider {
	return &Provider{cleanupInterval: cleanupInterval}
}

// Register registers session.Store into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	c.Singleton(func() session.Store {
		p.store = New(p.cleanupInterval...)
		return p.store
	})
	return nil
}

// Boot verifies session store readiness.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var store session.Store
	return app.Container().Resolve(&store)
}

// Shutdown stops background cleaners and flushes session data.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.store != nil {
		return p.store.Close()
	}
	return nil
}

var _ core.Provider = (*Provider)(nil)
