// Package localfs provides a local disk filesystem implementation of storage.Storage.
package localfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ishubhamsingh2e/bourbon/core"
	"github.com/ishubhamsingh2e/bourbon/storage"
)

// Config configures local filesystem storage settings.
type Config struct {
	RootDir string
	BaseURL string
}

// Open satisfies the storage.Driver interface.
func (c Config) Open() (storage.Storage, error) {
	root := c.RootDir
	if root == "" {
		root = "storage/app"
	}
	return New(root, c.BaseURL)
}

// Storage implements storage.Storage on top of the local operating system filesystem.
type Storage struct {
	rootDir string
	baseURL string
}

// New constructs a LocalFS storage driver rooted at rootDir.
func New(rootDir string, baseURL ...string) (*Storage, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("localfs: invalid root directory: %w", err)
	}

	if err := os.MkdirAll(absRoot, 0755); err != nil {
		return nil, fmt.Errorf("localfs: failed creating root directory: %w", err)
	}

	base := "/storage"
	if len(baseURL) > 0 && baseURL[0] != "" {
		base = baseURL[0]
	}

	return &Storage{
		rootDir: absRoot,
		baseURL: strings.TrimRight(base, "/"),
	}, nil
}

func (s *Storage) resolvePath(relPath string) (string, error) {
	cleaned := filepath.Clean(relPath)
	if strings.HasPrefix(cleaned, "..") || strings.HasPrefix(cleaned, "/") {
		return "", errors.New("localfs: path traversal attempted")
	}
	return filepath.Join(s.rootDir, cleaned), nil
}

// Put writes stream content to the specified relative path.
func (s *Storage) Put(ctx context.Context, path string, content io.Reader) error {
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("localfs: failed creating directories for '%s': %w", path, err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("localfs: failed creating file '%s': %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, content); err != nil {
		return fmt.Errorf("localfs: failed writing content to '%s': %w", path, err)
	}

	return nil
}

// Get opens the file for reading.
func (s *Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("localfs: file '%s' not found: %w", path, err)
	}

	return file, nil
}

// Delete removes the file at relative path.
func (s *Storage) Delete(ctx context.Context, path string) error {
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return err
	}

	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("localfs: failed deleting file '%s': %w", path, err)
	}

	return nil
}

// Exists checks if the file exists.
func (s *Storage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return false, err
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}

// Size returns the file size in bytes.
func (s *Storage) Size(ctx context.Context, path string) (int64, error) {
	fullPath, err := s.resolvePath(path)
	if err != nil {
		return 0, err
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("localfs: failed reading file stats '%s': %w", path, err)
	}

	return stat.Size(), nil
}

// URL returns the public URL for the relative path.
func (s *Storage) URL(ctx context.Context, path string, ttl ...time.Duration) (string, error) {
	cleaned := strings.TrimLeft(filepath.ToSlash(filepath.Clean(path)), "/")
	return fmt.Sprintf("%s/%s", s.baseURL, cleaned), nil
}

// Provider implements core.Provider for local filesystem storage integration.
type Provider struct {
	cfg Config
	fs  *Storage
}

// NewProvider returns a new LocalFS Provider with the given configuration.
func NewProvider(cfg Config) *Provider {
	return &Provider{cfg: cfg}
}

// Register registers storage.Storage into the DI container as a singleton.
func (p *Provider) Register(c core.Container) error {
	c.Singleton(func() (storage.Storage, error) {
		s, err := p.cfg.Open()
		if err != nil {
			return nil, err
		}
		if typed, ok := s.(*Storage); ok {
			p.fs = typed
		}
		return s, nil
	})
	return nil
}

// Boot verifies storage readiness.
func (p *Provider) Boot(ctx context.Context, app core.Application) error {
	var st storage.Storage
	return app.Container().Resolve(&st)
}

// Shutdown releases storage resources.
func (p *Provider) Shutdown(ctx context.Context) error {
	return nil
}

var _ core.Provider = (*Provider)(nil)
