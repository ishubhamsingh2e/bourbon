package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProblemDetailSerialization(t *testing.T) {
	prob := NewProblemDetail(http.StatusNotFound, "Resource Not Found", "The requested user ID does not exist.")

	data, err := json.Marshal(prob)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	str := string(data)
	if !strings.Contains(str, `"title":"Resource Not Found"`) || !strings.Contains(str, `"status":404`) {
		t.Errorf("unexpected serialized problem detail: %v", str)
	}

	if prob.Error() != "Resource Not Found (404): The requested user ID does not exist." {
		t.Errorf("prob.Error() = %v", prob.Error())
	}
}

func TestValidationErrorSerialization(t *testing.T) {
	params := []InvalidParam{
		{Field: "email", Tag: "required", Message: "Email is required"},
		{Field: "age", Tag: "gte", Message: "Age must be at least 18"},
	}

	valErr := NewValidationError("/users", params)
	data, err := json.Marshal(valErr)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	str := string(data)
	if !strings.Contains(str, `"status":422`) || !strings.Contains(str, `"field":"email"`) || !strings.Contains(str, `"instance":"/users"`) {
		t.Errorf("unexpected validation error JSON: %v", str)
	}
}
