package s3

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestS3StorageOperations(t *testing.T) {
	s := New(Config{
		Bucket: "my-bucket",
		Region: "us-east-1",
	})

	ctx := context.Background()
	content := []byte("cloud storage data")

	err := s.Put(ctx, "avatars/user_1.png", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	exists, err := s.Exists(ctx, "avatars/user_1.png")
	if err != nil || !exists {
		t.Fatalf("expected file to exist")
	}

	size, err := s.Size(ctx, "avatars/user_1.png")
	if err != nil || size != int64(len(content)) {
		t.Fatalf("size mismatch: %d", size)
	}

	rc, err := s.Get(ctx, "avatars/user_1.png")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer rc.Close()

	readBytes, _ := io.ReadAll(rc)
	if string(readBytes) != "cloud storage data" {
		t.Errorf("content mismatch: %s", string(readBytes))
	}

	url, _ := s.URL(ctx, "avatars/user_1.png")
	if url != "https://my-bucket.s3.us-east-1.amazonaws.com/avatars/user_1.png" {
		t.Errorf("URL mismatch: %s", url)
	}
}
