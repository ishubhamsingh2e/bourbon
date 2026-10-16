package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// TestResponse wraps the raw HTTP test response with fluent assertion helpers.
type TestResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	t          *testing.T
}

// AssertStatus verifies that the response status code equals expected.
func (r *TestResponse) AssertStatus(expected int) *TestResponse {
	if r.StatusCode != expected {
		r.fail(fmt.Sprintf("status code: expected %d, got %d. Body: %s", expected, r.StatusCode, string(r.Body)))
	}
	return r
}

// AssertHeader verifies that the given HTTP header matches expected.
func (r *TestResponse) AssertHeader(key, expected string) *TestResponse {
	actual := r.Headers.Get(key)
	if actual != expected {
		r.fail(fmt.Sprintf("header %q: expected %q, got %q", key, expected, actual))
	}
	return r
}

// AssertHeaderNotEmpty verifies that the given HTTP header exists and is not empty.
func (r *TestResponse) AssertHeaderNotEmpty(key string) *TestResponse {
	actual := r.Headers.Get(key)
	if actual == "" {
		r.fail(fmt.Sprintf("header %q is empty or missing", key))
	}
	return r
}

// AssertBodyContains verifies that the response body contains the substring.
func (r *TestResponse) AssertBodyContains(substr string) *TestResponse {
	if !strings.Contains(string(r.Body), substr) {
		r.fail(fmt.Sprintf("body %q does not contain %q", string(r.Body), substr))
	}
	return r
}

// AssertJSONPath verifies the value at a simple dot-separated JSON path.
func (r *TestResponse) AssertJSONPath(path string, expected any) *TestResponse {
	var parsed any
	if err := json.Unmarshal(r.Body, &parsed); err != nil {
		r.fail(fmt.Sprintf("failed parsing response as JSON: %v. Body: %s", err, string(r.Body)))
		return r
	}

	parts := strings.Split(path, ".")
	curr := parsed

	for _, part := range parts {
		m, ok := curr.(map[string]any)
		if !ok {
			r.fail(fmt.Sprintf("json path %q: segment %q is not an object", path, part))
			return r
		}
		val, exists := m[part]
		if !exists {
			r.fail(fmt.Sprintf("json path %q: key %q not found", path, part))
			return r
		}
		curr = val
	}

	// Compare with expected (handling float64 vs int JSON normalization)
	if !jsonEqual(curr, expected) {
		r.fail(fmt.Sprintf("json path %q: expected %v (%T), got %v (%T)", path, expected, expected, curr, curr))
	}

	return r
}

// BindJSON decodes the response body into dst.
func (r *TestResponse) BindJSON(dst any) error {
	return json.Unmarshal(r.Body, dst)
}

func (r *TestResponse) fail(msg string) {
	if r.t != nil {
		r.t.Helper()
		r.t.Fatalf("testutil: %s", msg)
	} else {
		panic("testutil: " + msg)
	}
}

func jsonEqual(actual, expected any) bool {
	if reflect.DeepEqual(actual, expected) {
		return true
	}

	// Handle numeric type discrepancies between JSON unmarshaling (float64) and test literals (int, int64)
	actVal := reflect.ValueOf(actual)
	expVal := reflect.ValueOf(expected)

	if actVal.IsValid() && expVal.IsValid() {
		if actVal.Kind() == reflect.Float64 && (expVal.Kind() == reflect.Int || expVal.Kind() == reflect.Int64) {
			return actVal.Float() == float64(expVal.Int())
		}
	}

	return false
}
