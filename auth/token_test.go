package auth

import (
	"errors"
	"testing"
	"time"
)

func TestHMACTokenProvider(t *testing.T) {
	secret := []byte("test_secret_key_9876543210")
	provider := NewHMACTokenProvider(secret)

	principal := NewPrincipal("usr_42").
		WithRoles("admin", "editor").
		WithPermissions("articles:write").
		WithClaims(map[string]any{"org_id": "org_99"})

	// 1. Generate and Validate
	token, err := provider.Generate(principal, time.Hour)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	validated, err := provider.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if validated.ID() != "usr_42" {
		t.Errorf("validated ID = %v, want usr_42", validated.ID())
	}
	if !validated.HasRole("admin") || !validated.HasRole("editor") {
		t.Errorf("validated roles mismatch: %v", validated.Roles())
	}
	if !validated.HasPermission("articles:write") {
		t.Errorf("validated permissions mismatch: %v", validated.Permissions())
	}
	if validated.Claims()["org_id"] != "org_99" {
		t.Errorf("validated claims mismatch: %v", validated.Claims())
	}

	// 2. Tampered Token
	tamperedToken := token + "tampered"
	_, err = provider.Validate(tamperedToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken on tampered token, got %v", err)
	}

	// 3. Expired Token
	shortProvider := NewHMACTokenProvider(secret)
	expiredToken, _ := shortProvider.Generate(principal, -time.Second)
	_, err = shortProvider.Validate(expiredToken)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken on expired token, got %v", err)
	}
}
