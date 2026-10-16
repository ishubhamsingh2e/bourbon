package session

import (
	"context"
	"time"
)

// Store defines persistence operations for saving and loading stateful sessions.
type Store interface {
	Load(ctx context.Context, id string) (*Session, error)
	Save(ctx context.Context, s *Session, ttl time.Duration) error
	Destroy(ctx context.Context, id string) error
}
