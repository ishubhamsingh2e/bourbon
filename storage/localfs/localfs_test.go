package localfs

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/app"
	"github.com/ishubhamsingh2e/bourbon/storage"
)

func TestLocalFSOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_storage_test_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := New(tempDir, "https://cdn.example.com/uploads")
	if err != nil {
		t.Fatalf("New LocalFS failed: %v", err)
	}

	ctx := context.Background()
	content := []byte("hello storage world")

	// 1. Put file
	err = store.Put(ctx, "sub/dir/test.txt", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// 2. Exists
	exists, err := store.Exists(ctx, "sub/dir/test.txt")
	if err != nil || !exists {
		t.Fatalf("expected file to exist, got: %v, err: %v", exists, err)
	}

	// 3. Size
	size, err := store.Size(ctx, "sub/dir/test.txt")
	if err != nil || size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d, err: %v", len(content), size, err)
	}

	// 4. Get file
	rc, err := store.Get(ctx, "sub/dir/test.txt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer rc.Close()

	readBytes, err := io.ReadAll(rc)
	if err != nil || string(readBytes) != "hello storage world" {
		t.Fatalf("content mismatch: %s", string(readBytes))
	}

	// 5. URL
	url, _ := store.URL(ctx, "sub/dir/test.txt")
	if url != "https://cdn.example.com/uploads/sub/dir/test.txt" {
		t.Errorf("URL mismatch: %s", url)
	}

	// 6. Delete
	err = store.Delete(ctx, "sub/dir/test.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	existsAfter, _ := store.Exists(ctx, "sub/dir/test.txt")
	if existsAfter {
		t.Errorf("file should not exist after deletion")
	}

	// 7. Prevent directory traversal
	err = store.Put(ctx, "../../secret.txt", bytes.NewReader(content))
	if err == nil || !strings.Contains(err.Error(), "traversal") {
		t.Errorf("expected directory traversal error, got: %v", err)
	}
}

func TestLocalFSProviderLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bourbon_storage_provider_*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	provider := NewProvider(Config{RootDir: tempDir, BaseURL: "http://localhost:8080/storage"})
	application := app.New()

	if err := provider.Register(application.Container()); err != nil {
		t.Fatalf("failed registering localfs provider: %v", err)
	}

	var st storage.Storage
	if err := application.Container().Resolve(&st); err != nil {
		t.Fatalf("failed resolving storage.Storage: %v", err)
	}

	ctx := context.Background()
	if err := provider.Boot(ctx, application); err != nil {
		t.Fatalf("failed booting localfs provider: %v", err)
	}

	if err := provider.Shutdown(ctx); err != nil {
		t.Fatalf("failed shutting down localfs provider: %v", err)
	}
}
