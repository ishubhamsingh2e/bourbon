package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// TokenProvider defines the contract for generating and validating stateless bearer tokens.
type TokenProvider interface {
	Generate(p Principal, ttl time.Duration) (string, error)
	Validate(token string) (Principal, error)
}

type tokenPayload struct {
	Subject     string         `json:"sub"`
	Roles       []string       `json:"roles,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
	Claims      map[string]any `json:"claims,omitempty"`
	IssuedAt    int64          `json:"iat"`
	ExpiresAt   int64          `json:"exp"`
}

// HMACTokenProvider issues and verifies tamper-proof HMAC-SHA256 signed stateless tokens.
type HMACTokenProvider struct {
	secret []byte
}

// NewHMACTokenProvider creates an initialized HMAC token provider with the given signing secret.
func NewHMACTokenProvider(secret []byte) *HMACTokenProvider {
	if len(secret) == 0 {
		secret = []byte("bourbon_default_insecure_secret_key_123456")
	}
	return &HMACTokenProvider{secret: secret}
}

// Generate issues a signed bearer token valid for the specified duration.
func (h *HMACTokenProvider) Generate(p Principal, ttl time.Duration) (string, error) {
	if p == nil {
		return "", fmt.Errorf("auth: principal cannot be nil")
	}

	now := time.Now()
	payload := tokenPayload{
		Subject:     p.ID(),
		Roles:       p.Roles(),
		Permissions: p.Permissions(),
		Claims:      p.Claims(),
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(ttl).Unix(),
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("auth: failed marshaling token payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(rawPayload)
	signature := h.sign(encodedPayload)

	return encodedPayload + "." + signature, nil
}

// Validate verifies the token signature and expiration, returning the authenticated Principal.
func (h *HMACTokenProvider) Validate(token string) (Principal, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	encodedPayload := parts[0]
	signature := parts[1]

	expectedSignature := h.sign(encodedPayload)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, ErrInvalidToken
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var payload tokenPayload
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return nil, ErrInvalidToken
	}

	principal := NewPrincipal(payload.Subject).
		WithRoles(payload.Roles...).
		WithPermissions(payload.Permissions...).
		WithClaims(payload.Claims)

	return principal, nil
}

func (h *HMACTokenProvider) sign(data string) string {
	mac := hmac.New(sha256.New, h.secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
