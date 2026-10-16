package contracts

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
)

// VerifyTokenProviderContract executes a standard compliance test suite against any auth.TokenProvider implementation.
func VerifyTokenProviderContract(t *testing.T, provider auth.TokenProvider) {
	principal := auth.NewPrincipal("test_contract_user").
		WithRoles("admin", "billing").
		WithPermissions("invoices:read").
		WithClaims(map[string]any{"tenant": "acme"})

	t.Run("Generate_and_Validate", func(t *testing.T) {
		token, err := provider.Generate(principal, 5*time.Minute)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		if token == "" {
			t.Fatalf("expected non-empty token string")
		}

		p, err := provider.Validate(token)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}
		if p.ID() != "test_contract_user" {
			t.Errorf("p.ID() = %v, want test_contract_user", p.ID())
		}
		if !p.HasRole("admin") || !p.HasRole("billing") {
			t.Errorf("roles mismatch: %v", p.Roles())
		}
		if !p.HasPermission("invoices:read") {
			t.Errorf("permissions mismatch: %v", p.Permissions())
		}
	})

	t.Run("Tampered_Token_Rejection", func(t *testing.T) {
		token, _ := provider.Generate(principal, 5*time.Minute)
		_, err := provider.Validate(token + "_corrupted")
		if !errors.Is(err, auth.ErrInvalidToken) && err == nil {
			t.Errorf("expected ErrInvalidToken on corrupted token, got %v", err)
		}
	})
}

// VerifyGuardContract executes a standard compliance test suite against any auth.Guard implementation.
func VerifyGuardContract(t *testing.T, guard auth.Guard, validReq, invalidReq *http.Request) {
	ctx := context.Background()

	t.Run("Valid_Request_Authentication", func(t *testing.T) {
		p, err := guard.Authenticate(ctx, validReq)
		if err != nil {
			t.Fatalf("Authenticate valid request failed: %v", err)
		}
		if p == nil || p.ID() == "" {
			t.Errorf("expected non-empty principal ID")
		}
	})

	t.Run("Invalid_Request_Rejection", func(t *testing.T) {
		_, err := guard.Authenticate(ctx, invalidReq)
		if err == nil {
			t.Errorf("expected error on invalid request, got nil")
		}
	})
}
