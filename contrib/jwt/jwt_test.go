package jwt

import (
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
)

func TestJWTProviderOperations(t *testing.T) {
	provider := NewTokenProvider([]byte("jwt_super_secret_key_1234567890"))

	principal := auth.NewPrincipal("usr_99").
		WithRoles("admin", "moderator").
		WithPermissions("users.create", "users.delete").
		WithClaims(map[string]any{"email": "admin@example.com"})

	token, err := provider.Generate(principal, time.Hour)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	validated, err := provider.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validated.ID() != "usr_99" {
		t.Errorf("ID mismatch: %s", validated.ID())
	}
	if !validated.HasRole("admin") {
		t.Errorf("missing admin role")
	}
	if !validated.HasPermission("users.delete") {
		t.Errorf("missing users.delete permission")
	}

	// Tampered token test
	tampered := token + "evil"
	_, err = provider.Validate(tampered)
	if err == nil {
		t.Errorf("expected error on tampered token")
	}
}
