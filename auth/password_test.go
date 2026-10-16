package auth

import (
	"testing"
)

func TestReferencePasswordHasher(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "CorrectHorseBatteryStaple!"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	valid, err := hasher.Verify(password, hash)
	if err != nil || !valid {
		t.Fatalf("Verify failed for valid password: %v, valid=%v", err, valid)
	}

	invalid, err := hasher.Verify("WrongPassword123", hash)
	if err != nil || invalid {
		t.Fatalf("Verify should return false for wrong password: %v, invalid=%v", err, invalid)
	}
}
