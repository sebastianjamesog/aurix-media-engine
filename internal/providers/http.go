package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// HTTPProvider resolves and streams direct audio URLs.
type HTTPProvider struct {
	client *http.Client
}

// NewHTTPProvider creates a new HTTP audio provider.
func NewHTTPProvider() *HTTPProvider {
	return &HTTPProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *HTTPProvider) Name() string {
	return "http"
}

func (h *HTTPProvider) CanHandle(identifier string) bool {
	u, err := url.Parse(identifier)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	ext := strings.ToLower(path.Ext(u.Path))
	switch ext {
	case ".mp3", ".ogg", ".wav", ".flac", ".aac", ".m4a", ".webm", ".opus", ".m3u8":
		return true
	}
	return true // Fallback for direct stream links
}

func (h *HTTPProvider) Resolve(ctx context.Context, identifier string) (*TrackResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, identifier, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &TrackResult{
			LoadType:  "error",
			Exception: fmt.Errorf("HTTP error %d", resp.StatusCode),
		}, nil
	}

	u, _ := url.Parse(identifier)
	fileName := path.Base(u.Path)
	if fileName == "." || fileName == "/" {
		fileName = "HTTP Audio Stream"
	}

	track := &TrackInfo{
		Identifier: identifier,
		Title:      fileName,
		Author:     u.Host,
		Length:     0, // Live stream or unknown length
		URI:        identifier,
		SourceName: "http",
		IsStream:   true,
	}

	return &TrackResult{
		LoadType: "track",
		Tracks:   []*TrackInfo{track},
	}, nil
}

func (h *HTTPProvider) GetStream(ctx context.Context, track *TrackInfo) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, track.URI, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP stream error %d", resp.StatusCode)
	}

	return resp.Body, nil
}
