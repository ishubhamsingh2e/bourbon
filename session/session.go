package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type sessionContextKey struct{}

var sessionKey = sessionContextKey{}

// WithSession returns a copy of parent context with the session attached.
func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

// FromContext extracts the active session from context.
func FromContext(ctx context.Context) *Session {
	if ctx == nil {
		return nil
	}
	s, _ := ctx.Value(sessionKey).(*Session)
	return s
}

// Session represents stateful client session data including temporary flash attributes.
type Session struct {
	mu        sync.RWMutex
	id        string
	values    map[string]any
	flashes   map[string]any
	oldFlash  map[string]any
	modified  bool
	destroyed bool
}

// GenerateID produces a crypto-random hex session identifier.
func GenerateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}

// New initializes an active Session with the given ID (or generates one if empty).
func New(id string) *Session {
	if id == "" {
		id = GenerateID()
	}
	return &Session{
		id:       id,
		values:   make(map[string]any),
		flashes:  make(map[string]any),
		oldFlash: make(map[string]any),
	}
}

// ID returns the current session identifier.
func (s *Session) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

// Get retrieves a value stored in this session.
func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.values[key]
	return val, ok
}

// GetString retrieves a string value from session or empty string.
func (s *Session) GetString(key string) string {
	if val, ok := s.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt retrieves an integer value from session, handling float64 JSON numbers.
func (s *Session) GetInt(key string) (int, bool) {
	if val, ok := s.Get(key); ok {
		switch v := val.(type) {
		case int:
			return v, true
		case int64:
			return int(v), true
		case float64:
			return int(v), true
		}
	}
	return 0, false
}

// GetBool retrieves a boolean value from session.
func (s *Session) GetBool(key string) bool {
	if val, ok := s.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Set stores a key-value pair in this session and marks it modified.
func (s *Session) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = val
	s.modified = true
}

// Delete removes a key from this session and marks it modified.
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	s.modified = true
}

// Flash registers a temporary key-value pair to be read on the next request.
func (s *Session) Flash(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flashes[key] = val
	s.modified = true
}

// Flashed retrieves a flash attribute stored from the preceding request.
func (s *Session) Flashed(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.oldFlash[key]
	return val, ok
}

// Destroy marks the session for immediate deletion and clears data.
func (s *Session) Destroy() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.destroyed = true
	s.values = make(map[string]any)
	s.flashes = make(map[string]any)
	s.oldFlash = make(map[string]any)
}

// Regenerate issues a new unique session identifier while preserving values.
func (s *Session) Regenerate() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id = GenerateID()
	s.modified = true
	return s.id
}

// IsModified returns whether session state was mutated during the request.
func (s *Session) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

// IsDestroyed returns whether session was destroyed.
func (s *Session) IsDestroyed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.destroyed
}

// Values returns a copy of internal session values.
func (s *Session) Values() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]any, len(s.values))
	for k, v := range s.values {
		cp[k] = v
	}
	return cp
}

// Flashes returns a copy of active flash data to be persisted for next request.
func (s *Session) Flashes() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make(map[string]any, len(s.flashes))
	for k, v := range s.flashes {
		cp[k] = v
	}
	return cp
}

// Restore hydrates session data and advances flash messages.
func (s *Session) Restore(values map[string]any, flashes map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = values
	if s.values == nil {
		s.values = make(map[string]any)
	}
	// Flashes from previous request become oldFlash for consumption in this request
	s.oldFlash = flashes
	if s.oldFlash == nil {
		s.oldFlash = make(map[string]any)
	}
	s.flashes = make(map[string]any)
	s.modified = false
}
