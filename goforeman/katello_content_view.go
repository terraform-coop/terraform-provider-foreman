package goforeman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

type KatelloContentViewRequest struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Label          string `json:"label,omitempty"`
	OrganizationID int    `json:"organization_id,omitempty"`
	// Composite/AutoPublish/SolveDependencies/Filtered are *bool WITH
	// omitempty (not plain bool+omitempty, and not *bool without
	// omitempty): plain bool+omitempty can never send an explicit "false"
	// (Go's zero value for bool is false), but *bool without omitempty
	// sends an explicit JSON null when genuinely unset - these columns are
	// NOT NULL like almost every boolean column, so that null gets
	// rejected outright. Go's encoding/json only treats a nil pointer as
	// "empty" for omitempty, so *bool+omitempty gets both right: nil omits
	// the key entirely, non-nil (even pointing at false) always sends it.
	Composite         *bool `json:"composite,omitempty"`
	AutoPublish       *bool `json:"auto_publish,omitempty"`
	SolveDependencies *bool `json:"solve_dependencies,omitempty"`
	Filtered          *bool `json:"filtered,omitempty"`
	RepositoryIDs     []int `json:"repository_ids,omitempty"`
	ComponentIDs      []int `json:"component_ids,omitempty"`
}

type KatelloContentView struct {
	Base
	Description       string `json:"description"`
	Label             string `json:"label"`
	OrganizationID    int    `json:"organization_id"`
	Composite         bool   `json:"composite"`
	AutoPublish       bool   `json:"auto_publish"`
	SolveDependencies bool   `json:"solve_dependencies"`
	Filtered          bool   `json:"filtered"`
	RepositoryIDs     []int  `json:"repository_ids"`
	ComponentIDs      []int  `json:"component_ids"`
	LatestVersionID   int    `json:"latest_version_id"`
	LatestVersion     string `json:"latest_version"`
	ContentHostCount  int    `json:"content_host_count"`
	VersionCount      int    `json:"version_count"`
}

func (c *Client) CreateKatelloContentView(ctx context.Context, req *KatelloContentViewRequest) (*KatelloContentView, error) {
	var resp KatelloContentView
	err := c.Post(ctx, "/katello/api/content_views", "content_view", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ReadKatelloContentView(ctx context.Context, id int) (*KatelloContentView, error) {
	var resp KatelloContentView
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateKatelloContentView(ctx context.Context, id int, req *KatelloContentViewRequest) (*KatelloContentView, error) {
	var resp KatelloContentView
	err := c.Put(ctx, fmt.Sprintf("/katello/api/content_views/%d", id), "content_view", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteKatelloContentView(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/content_views/%d", id))
}

// PublishContentView publishes a new version of a content view. The
// .../publish endpoint's response body is the async task, not the content
// view (unlike most other Katello 202s) - passing a nil respObj here avoids
// do() misinterpreting the task JSON as the content view (wrong ID, every
// other field blank), and the up-to-date content view is fetched
// separately once the task completes.
func (c *Client) PublishContentView(ctx context.Context, id int) (*KatelloContentView, error) {
	if err := c.Post(ctx, fmt.Sprintf("/katello/api/content_views/%d/publish", id), "", nil, nil); err != nil {
		return nil, err
	}
	return c.ReadKatelloContentView(ctx, id)
}

func (c *Client) FindKatelloContentViewByName(ctx context.Context, name string) (*KatelloContentView, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj KatelloContentView
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ---------------------------------------------------------------------------
// Content View Filters — managed via separate API endpoints
// ---------------------------------------------------------------------------

type KatelloContentViewFilter struct {
	ID          int                            `json:"id"`
	Name        string                         `json:"name"`
	Type        string                         `json:"type"`
	Inclusion   bool                           `json:"inclusion"`
	Description string                         `json:"description"`
	Rules       []KatelloContentViewFilterRule `json:"rules"`
}

type KatelloContentViewFilterRule struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Architecture string `json:"architecture,omitempty"`
}

// ReadContentViewFilters returns all filters (including rules) for a content view.
func (c *Client) ReadContentViewFilters(ctx context.Context, cvID int) ([]KatelloContentViewFilter, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters", cvID), &response)
	if err != nil {
		return nil, err
	}
	var filters []KatelloContentViewFilter
	var parseErrors []error
	for _, raw := range response.Results {
		var f KatelloContentViewFilter
		if err := json.Unmarshal(raw, &f); err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("parsing filter: %w", err))
			continue
		}
		filters = append(filters, f)
	}
	if len(parseErrors) > 0 {
		return filters, fmt.Errorf("failed to parse %d filters: %w", len(parseErrors), errors.Join(parseErrors...))
	}

	for i := range filters {
		rules, err := c.ReadContentViewFilterRules(ctx, filters[i].ID)
		if err != nil {
			return filters, fmt.Errorf("reading rules for filter %d: %w", filters[i].ID, err)
		}
		filters[i].Rules = rules
	}

	return filters, nil
}

// CreateContentViewFilter creates a single filter on a content view, then creates its rules.
func (c *Client) CreateContentViewFilter(ctx context.Context, cvID int, filter *KatelloContentViewFilter) (*KatelloContentViewFilter, error) {
	var resp KatelloContentViewFilter
	err := c.Post(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters", cvID), "content_view_filter", filter, &resp)
	if err != nil {
		return nil, err
	}
	// Create rules
	if len(filter.Rules) > 0 {
		rules, err := c.CreateContentViewFilterRules(ctx, resp.ID, filter.Rules)
		if err != nil {
			return nil, err
		}
		resp.Rules = rules
	}
	return &resp, nil
}

// UpdateContentViewFilter updates a single filter on a content view.
func (c *Client) UpdateContentViewFilter(ctx context.Context, cvID int, filter *KatelloContentViewFilter) (*KatelloContentViewFilter, error) {
	var resp KatelloContentViewFilter
	err := c.Put(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters/%d", cvID, filter.ID), "content_view_filter", filter, &resp)
	if err != nil {
		return nil, err
	}
	// Update rules
	if len(filter.Rules) > 0 {
		rules, err := c.UpdateContentViewFilterRules(ctx, filter.ID, filter.Rules)
		if err != nil {
			return nil, err
		}
		resp.Rules = rules
	}
	return &resp, nil
}

// DeleteContentViewFilter deletes a single filter from a content view.
func (c *Client) DeleteContentViewFilter(ctx context.Context, cvID int, filterID int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters/%d", cvID, filterID))
}

// ReadContentViewFilterRules returns all rules for a filter.
func (c *Client) ReadContentViewFilterRules(ctx context.Context, filterID int) ([]KatelloContentViewFilterRule, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules", filterID), &response)
	if err != nil {
		return nil, err
	}
	var rules []KatelloContentViewFilterRule
	var parseErrors []error
	for _, raw := range response.Results {
		var r KatelloContentViewFilterRule
		if err := json.Unmarshal(raw, &r); err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("parsing rule: %w", err))
			continue
		}
		rules = append(rules, r)
	}
	if len(parseErrors) > 0 {
		return rules, fmt.Errorf("failed to parse %d rules: %w", len(parseErrors), errors.Join(parseErrors...))
	}
	return rules, nil
}

// CreateContentViewFilterRules creates rules on a filter.
func (c *Client) CreateContentViewFilterRules(ctx context.Context, filterID int, rules []KatelloContentViewFilterRule) ([]KatelloContentViewFilterRule, error) {
	var created []KatelloContentViewFilterRule
	for _, rule := range rules {
		var resp KatelloContentViewFilterRule
		err := c.Post(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules", filterID), "content_view_filter_rule", &rule, &resp)
		if err != nil {
			return nil, err
		}
		created = append(created, resp)
	}
	return created, nil
}

// UpdateContentViewFilterRules updates rules on a filter.
func (c *Client) UpdateContentViewFilterRules(ctx context.Context, filterID int, rules []KatelloContentViewFilterRule) ([]KatelloContentViewFilterRule, error) {
	var updated []KatelloContentViewFilterRule
	for _, rule := range rules {
		var resp KatelloContentViewFilterRule
		err := c.Put(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules/%d", filterID, rule.ID), "content_view_filter_rule", &rule, &resp)
		if err != nil {
			return nil, err
		}
		updated = append(updated, resp)
	}
	return updated, nil
}

// SyncContentViewFilters syncs filters for a content view: creates new, updates existing, deletes removed.
func (c *Client) SyncContentViewFilters(ctx context.Context, cvID int, desired []KatelloContentViewFilter) error {
	existing, err := c.ReadContentViewFilters(ctx, cvID)
	if err != nil {
		// If read fails (e.g., 404), treat as no existing filters
		existing = nil
	}

	existingByID := make(map[int]KatelloContentViewFilter)
	for _, f := range existing {
		existingByID[f.ID] = f
	}
	desiredByID := make(map[int]KatelloContentViewFilter)
	for _, f := range desired {
		if f.ID != 0 {
			desiredByID[f.ID] = f
		}
	}

	// Create new filters (no ID)
	for _, f := range desired {
		if f.ID == 0 {
			if _, err := c.CreateContentViewFilter(ctx, cvID, &f); err != nil {
				return fmt.Errorf("creating filter %q: %w", f.Name, err)
			}
		}
	}

	// Update existing filters
	for _, f := range desired {
		if f.ID != 0 {
			if _, err := c.UpdateContentViewFilter(ctx, cvID, &f); err != nil {
				return fmt.Errorf("updating filter %q: %w", f.Name, err)
			}
		}
	}

	// Delete removed filters
	for id := range existingByID {
		if _, ok := desiredByID[id]; !ok {
			if err := c.DeleteContentViewFilter(ctx, cvID, id); err != nil {
				return fmt.Errorf("deleting filter %d: %w", id, err)
			}
		}
	}

	return nil
}
