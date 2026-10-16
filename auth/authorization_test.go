package auth

import (
	"context"
	"testing"
)

type Document struct {
	OwnerID string
	Public  bool
}

func TestPolicyEnforcer(t *testing.T) {
	ctx := context.Background()
	rbac := NewRBACRegistry()
	rbac.DefineRole("moderator", "documents:delete")

	enforcer := NewPolicyEnforcer(rbac)

	// Define dynamic policy
	enforcer.Define("documents:edit", func(ctx context.Context, p Principal, resource any) bool {
		doc, ok := resource.(Document)
		if !ok {
			return false
		}
		return doc.OwnerID == p.ID()
	})

	owner := NewPrincipal("usr_1")
	other := NewPrincipal("usr_2").WithRoles("moderator")

	myDoc := Document{OwnerID: "usr_1", Public: false}

	// 1. Dynamic policy check
	canEdit, _ := enforcer.Can(ctx, owner, "documents:edit", myDoc)
	if !canEdit {
		t.Errorf("owner should be allowed to edit document")
	}

	canOtherEdit, _ := enforcer.Can(ctx, other, "documents:edit", myDoc)
	if canOtherEdit {
		t.Errorf("non-owner should not be allowed to edit document")
	}

	// 2. RBAC fallback check
	canDelete, _ := enforcer.Can(ctx, other, "documents:delete", myDoc)
	if !canDelete {
		t.Errorf("moderator should be allowed to delete document via RBAC fallback")
	}

	canOwnerDelete, _ := enforcer.Can(ctx, owner, "documents:delete", myDoc)
	if canOwnerDelete {
		t.Errorf("owner without moderator role should not be allowed to delete")
	}
}
