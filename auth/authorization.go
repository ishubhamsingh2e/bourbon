package auth

import (
	"context"
	"sync"
)

// PolicyFunc defines an authorization evaluation rule for a specific action and resource.
type PolicyFunc func(ctx context.Context, p Principal, resource any) bool

// Enforcer evaluates whether a Principal has permission to perform an action or hold a privilege.
type Enforcer interface {
	Can(ctx context.Context, p Principal, action string, resource any) (bool, error)
	HasPermission(ctx context.Context, p Principal, permission string) (bool, error)
}

// PolicyEnforcer evaluates authorization requests using registered policies and RBAC definitions.
type PolicyEnforcer struct {
	mu       sync.RWMutex
	rbac     *RBACRegistry
	policies map[string]PolicyFunc
}

// NewPolicyEnforcer creates an Enforcer backed by optional RBAC registries.
func NewPolicyEnforcer(rbac ...*RBACRegistry) *PolicyEnforcer {
	var reg *RBACRegistry
	if len(rbac) > 0 && rbac[0] != nil {
		reg = rbac[0]
	} else {
		reg = NewRBACRegistry()
	}

	return &PolicyEnforcer{
		rbac:     reg,
		policies: make(map[string]PolicyFunc),
	}
}

// Define registers an authorization policy for a specific action.
func (e *PolicyEnforcer) Define(action string, policy PolicyFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.policies[action] = policy
}

// Can checks if the principal is authorized to perform action on resource.
func (e *PolicyEnforcer) Can(ctx context.Context, p Principal, action string, resource any) (bool, error) {
	if p == nil {
		return false, nil
	}

	e.mu.RLock()
	policy, exists := e.policies[action]
	e.mu.RUnlock()

	if exists && policy != nil {
		return policy(ctx, p, resource), nil
	}

	// Fallback to RBAC permission matching
	return e.rbac.CheckPermission(p, action), nil
}

// HasPermission verifies if the principal holds the specified granular permission.
func (e *PolicyEnforcer) HasPermission(ctx context.Context, p Principal, permission string) (bool, error) {
	if p == nil {
		return false, nil
	}
	return e.rbac.CheckPermission(p, permission), nil
}
