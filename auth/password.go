package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

// PasswordHasher defines password hashing and verification contracts.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hash string) (bool, error)
}

// ReferencePasswordHasher provides a zero-dependency salted SHA-256 password hasher.
type ReferencePasswordHasher struct {
	iterations int
}

// NewPasswordHasher creates a new reference password hasher.
func NewPasswordHasher() *ReferencePasswordHasher {
	return &ReferencePasswordHasher{iterations: 10000}
}

// Hash generates a salted cryptographic digest for the plain text password.
func (h *ReferencePasswordHasher) Hash(plain string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: failed generating salt: %w", err)
	}

	digest := h.computeDigest(plain, salt)
	return fmt.Sprintf("v1$%s$%s", hex.EncodeToString(salt), hex.EncodeToString(digest)), nil
}

// Verify checks whether a plain text password matches the stored hash in constant time.
func (h *ReferencePasswordHasher) Verify(plain, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 3 || parts[0] != "v1" {
		return false, fmt.Errorf("auth: invalid password hash format")
	}

	salt, err := hex.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("auth: invalid salt encoding: %w", err)
	}

	expectedDigest, err := hex.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("auth: invalid digest encoding: %w", err)
	}

	actualDigest := h.computeDigest(plain, salt)
	if subtle.ConstantTimeCompare(actualDigest, expectedDigest) == 1 {
		return true, nil
	}

	return false, nil
}

func (h *ReferencePasswordHasher) computeDigest(plain string, salt []byte) []byte {
	hasher := sha256.New()
	hasher.Write(salt)
	hasher.Write([]byte(plain))
	digest := hasher.Sum(nil)

	for i := 1; i < h.iterations; i++ {
		hasher.Reset()
		hasher.Write(digest)
		digest = hasher.Sum(nil)
	}

	return digest
}
