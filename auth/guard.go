package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var (
	// ErrUnauthenticated is returned when no valid credentials or session are found.
	ErrUnauthenticated = errors.New("auth: unauthenticated")
	// ErrUnauthorized is returned when an authenticated identity lacks required permissions.
	ErrUnauthorized = errors.New("auth: unauthorized")
	// ErrGuardNotFound is returned when a requested authentication guard is not registered.
	ErrGuardNotFound = errors.New("auth: guard not registered")
	// ErrInvalidToken is returned when a bearer token is expired, corrupted, or signature-invalid.
	ErrInvalidToken = errors.New("auth: invalid or expired token")
)

// Guard inspects an incoming HTTP request to authenticate a Principal.
type Guard interface {
	Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}

// Manager orchestrates multiple registered authentication guards.
type Manager struct {
	mu           sync.RWMutex
	defaultGuard string
	guards       map[string]Guard
}

// NewManager constructs a multi-guard manager with an optional default guard name.
func NewManager(defaultGuard string) *Manager {
	return &Manager{
		defaultGuard: defaultGuard,
		guards:       make(map[string]Guard),
	}
}

// Register adds a named guard to the manager.
func (m *Manager) Register(name string, guard Guard) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.guards[name] = guard
}

// Guard returns the requested guard by name, or the default guard if no name is passed.
func (m *Manager) Guard(name ...string) (Guard, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	targetName := m.defaultGuard
	if len(name) > 0 && name[0] != "" {
		targetName = name[0]
	}

	if targetName == "" {
		return nil, ErrGuardNotFound
	}

	g, exists := m.guards[targetName]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrGuardNotFound, targetName)
	}
	return g, nil
}

// Authenticate attempts to authenticate a request against specified guards (or default guard).
func (m *Manager) Authenticate(ctx context.Context, r *http.Request, guardNames ...string) (Principal, error) {
	targets := guardNames
	if len(targets) == 0 && m.defaultGuard != "" {
		targets = []string{m.defaultGuard}
	}

	if len(targets) == 0 {
		return nil, ErrGuardNotFound
	}

	for _, name := range targets {
		guard, err := m.Guard(name)
		if err != nil {
			continue
		}

		principal, err := guard.Authenticate(ctx, r)
		if err == nil && principal != nil {
			return principal, nil
		}
	}

	return nil, ErrUnauthenticated
}
