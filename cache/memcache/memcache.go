// Package memcache provides an in-core, zero-dependency in-memory cache reference adapter.
package memcache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/cache"
	"github.com/ishubhamsingh2e/bourbon/core"
)

type cacheItem struct {
	data      []byte
	isNumeric bool
	numVal    int64
	expiresAt time.Time
}

// MemCache provides a thread-safe, in-memory cache implementation.
type MemCache struct {
	mu     sync.RWMutex
	items  map[string]cacheItem
	stopCh chan struct{}
}

// Config configures in-memory cache settings.
type Config struct {
	CleanupInterval time.Duration
}

// Open satisfies the cache.Driver interface.
func (c Config) Open() (cache.Cache, error) {
	interval := c.CleanupInterval
	if interval <= 0 {
		interval = time.Minute
	}
	return New(interval), nil
}

// New initializes an in-memory cache and starts an eviction worker if cleanupInterval > 0.
func New(cleanupInterval ...time.Duration) *MemCache {
	interval := time.Minute
	if len(cleanupInterval) > 0 && cleanupInterval[0] > 0 {
		interval = cleanupInterval[0]
	}

	mc := &MemCache{
		items:  make(map[string]cacheItem),
		stopCh: make(chan struct{}),
	}

	go mc.evictionLoop(interval)

	return mc
}

func (m *MemCache) evictionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case now := <-ticker.C:
			m.mu.Lock()
			for k, item := range m.items {
				if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
					delete(m.items, k)
				}
			}
			m.mu.Unlock()
		}
	}
}

// Close stops the background eviction worker and flushes cache items.
func (m *MemCache) Close() error {
	select {
	case <-m.stopCh:
	default:
		close(m.stopCh)
	}
	return m.Flush(context.Background())
}

// Get retrieves a key and decodes its JSON representation into dst.
func (m *MemCache) Get(ctx context.Context, key string, dst any) error {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || (!item.expiresAt.IsZero() && time.Now().After(item.expiresAt)) {
		return cache.ErrCacheMiss
	}

	if item.isNumeric {
		switch target := dst.(type) {
		case *int64:
			*target = item.numVal
			return nil
		case *int:
			*target = int(item.numVal)
			return nil
		case *string:
			*target = strconv.FormatInt(item.numVal, 10)
			return nil
		}
	}

	if err := json.Unmarshal(item.data, dst); err != nil {
		return fmt.Errorf("memcache: failed unmarshaling cache data: %w", err)
	}

	return nil
}

// Set serializes val into JSON and stores it with the specified TTL.
func (m *MemCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	// Check direct numeric types
	switch n := val.(type) {
	case int64:
		m.mu.Lock()
		m.items[key] = cacheItem{
			isNumeric: true,
			numVal:    n,
			expiresAt: expiresAt,
		}
		m.mu.Unlock()
		return nil
	case int:
		m.mu.Lock()
		m.items[key] = cacheItem{
			isNumeric: true,
			numVal:    int64(n),
			expiresAt: expiresAt,
		}
		m.mu.Unlock()
		return nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("memcache: failed marshaling value: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = cacheItem{
		data:      data,
		expiresAt: expiresAt,
	}

	return nil
}

// Delete removes a key from cache.
func (m *MemCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
	return nil
}

// Has checks if a key exists and is non-expired.
func (m *MemCache) Has(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || (!item.expiresAt.IsZero() && time.Now().After(item.expiresAt)) {
		return false, nil
	}
	return true, nil
}

// Flush removes all keys from cache.
func (m *MemCache) Flush(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string]cacheItem)
	return nil
}

// Remember retrieves key if present, or executes fn, stores result, and populates dst.
func (m *MemCache) Remember(ctx context.Context, key string, ttl time.Duration, dst any, fn func() (any, error)) error {
	if err := m.Get(ctx, key, dst); err == nil {
		return nil
	}

	val, err := fn()
	if err != nil {
		return err
	}

	if err := m.Set(ctx, key, val, ttl); err != nil {
		return err
	}

	return m.Get(ctx, key, dst)
}

// Increment atomically increases an integer key value.
func (m *MemCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.items[key]
	var current int64
	var expiresAt time.Time

	if exists && (item.expiresAt.IsZero() || time.Now().Before(item.expiresAt)) {
		current = item.numVal
		expiresAt = item.expiresAt
	}

	newVal := current + value
	m.items[key] = cacheItem{
		isNumeric: true,
		numVal:    newVal,
		expiresAt: expiresAt,
	}

	return newVal, nil
}

// Decrement atomically decreases an integer key value.
func (m *MemCache) Decrement(ctx context.Context, key string, value int64) (int64, error) {
	return m.Increment(ctx, key, -value)
}

// Provider implements core.Provider for in-memory cache integration.
type Provider struct {
	cfg   Config
	cache *MemCache
}

// NewProvider returns a new MemCache Provider with the given configuration.
func NewProvider(cfg Config) *Provider {
	return &Provider{cfg: cfg}
}

// Register registers cache.Cache into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	c.Singleton(func() (cache.Cache, error) {
		mc, err := p.cfg.Open()
		if err != nil {
			return nil, err
		}
		if typed, ok := mc.(*MemCache); ok {
			p.cache = typed
		}
		return mc, nil
	})
	return nil
}

// Boot verifies cache readiness by resolving it from the container.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var ca cache.Cache
	return app.Container().Resolve(&ca)
}

// Shutdown stops background eviction and flushes the in-memory cache.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.cache != nil {
		return p.cache.Close()
	}
	return nil
}

var _ core.Provider = (*Provider)(nil)
