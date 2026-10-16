package auth

import (
	"strings"
	"sync"
)

// MatchPermission evaluates if a granted permission pattern matches a requested permission target.
// Supports full wildcard ("*") and namespace wildcards ("users:*", "reports:export:*").
func MatchPermission(pattern, target string) bool {
	if pattern == "*" || pattern == target {
		return true
	}
	if strings.HasSuffix(pattern, ":*") {
		prefix := strings.TrimSuffix(pattern, ":*")
		if target == prefix || strings.HasPrefix(target, prefix+":") {
			return true
		}
	}
	return false
}

// RBACRegistry manages role definitions, role-to-permission mappings, and wildcard permission matching.
type RBACRegistry struct {
	mu    sync.RWMutex
	roles map[string]map[string]struct{}
}

// NewRBACRegistry creates an initialized RBAC registry.
func NewRBACRegistry() *RBACRegistry {
	return &RBACRegistry{
		roles: make(map[string]map[string]struct{}),
	}
}

// DefineRole registers a role and its associated permissions.
func (r *RBACRegistry) DefineRole(role string, permissions ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[role]; !exists {
		r.roles[role] = make(map[string]struct{})
	}
	for _, p := range permissions {
		if p != "" {
			r.roles[role][p] = struct{}{}
		}
	}
}

// DefineRoles registers a batch of role definitions.
func (r *RBACRegistry) DefineRoles(definitions map[string][]string) {
	for role, perms := range definitions {
		r.DefineRole(role, perms...)
	}
}

// PermissionsForRole returns all permissions explicitly assigned to a single role.
func (r *RBACRegistry) PermissionsForRole(role string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	permsMap, exists := r.roles[role]
	if !exists {
		return nil
	}

	result := make([]string, 0, len(permsMap))
	for p := range permsMap {
		result = append(result, p)
	}
	return result
}

// PermissionsForRoles returns deduplicated permissions aggregated across multiple roles.
func (r *RBACRegistry) PermissionsForRoles(roles ...string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dedup := make(map[string]struct{})
	for _, role := range roles {
		if permsMap, exists := r.roles[role]; exists {
			for p := range permsMap {
				dedup[p] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(dedup))
	for p := range dedup {
		result = append(result, p)
	}
	return result
}

// CheckPermission verifies whether a Principal has the given permission (either directly or via roles).
func (r *RBACRegistry) CheckPermission(p Principal, permission string) bool {
	if p == nil {
		return false
	}

	// 1. Direct custom permission on principal
	if p.HasPermission(permission) {
		return true
	}

	// 2. Inherited permissions from any of the principal's assigned roles
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, role := range p.Roles() {
		if permsMap, exists := r.roles[role]; exists {
			for pattern := range permsMap {
				if MatchPermission(pattern, permission) {
					return true
				}
			}
		}
	}

	return false
}

// CheckAllPermissions verifies that a Principal has all specified permissions.
func (r *RBACRegistry) CheckAllPermissions(p Principal, permissions ...string) bool {
	for _, perm := range permissions {
		if !r.CheckPermission(p, perm) {
			return false
		}
	}
	return true
}

// CheckAnyPermission verifies that a Principal has at least one of the specified permissions.
func (r *RBACRegistry) CheckAnyPermission(p Principal, permissions ...string) bool {
	for _, perm := range permissions {
		if r.CheckPermission(p, perm) {
			return true
		}
	}
	return false
}
