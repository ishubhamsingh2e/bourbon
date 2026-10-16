package contracts

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/storage"
)

// VerifyStorageContract executes a standard compliance test suite against any storage.Storage implementation.
func VerifyStorageContract(t *testing.T, store storage.Storage) {
	ctx := context.Background()
	testPath := "contract_tests/file.txt"
	testContent := []byte("contract test content 12345")

	t.Run("Put_Exists_Get_Delete", func(t *testing.T) {
		// Put
		err := store.Put(ctx, testPath, bytes.NewReader(testContent))
		if err != nil {
			t.Fatalf("Put failed: %v", err)
		}

		// Exists
		exists, err := store.Exists(ctx, testPath)
		if err != nil || !exists {
			t.Fatalf("expected file to exist, got: %v, err: %v", exists, err)
		}

		// Size
		size, err := store.Size(ctx, testPath)
		if err != nil || size != int64(len(testContent)) {
			t.Fatalf("expected size %d, got %d, err: %v", len(testContent), size, err)
		}

		// Get
		rc, err := store.Get(ctx, testPath)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		defer rc.Close()

		readBytes, err := io.ReadAll(rc)
		if err != nil || string(readBytes) != string(testContent) {
			t.Fatalf("content mismatch: %s", string(readBytes))
		}

		// URL
		url, err := store.URL(ctx, testPath)
		if err != nil || url == "" {
			t.Fatalf("URL generation failed: %v, url: %s", err, url)
		}

		// Delete
		err = store.Delete(ctx, testPath)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		existsAfter, _ := store.Exists(ctx, testPath)
		if existsAfter {
			t.Errorf("file should not exist after deletion")
		}
	})
}
