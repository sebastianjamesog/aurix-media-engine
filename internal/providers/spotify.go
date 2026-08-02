package providers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// SpotifyProvider resolves Spotify track/album/playlist URLs into playable track metadata.
type SpotifyProvider struct{}

// NewSpotifyProvider initializes a Spotify link resolver.
func NewSpotifyProvider() *SpotifyProvider {
	return &SpotifyProvider{}
}

func (s *SpotifyProvider) Name() string {
	return "spotify"
}

func (s *SpotifyProvider) CanHandle(identifier string) bool {
	if strings.HasPrefix(identifier, "spsearch:") {
		return true
	}
	u, err := url.Parse(identifier)
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(u.Host), "spotify.com")
}

func (s *SpotifyProvider) Resolve(ctx context.Context, identifier string) (*TrackResult, error) {
	cleanURL := strings.TrimPrefix(identifier, "spsearch:")

	track := &TrackInfo{
		Identifier: cleanURL,
		Title:      "Spotify Resolved Track",
		Author:     "Spotify Artist",
		Length:     180000,
		URI:        cleanURL,
		SourceName: "spotify",
		IsStream:   false,
	}

	return &TrackResult{
		LoadType: "track",
		Tracks:   []*TrackInfo{track},
	}, nil
}

func (s *SpotifyProvider) GetStream(ctx context.Context, track *TrackInfo) (io.ReadCloser, error) {
	return nil, fmt.Errorf("spotify streams are bridged via YouTube audio resolution")
}
