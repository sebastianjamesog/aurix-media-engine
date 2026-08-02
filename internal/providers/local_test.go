package providers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProvider(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.mp3")

	err := os.WriteFile(testFile, []byte("fake mp3 data"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	provider := NewLocalProvider(tempDir)
	ctx := context.Background()

	if !provider.CanHandle("test.mp3") {
		t.Errorf("expected CanHandle(test.mp3) to be true")
	}

	result, err := provider.Resolve(ctx, "test.mp3")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if result.LoadType != "track" || len(result.Tracks) != 1 {
		t.Fatalf("expected 1 track loaded, got loadType=%s, len=%d", result.LoadType, len(result.Tracks))
	}

	if result.Tracks[0].Title != "test.mp3" {
		t.Errorf("expected track title test.mp3, got %s", result.Tracks[0].Title)
	}

	// Security path traversal test
	traversalRes, _ := provider.Resolve(ctx, "../secret.txt")
	if traversalRes.LoadType != "error" {
		t.Errorf("expected error for path traversal attempt, got loadType=%s", traversalRes.LoadType)
	}
}
