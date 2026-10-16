package auth

import (
	"context"
	"sync"
)

type authContextKey struct{}

var principalKey = authContextKey{}

// WithPrincipal returns a copy of parent context with the principal attached.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFromContext extracts the authenticated principal from context.
func PrincipalFromContext(ctx context.Context) Principal {
	if ctx == nil {
		return nil
	}
	p, _ := ctx.Value(principalKey).(Principal)
	return p
}

// Principal represents an authenticated identity with multi-role and custom permission support.
type Principal interface {
	ID() string
	Claims() map[string]any
	Roles() []string
	HasRole(role string) bool
	HasAnyRole(roles ...string) bool
	Permissions() []string
	HasPermission(perm string) bool
	HasAnyPermission(perms ...string) bool
	HasAllPermissions(perms ...string) bool
}

// UserPrincipal is a concrete reference implementation of Principal.
type UserPrincipal struct {
	mu       sync.RWMutex
	Subject  string         `json:"sub"`
	RoleList []string       `json:"roles"`
	PermList []string       `json:"permissions"`
	Data     map[string]any `json:"data"`
}

// NewPrincipal constructs a new UserPrincipal with the given subject identifier.
func NewPrincipal(subject string) *UserPrincipal {
	return &UserPrincipal{
		Subject:  subject,
		RoleList: make([]string, 0),
		PermList: make([]string, 0),
		Data:     make(map[string]any),
	}
}

// WithRoles assigns multiple roles to the principal.
func (u *UserPrincipal) WithRoles(roles ...string) *UserPrincipal {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, r := range roles {
		if r != "" {
			u.RoleList = append(u.RoleList, r)
		}
	}
	return u
}

// WithPermissions grants direct custom permissions to the principal.
func (u *UserPrincipal) WithPermissions(permissions ...string) *UserPrincipal {
	u.mu.Lock()
	defer u.mu.Unlock()
	for _, p := range permissions {
		if p != "" {
			u.PermList = append(u.PermList, p)
		}
	}
	return u
}

// WithClaims attaches custom claim key-value pairs to the principal.
func (u *UserPrincipal) WithClaims(data map[string]any) *UserPrincipal {
	u.mu.Lock()
	defer u.mu.Unlock()
	for k, v := range data {
		u.Data[k] = v
	}
	return u
}

// ID returns the unique subject identifier.
func (u *UserPrincipal) ID() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Subject
}

// Claims returns a copy of custom claims.
func (u *UserPrincipal) Claims() map[string]any {
	u.mu.RLock()
	defer u.mu.RUnlock()
	cp := make(map[string]any, len(u.Data))
	for k, v := range u.Data {
		cp[k] = v
	}
	return cp
}

// Roles returns a copy of assigned roles.
func (u *UserPrincipal) Roles() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	cp := make([]string, len(u.RoleList))
	copy(cp, u.RoleList)
	return cp
}

// HasRole checks if the principal has the specified role.
func (u *UserPrincipal) HasRole(role string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, r := range u.RoleList {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the principal has at least one of the specified roles.
func (u *UserPrincipal) HasAnyRole(roles ...string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, r := range u.RoleList {
		for _, target := range roles {
			if r == target {
				return true
			}
		}
	}
	return false
}

// Permissions returns a copy of direct custom permissions.
func (u *UserPrincipal) Permissions() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	cp := make([]string, len(u.PermList))
	copy(cp, u.PermList)
	return cp
}

// HasPermission checks if the principal has the direct custom permission.
func (u *UserPrincipal) HasPermission(perm string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, p := range u.PermList {
		if MatchPermission(p, perm) {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the principal has at least one of the specified permissions.
func (u *UserPrincipal) HasAnyPermission(perms ...string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	for _, target := range perms {
		for _, p := range u.PermList {
			if MatchPermission(p, target) {
				return true
			}
		}
	}
	return false
}

// HasAllPermissions checks if the principal has all specified permissions.
func (u *UserPrincipal) HasAllPermissions(perms ...string) bool {
	for _, target := range perms {
		if !u.HasPermission(target) {
			return false
		}
	}
	return true
}
