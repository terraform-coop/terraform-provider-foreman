package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Autosign represents a smart proxy autosign entry: a hostname or
// wildcard pattern allowed to automatically sign its certificate on next
// registration. Unlike most Foreman entities, its own "id" is not a
// server-assigned integer but the pattern string itself, it lives under a
// specific smart proxy (/api/smart_proxies/:smart_proxy_id/autosign), and
// there is no "show" endpoint to read a single entry back by id.
type Autosign struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateAutosign creates an autosign entry for the given pattern on
// the given smart proxy.
func (c *Client) CreateAutosign(ctx context.Context, smartProxyID int, pattern string) (*Autosign, error) {
	body := map[string]string{"id": pattern}
	var resp Autosign
	if err := c.Post(ctx, fmt.Sprintf("smart_proxies/%d/autosign", smartProxyID), "", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteAutosign removes an autosign entry.
func (c *Client) DeleteAutosign(ctx context.Context, smartProxyID int, pattern string) error {
	return c.Delete(ctx, fmt.Sprintf("smart_proxies/%d/autosign/%s", smartProxyID, url.PathEscape(pattern)))
}

// ReadAutosign finds an autosign entry by pattern. There is no
// per-entry show endpoint, so this lists all entries for the smart proxy
// and matches client-side. Returns (nil, nil) if not found.
func (c *Client) ReadAutosign(ctx context.Context, smartProxyID int, pattern string) (*Autosign, error) {
	var response QueryResponse
	if err := c.Get(ctx, fmt.Sprintf("smart_proxies/%d/autosign", smartProxyID), &response); err != nil {
		return nil, err
	}
	for _, raw := range response.Results {
		var entry Autosign
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		if entry.ID == pattern {
			return &entry, nil
		}
	}
	return nil, nil
}
