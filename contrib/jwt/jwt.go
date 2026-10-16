// Package jwt provides a standard JWT token provider adapter for auth.TokenProvider.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ishubhamsingh2e/bourbon/auth"
)

// Claims represents standard and custom JWT claims.
type Claims struct {
	Sub         string         `json:"sub"`
	Exp         int64          `json:"exp"`
	Iat         int64          `json:"iat"`
	Roles       []string       `json:"roles,omitempty"`
	Permissions []string       `json:"permissions,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

// TokenProvider implements auth.TokenProvider using HMAC-SHA256 JWT tokens.
type TokenProvider struct {
	secret []byte
}

// NewTokenProvider constructs a JWT TokenProvider with the given signing secret.
func NewTokenProvider(secret []byte) *TokenProvider {
	return &TokenProvider{secret: secret}
}

// Generate creates a signed JWT string for the principal.
func (p *TokenProvider) Generate(principal auth.Principal, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Sub:         principal.ID(),
		Iat:         now.Unix(),
		Exp:         now.Add(ttl).Unix(),
		Roles:       principal.Roles(),
		Permissions: principal.Permissions(),
		Data:        principal.Claims(),
	}

	headerJSON := `{"alg":"HS256","typ":"JWT"}`
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("jwt: marshal claims failed: %w", err)
	}
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	mac := hmac.New(sha256.New, p.secret)
	mac.Write([]byte(signingInput))
	sigB64 := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sigB64, nil
}

// Validate verifies signature and expiry, returning the authenticated principal.
func (p *TokenProvider) Validate(tokenString string) (auth.Principal, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("jwt: invalid token format")
	}

	headerB64, claimsB64, sigB64 := parts[0], parts[1], parts[2]
	signingInput := headerB64 + "." + claimsB64

	mac := hmac.New(sha256.New, p.secret)
	mac.Write([]byte(signingInput))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil || !hmac.Equal(expectedSig, actualSig) {
		return nil, errors.New("jwt: invalid token signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsB64)
	if err != nil {
		return nil, errors.New("jwt: invalid claims encoding")
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, errors.New("jwt: malformed claims json")
	}

	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, errors.New("jwt: token expired")
	}

	principal := auth.NewPrincipal(claims.Sub).
		WithRoles(claims.Roles...).
		WithPermissions(claims.Permissions...).
		WithClaims(claims.Data)

	return principal, nil
}

var _ auth.TokenProvider = (*TokenProvider)(nil)
