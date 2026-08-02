package aurix

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"aurix/internal/player"
	"aurix/internal/providers"
)

// Client represents an Aurix Media Engine Go SDK client.
type Client struct {
	baseURL    string
	password   string
	httpClient *http.Client
}

// NewClient initializes a new AME SDK client.
func NewClient(baseURL string, password string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:2333"
	}
	return &Client{
		baseURL:  baseURL,
		password: password,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoadTracks queries Aurix REST API to resolve a track query or URL.
func (c *Client) LoadTracks(ctx context.Context, identifier string) (*providers.TrackResult, error) {
	endpoint := fmt.Sprintf("%s/v4/loadtracks?identifier=%s", c.baseURL, url.QueryEscape(identifier))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected loadtracks status code: %d", resp.StatusCode)
	}

	var res providers.TrackResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

// GetStats fetches engine runtime telemetry.
func (c *Client) GetStats(ctx context.Context) (*player.Stats, error) {
	endpoint := fmt.Sprintf("%s/v4/stats", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats player.Stats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
