package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ForemanAutosign represents a smart proxy autosign entry: a hostname or
// wildcard pattern allowed to automatically sign its certificate on next
// registration. Unlike most Foreman entities, its own "id" is not a
// server-assigned integer but the pattern string itself, it lives under a
// specific smart proxy (/api/smart_proxies/:smart_proxy_id/autosign), and
// there is no "show" endpoint to read a single entry back by id.
type ForemanAutosign struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at,omitempty"`
}

// CreateForemanAutosign creates an autosign entry for the given pattern on
// the given smart proxy.
func (c *ForemanClient) CreateForemanAutosign(ctx context.Context, smartProxyID int, pattern string) (*ForemanAutosign, error) {
	body := map[string]string{"id": pattern}
	var resp ForemanAutosign
	if err := c.Post(ctx, fmt.Sprintf("smart_proxies/%d/autosign", smartProxyID), "", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteForemanAutosign removes an autosign entry.
func (c *ForemanClient) DeleteForemanAutosign(ctx context.Context, smartProxyID int, pattern string) error {
	return c.Delete(ctx, fmt.Sprintf("smart_proxies/%d/autosign/%s", smartProxyID, url.PathEscape(pattern)))
}

// ReadForemanAutosign finds an autosign entry by pattern. There is no
// per-entry show endpoint, so this lists all entries for the smart proxy
// and matches client-side. Returns (nil, nil) if not found.
func (c *ForemanClient) ReadForemanAutosign(ctx context.Context, smartProxyID int, pattern string) (*ForemanAutosign, error) {
	var response QueryResponse
	if err := c.Get(ctx, fmt.Sprintf("smart_proxies/%d/autosign", smartProxyID), &response); err != nil {
		return nil, err
	}
	for _, raw := range response.Results {
		var entry ForemanAutosign
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}
		if entry.ID == pattern {
			return &entry, nil
		}
	}
	return nil, nil
}
