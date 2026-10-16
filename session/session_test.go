package session

import (
	"testing"
)

func TestSessionValuesAndFlashes(t *testing.T) {
	s := New("sess_123")

	if s.ID() != "sess_123" {
		t.Errorf("s.ID() = %v, want sess_123", s.ID())
	}

	// Set & Get
	s.Set("user_id", 42)
	s.Set("role", "admin")

	uid, ok := s.Get("user_id")
	if !ok || uid != 42 {
		t.Errorf("s.Get(user_id) = %v, %v, want 42, true", uid, ok)
	}

	// Flash message lifecycle
	s.Flash("success", "Profile updated")
	if !s.IsModified() {
		t.Errorf("session should be marked modified")
	}

	// Flashes not available via Flashed until restored in next request cycle
	if _, exists := s.Flashed("success"); exists {
		t.Errorf("fresh flash should not be readable in same cycle via Flashed()")
	}

	// Simulate next request cycle:
	nextReqSession := New(s.ID())
	nextReqSession.Restore(s.Values(), s.Flashes())

	flashMsg, exists := nextReqSession.Flashed("success")
	if !exists || flashMsg != "Profile updated" {
		t.Errorf("next request cycle failed reading flash: %v, %v", flashMsg, exists)
	}

	// Regenerate ID
	oldID := nextReqSession.ID()
	newID := nextReqSession.Regenerate()
	if newID == oldID || nextReqSession.ID() != newID {
		t.Errorf("Regenerate() failed to update session ID: old=%v, new=%v", oldID, newID)
	}
}

func TestSessionDestroy(t *testing.T) {
	s := New("sess_del")
	s.Set("token", "secret")
	s.Destroy()

	if !s.IsDestroyed() {
		t.Errorf("session should be marked destroyed")
	}
	if _, ok := s.Get("token"); ok {
		t.Errorf("destroyed session should have empty values")
	}
}
