package providers

import (
	"context"
	"testing"
)

func TestHTTPProviderCanHandle(t *testing.T) {
	provider := NewHTTPProvider()

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/audio.mp3", true},
		{"http://example.com/stream.ogg", true},
		{"https://example.com/live.m3u8", true},
		{"not_a_url", false},
	}

	for _, tt := range tests {
		result := provider.CanHandle(tt.url)
		if result != tt.expected {
			t.Errorf("CanHandle(%s) = %v; want %v", tt.url, result, tt.expected)
		}
	}
}

func TestHTTPProviderResolve(t *testing.T) {
	provider := NewHTTPProvider()
	ctx := context.Background()

	// Resolving invalid domain should return loadType error or empty
	res, err := provider.Resolve(ctx, "https://invalid-non-existent-domain-1234567.com/file.mp3")
	if err == nil && res.LoadType == "track" {
		t.Errorf("expected error or empty for invalid domain, got track")
	}
}
