package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalProvider resolves and streams local media files safely.
type LocalProvider struct {
	baseDir string
}

// NewLocalProvider creates a new local file provider rooted at baseDir.
func NewLocalProvider(baseDir string) *LocalProvider {
	if baseDir == "" {
		baseDir = "."
	}
	abs, err := filepath.Abs(baseDir)
	if err == nil {
		baseDir = abs
	}
	return &LocalProvider{baseDir: baseDir}
}

func (l *LocalProvider) Name() string {
	return "local"
}

func (l *LocalProvider) CanHandle(identifier string) bool {
	if strings.HasPrefix(identifier, "local:") || strings.HasPrefix(identifier, "file://") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(identifier))
	switch ext {
	case ".mp3", ".flac", ".wav", ".ogg", ".aac", ".m4a", ".webm":
		return true
	}
	return false
}

func (l *LocalProvider) Resolve(ctx context.Context, identifier string) (*TrackResult, error) {
	cleanPath := strings.TrimPrefix(identifier, "local:")
	cleanPath = strings.TrimPrefix(cleanPath, "file://")

	targetPath := filepath.Clean(cleanPath)
	if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(l.baseDir, targetPath)
	}

	// Security check: ensure target path remains within baseDir if relative
	rel, err := filepath.Rel(l.baseDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return &TrackResult{
			LoadType:  "error",
			Exception: fmt.Errorf("access denied: path outside allowed directory"),
		}, nil
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TrackResult{LoadType: "empty", Tracks: []*TrackInfo{}}, nil
		}
		return &TrackResult{LoadType: "error", Exception: err}, nil
	}

	if info.IsDir() {
		return &TrackResult{LoadType: "empty", Tracks: []*TrackInfo{}}, nil
	}

	track := &TrackInfo{
		Identifier: targetPath,
		Title:      info.Name(),
		Author:     "Local File",
		Length:     0,
		URI:        targetPath,
		SourceName: "local",
		IsStream:   false,
	}

	return &TrackResult{
		LoadType: "track",
		Tracks:   []*TrackInfo{track},
	}, nil
}

func (l *LocalProvider) GetStream(ctx context.Context, track *TrackInfo) (io.ReadCloser, error) {
	return os.Open(track.URI)
}
