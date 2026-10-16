package auth

import (
	"testing"
)

func TestRBACMultiRoleAndPermissions(t *testing.T) {
	rbac := NewRBACRegistry()

	// Define roles with wildcard and specific permissions
	rbac.DefineRole("editor", "posts:read", "posts:create", "posts:update")
	rbac.DefineRole("billing_admin", "invoices:*", "payments:*")
	rbac.DefineRole("superadmin", "*")

	// Principal with multi-roles: "editor" AND "billing_admin" + direct custom permission "beta:access"
	user := NewPrincipal("usr_100").
		WithRoles("editor", "billing_admin").
		WithPermissions("beta:access")

	// 1. Multi-role checks
	if !user.HasRole("editor") || !user.HasRole("billing_admin") {
		t.Errorf("expected user to have both editor and billing_admin roles")
	}
	if user.HasRole("superadmin") {
		t.Errorf("user should not have superadmin role")
	}
	if !user.HasAnyRole("superadmin", "editor") {
		t.Errorf("HasAnyRole failed for editor")
	}

	// 2. Direct custom permission check
	if !rbac.CheckPermission(user, "beta:access") {
		t.Errorf("expected direct permission beta:access to pass")
	}

	// 3. Inherited role permissions (exact and wildcard)
	if !rbac.CheckPermission(user, "posts:create") {
		t.Errorf("expected editor permission posts:create to pass")
	}
	if !rbac.CheckPermission(user, "invoices:generate:pdf") {
		t.Errorf("expected wildcard permission invoices:* to match invoices:generate:pdf")
	}
	if !rbac.CheckPermission(user, "payments:refund") {
		t.Errorf("expected wildcard permission payments:* to match payments:refund")
	}

	// 4. Permissions not assigned
	if rbac.CheckPermission(user, "system:reboot") {
		t.Errorf("unassigned permission system:reboot should fail")
	}

	// 5. CheckAllPermissions & CheckAnyPermission
	if !rbac.CheckAllPermissions(user, "posts:read", "invoices:view", "beta:access") {
		t.Errorf("CheckAllPermissions failed for valid permission subset")
	}
	if rbac.CheckAllPermissions(user, "posts:read", "system:reboot") {
		t.Errorf("CheckAllPermissions should fail when one permission is missing")
	}
	if !rbac.CheckAnyPermission(user, "nonexistent:perm", "beta:access") {
		t.Errorf("CheckAnyPermission failed when one valid permission is present")
	}
}

func TestSuperadminWildcard(t *testing.T) {
	rbac := NewRBACRegistry()
	rbac.DefineRole("superadmin", "*")

	admin := NewPrincipal("root").WithRoles("superadmin")

	if !rbac.CheckPermission(admin, "anything:goes:here") {
		t.Errorf("superadmin wildcard '*' should match all permission queries")
	}
}
