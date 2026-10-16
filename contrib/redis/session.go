package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ishubhamsingh2e/bourbon/session"
)

type storedSession struct {
	Values    map[string]any `json:"values"`
	Flashes   map[string]any `json:"flashes"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// SessionStore implements session.Store backed by Redis.
type SessionStore struct {
	client Client
	prefix string
}

// NewSessionStore constructs a Redis session store adapter.
func NewSessionStore(client Client, prefix ...string) *SessionStore {
	p := "bourbon:session:"
	if len(prefix) > 0 && prefix[0] != "" {
		p = prefix[0]
	}
	return &SessionStore{client: client, prefix: p}
}

func (s *SessionStore) key(id string) string {
	return s.prefix + id
}

// Save stores session payload in Redis.
func (s *SessionStore) Save(ctx context.Context, sess *session.Session, ttl time.Duration) error {
	if sess.IsDestroyed() {
		return s.Destroy(ctx, sess.ID())
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	st := storedSession{
		Values:    sess.Values(),
		Flashes:   sess.Flashes(),
		ExpiresAt: expiresAt,
	}

	bytes, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.key(sess.ID()), bytes, ttl)
}

// Load retrieves session payload from Redis.
func (s *SessionStore) Load(ctx context.Context, id string) (*session.Session, error) {
	if id == "" {
		return session.New(""), nil
	}

	bytes, ok, err := s.client.Get(ctx, s.key(id))
	if err != nil || !ok {
		return session.New(id), nil
	}

	var st storedSession
	if err := json.Unmarshal(bytes, &st); err != nil {
		return session.New(id), nil
	}

	if !st.ExpiresAt.IsZero() && time.Now().After(st.ExpiresAt) {
		_ = s.Destroy(ctx, id)
		return session.New(id), nil
	}

	sess := session.New(id)
	sess.Restore(st.Values, st.Flashes)
	return sess, nil
}

// Destroy deletes the session from Redis.
func (s *SessionStore) Destroy(ctx context.Context, id string) error {
	return s.client.Delete(ctx, s.key(id))
}

var _ session.Store = (*SessionStore)(nil)
