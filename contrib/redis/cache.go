package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ishubhamsingh2e/bourbon/cache"
)

// Cache implements cache.Cache on top of Redis.
type Cache struct {
	client Client
	prefix string
}

// NewCache constructs a Redis cache adapter.
func NewCache(client Client, prefix ...string) *Cache {
	p := "bourbon:cache:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &Cache{client: client, prefix: p}
}

func (c *Cache) key(k string) string {
	return c.prefix + k
}

// Get retrieves a key from Redis.
func (c *Cache) Get(ctx context.Context, key string, dst any) error {
	data, ok, err := c.client.Get(ctx, c.key(key))
	if err != nil {
		return err
	}
	if !ok {
		return cache.ErrCacheMiss
	}
	return json.Unmarshal(data, dst)
}

// Set stores a key in Redis with TTL.
func (c *Cache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(key), data, ttl)
}

// Delete removes a key from Redis.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Delete(ctx, c.key(key))
}

// Has checks key presence.
func (c *Cache) Has(ctx context.Context, key string) (bool, error) {
	_, ok, err := c.client.Get(ctx, c.key(key))
	return ok, err
}

// Flush clears cached keys.
func (c *Cache) Flush(ctx context.Context) error {
	return c.client.FlushAll(ctx)
}

// Remember retrieves from cache or computes via fallback.
func (c *Cache) Remember(ctx context.Context, key string, ttl time.Duration, dst any, fn func() (any, error)) error {
	err := c.Get(ctx, key, dst)
	if err == nil {
		return nil
	}

	fresh, err := fn()
	if err != nil {
		return err
	}

	if err := c.Set(ctx, key, fresh, ttl); err != nil {
		return err
	}

	data, _ := json.Marshal(fresh)
	return json.Unmarshal(data, dst)
}

// Increment adds delta.
func (c *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.key(key), delta)
}

// Decrement subtracts delta.
func (c *Cache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.client.IncrBy(ctx, c.key(key), -delta)
}

var _ cache.Cache = (*Cache)(nil)
