package todos

import (
	"testing"
)

func TestTodoFlow(t *testing.T) {
	service := NewTodoService(nil)
	if service == nil {
		t.Fatal("expected service instance, got nil")
	}
}
