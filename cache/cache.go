// Package cache defines key-value caching abstractions and atomic operation contracts.
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned when a requested key is not present in the cache or has expired.
var ErrCacheMiss = errors.New("cache: key not found")

// Cache defines the standard key-value caching interface.
type Cache interface {
	Get(ctx context.Context, key string, dst any) error
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	Flush(ctx context.Context) error
	Remember(ctx context.Context, key string, ttl time.Duration, dst any, fn func() (any, error)) error
	Increment(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string, value int64) (int64, error)
}

// Driver represents a cache provider capable of opening a Cache instance.
type Driver interface {
	Open() (Cache, error)
}
