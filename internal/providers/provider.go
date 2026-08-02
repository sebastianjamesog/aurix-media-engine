package providers

import (
	"context"
	"io"
)

// TrackInfo holds details about a resolved audio track.
type TrackInfo struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	Length     int64  `json:"length"` // in milliseconds
	URI        string `json:"uri"`
	SourceName string `json:"sourceName"`
	IsStream   bool   `json:"isStream"`
}

// TrackResult represents the output of resolving a search query or URL.
type TrackResult struct {
	LoadType   string       `json:"loadType"` // "track", "playlist", "search", "empty", "error"
	Tracks     []*TrackInfo `json:"tracks"`
	Exception  error        `json:"-"`
}

// SourceProvider is the common interface implemented by all audio source providers.
type SourceProvider interface {
	Name() string
	CanHandle(identifier string) bool
	Resolve(ctx context.Context, identifier string) (*TrackResult, error)
	GetStream(ctx context.Context, track *TrackInfo) (io.ReadCloser, error)
}

// Registry manages registered audio source providers.
type Registry struct {
	providers []SourceProvider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make([]SourceProvider, 0),
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p SourceProvider) {
	r.providers = append(r.providers, p)
}

// Resolve iterates through registered providers to resolve a track query.
func (r *Registry) Resolve(ctx context.Context, identifier string) (*TrackResult, error) {
	for _, p := range r.providers {
		if p.CanHandle(identifier) {
			return p.Resolve(ctx, identifier)
		}
	}
	return &TrackResult{LoadType: "empty", Tracks: []*TrackInfo{}}, nil
}
