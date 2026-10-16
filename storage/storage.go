// Package storage provides file storage abstractions and contracts.
package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the contract for persisting, reading, and managing files.
type Storage interface {
	// Put writes content from reader to the specified path.
	Put(ctx context.Context, path string, content io.Reader) error

	// Get opens the file at path for reading. The caller is responsible for closing the ReadCloser.
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the file at path.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at path.
	Exists(ctx context.Context, path string) (bool, error)

	// Size returns the file size in bytes.
	Size(ctx context.Context, path string) (int64, error)

	// URL returns a public or presigned access URL for the file.
	URL(ctx context.Context, path string, ttl ...time.Duration) (string, error)
}

// Driver represents a storage provider capable of opening a Storage instance.
type Driver interface {
	Open() (Storage, error)
}
