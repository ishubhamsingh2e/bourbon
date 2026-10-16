package validation

import (
	"context"
	"testing"
)

type SignUpRequest struct {
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

func TestSchemaValidatorRules(t *testing.T) {
	ctx := context.Background()

	validator := New().
		Field("email", Required(), Email()).
		Field("age", Required(), Min(18), Max(120)).
		Field("role", Required(), In("admin", "member", "guest")).
		Field("password", Required(), MinLength(8), MaxLength(32))

	// 1. Valid payload
	validReq := SignUpRequest{
		Email:    "alice@example.com",
		Age:      25,
		Role:     "member",
		Password: "SuperSecretPassword123!",
	}

	if err := validator.Validate(validReq); err != nil {
		t.Fatalf("expected valid payload to pass, got: %v", err)
	}

	// 2. Invalid payload
	invalidReq := SignUpRequest{
		Email:    "not-an-email",
		Age:      15,
		Role:     "superuser",
		Password: "short",
	}

	err := validator.Validate(invalidReq)
	if err == nil {
		t.Fatalf("expected validation errors, got nil")
	}

	valErrs, ok := err.(Errors)
	if !ok {
		t.Fatalf("expected err to be of type validation.Errors, got %T", err)
	}

	if !valErrs.Has("email") {
		t.Errorf("expected error on email")
	}
	if !valErrs.Has("age") {
		t.Errorf("expected error on age")
	}
	if !valErrs.Has("role") {
		t.Errorf("expected error on role")
	}
	if !valErrs.Has("password") {
		t.Errorf("expected error on password")
	}

	// 3. Map validation
	mapData := map[string]any{
		"email":    "bob@test.com",
		"age":      30,
		"role":     "admin",
		"password": "valid_password_length",
	}
	if err := validator.ValidateMap(ctx, mapData); err != nil {
		t.Errorf("ValidateMap failed for valid map: %v", err)
	}
}
