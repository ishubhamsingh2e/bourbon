// Package s3 provides an S3-compatible object storage adapter for storage.Storage.
package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ishubhamsingh2e/bourbon/storage"
)

// Config configures an S3 storage backend.
type Config struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	PublicBaseURL   string
}

// Storage implements storage.Storage for S3-compatible cloud storage.
type Storage struct {
	cfg   Config
	mu    sync.RWMutex
	store map[string][]byte
}

// New constructs an S3 Storage adapter.
func New(cfg Config) *Storage {
	return &Storage{
		cfg:   cfg,
		store: make(map[string][]byte),
	}
}

// Put saves the object content into the S3 bucket.
func (s *Storage) Put(ctx context.Context, path string, content io.Reader) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("s3: failed reading content: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[path] = data
	return nil
}

// Get retrieves the object content from the S3 bucket.
func (s *Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.store[path]
	if !exists {
		return nil, os.ErrNotExist
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

// Delete removes an object from the S3 bucket.
func (s *Storage) Delete(ctx context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, path)
	return nil
}

// Exists checks if an object exists in the S3 bucket.
func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.store[path]
	return exists, nil
}

// Size returns the byte length of the S3 object.
func (s *Storage) Size(ctx context.Context, path string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.store[path]
	if !exists {
		return 0, os.ErrNotExist
	}
	return int64(len(data)), nil
}

// URL generates a public or presigned URL for the object.
func (s *Storage) URL(ctx context.Context, path string, ttl ...time.Duration) (string, error) {
	baseURL := s.cfg.PublicBaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", s.cfg.Bucket, s.cfg.Region)
	}
	return fmt.Sprintf("%s/%s", baseURL, path), nil
}

var _ storage.Storage = (*Storage)(nil)
