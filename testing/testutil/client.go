package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestClient provides fluent HTTP client methods for sending requests and asserting responses.
type TestClient struct {
	baseURL string
	t       *testing.T
	client  *http.Client
}

// NewClient initializes a TestClient against the given base URL.
func NewClient(t *testing.T, baseURL string) *TestClient {
	return &TestClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		t:       t,
		client:  &http.Client{},
	}
}

// Request executes an arbitrary HTTP request and returns a TestResponse.
func (c *TestClient) Request(method, path string, body io.Reader, headers map[string]string) *TestResponse {
	url := c.baseURL + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		if c.t != nil {
			c.t.Fatalf("testclient: failed creating request: %v", err)
		}
		return nil
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if c.t != nil {
			c.t.Fatalf("testclient: request failed: %v", err)
		}
		return nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		if c.t != nil {
			c.t.Fatalf("testclient: failed reading response body: %v", err)
		}
		return nil
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
		t:          c.t,
	}
}

// Get sends a GET request to path.
func (c *TestClient) Get(path string) *TestResponse {
	return c.Request(http.MethodGet, path, nil, nil)
}

// PostJSON sends a POST request with JSON payload.
func (c *TestClient) PostJSON(path string, body any) *TestResponse {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	return c.Request(http.MethodPost, path, &buf, map[string]string{
		"Content-Type": "application/json",
	})
}

// PutJSON sends a PUT request with JSON payload.
func (c *TestClient) PutJSON(path string, body any) *TestResponse {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	return c.Request(http.MethodPut, path, &buf, map[string]string{
		"Content-Type": "application/json",
	})
}

// PatchJSON sends a PATCH request with JSON payload.
func (c *TestClient) PatchJSON(path string, body any) *TestResponse {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	return c.Request(http.MethodPatch, path, &buf, map[string]string{
		"Content-Type": "application/json",
	})
}

// Delete sends a DELETE request to path.
func (c *TestClient) Delete(path string) *TestResponse {
	return c.Request(http.MethodDelete, path, nil, nil)
}
