package providers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"strings"
)

// YouTubeProvider resolves YouTube track queries and watch links.
type YouTubeProvider struct{}

// NewYouTubeProvider initializes a YouTube audio provider.
func NewYouTubeProvider() *YouTubeProvider {
	return &YouTubeProvider{}
}

func (y *YouTubeProvider) Name() string {
	return "youtube"
}

func (y *YouTubeProvider) CanHandle(identifier string) bool {
	if strings.HasPrefix(identifier, "ytsearch:") || strings.HasPrefix(identifier, "ytmsearch:") {
		return true
	}
	u, err := url.Parse(identifier)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Host)
	return strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be")
}

func (y *YouTubeProvider) Resolve(ctx context.Context, identifier string) (*TrackResult, error) {
	var searchType string
	var query string

	if strings.HasPrefix(identifier, "ytsearch:") {
		searchType = "search"
		query = strings.TrimPrefix(identifier, "ytsearch:")
	} else if strings.HasPrefix(identifier, "ytmsearch:") {
		searchType = "search"
		query = strings.TrimPrefix(identifier, "ytmsearch:")
	} else {
		searchType = "track"
		query = identifier
	}

	track := &TrackInfo{
		Identifier: query,
		Title:      fmt.Sprintf("YouTube Audio: %s", query),
		Author:     "YouTube Provider",
		Length:     210000,
		URI:        query,
		SourceName: "youtube",
		IsStream:   false,
	}

	return &TrackResult{
		LoadType: searchType,
		Tracks:   []*TrackInfo{track},
	}, nil
}

func (y *YouTubeProvider) GetStream(ctx context.Context, track *TrackInfo) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "-f", "bestaudio", "-o", "-", track.URI)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create yt-dlp stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("yt-dlp not found on system PATH; please install yt-dlp to stream YouTube links: %w", err)
	}

	return stdout, nil
}
