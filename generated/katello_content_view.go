package generated

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/sync/errgroup"
)

type ForemanKatelloContentViewRequest struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	Label             string `json:"label,omitempty"`
	OrganizationID    int    `json:"organization_id,omitempty"`
	Composite         bool   `json:"composite,omitempty"`
	AutoPublish       bool   `json:"auto_publish,omitempty"`
	SolveDependencies bool   `json:"solve_dependencies,omitempty"`
	Filtered          bool   `json:"filtered,omitempty"`
	RepositoryIDs     []int  `json:"repository_ids,omitempty"`
	ComponentIDs      []int  `json:"component_ids,omitempty"`
}

type ForemanKatelloContentView struct {
	ForemanObject
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

func (c *ForemanClient) CreateForemanKatelloContentView(ctx context.Context, req *ForemanKatelloContentViewRequest) (*ForemanKatelloContentView, error) {
	var resp ForemanKatelloContentView
	err := c.Post(ctx, "/katello/api/content_views", "content_view", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloContentView(ctx context.Context, id int) (*ForemanKatelloContentView, error) {
	var resp ForemanKatelloContentView
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloContentView(ctx context.Context, id int, req *ForemanKatelloContentViewRequest) (*ForemanKatelloContentView, error) {
	var resp ForemanKatelloContentView
	err := c.Put(ctx, fmt.Sprintf("/katello/api/content_views/%d", id), "content_view", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloContentView(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/content_views/%d", id))
}

// PublishContentView publishes a new version of a content view.
func (c *ForemanClient) PublishContentView(ctx context.Context, id int) (*ForemanKatelloContentView, error) {
	var resp ForemanKatelloContentView
	err := c.Post(ctx, fmt.Sprintf("/katello/api/content_views/%d/publish", id), "", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) QueryForemanKatelloContentView(ctx context.Context, name string) (*ForemanKatelloContentView, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloContentView
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

// ---------------------------------------------------------------------------
// Content View Filters — managed via separate API endpoints
// ---------------------------------------------------------------------------

type ForemanKatelloContentViewFilter struct {
	ID          int                                   `json:"id"`
	Name        string                                `json:"name"`
	Type        string                                `json:"type"`
	Inclusion   bool                                  `json:"inclusion"`
	Description string                                `json:"description"`
	Rules       []ForemanKatelloContentViewFilterRule `json:"rules"`
}

type ForemanKatelloContentViewFilterRule struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Architecture string `json:"architecture,omitempty"`
}

// ReadContentViewFilters returns all filters (including rules) for a content view.
func (c *ForemanClient) ReadContentViewFilters(ctx context.Context, cvID int) ([]ForemanKatelloContentViewFilter, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters", cvID), &response)
	if err != nil {
		return nil, err
	}
	var filters []ForemanKatelloContentViewFilter
	var parseErrors []error
	for _, raw := range response.Results {
		var f ForemanKatelloContentViewFilter
		if err := json.Unmarshal(raw, &f); err != nil {
			parseErrors = append(parseErrors, fmt.Errorf("parsing filter: %w", err))
			continue
		}
		filters = append(filters, f)
	}
	if len(parseErrors) > 0 {
		return filters, fmt.Errorf("failed to parse %d filters: %w", len(parseErrors), errors.Join(parseErrors...))
	}

	// Fetch rules concurrently for all filters
	g, gctx := errgroup.WithContext(ctx)
	for i := range filters {
		i := i
		g.Go(func() error {
			rules, err := c.ReadContentViewFilterRules(gctx, filters[i].ID)
			if err != nil {
				return fmt.Errorf("reading rules for filter %d: %w", filters[i].ID, err)
			}
			filters[i].Rules = rules
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return filters, err
	}

	return filters, nil
}

// CreateContentViewFilter creates a single filter on a content view, then creates its rules.
func (c *ForemanClient) CreateContentViewFilter(ctx context.Context, cvID int, filter *ForemanKatelloContentViewFilter) (*ForemanKatelloContentViewFilter, error) {
	var resp ForemanKatelloContentViewFilter
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
func (c *ForemanClient) UpdateContentViewFilter(ctx context.Context, cvID int, filter *ForemanKatelloContentViewFilter) (*ForemanKatelloContentViewFilter, error) {
	var resp ForemanKatelloContentViewFilter
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
func (c *ForemanClient) DeleteContentViewFilter(ctx context.Context, cvID int, filterID int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/content_views/%d/filters/%d", cvID, filterID))
}

// ReadContentViewFilterRules returns all rules for a filter.
func (c *ForemanClient) ReadContentViewFilterRules(ctx context.Context, filterID int) ([]ForemanKatelloContentViewFilterRule, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules", filterID), &response)
	if err != nil {
		return nil, err
	}
	var rules []ForemanKatelloContentViewFilterRule
	var parseErrors []error
	for _, raw := range response.Results {
		var r ForemanKatelloContentViewFilterRule
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
func (c *ForemanClient) CreateContentViewFilterRules(ctx context.Context, filterID int, rules []ForemanKatelloContentViewFilterRule) ([]ForemanKatelloContentViewFilterRule, error) {
	var created []ForemanKatelloContentViewFilterRule
	for _, rule := range rules {
		var resp ForemanKatelloContentViewFilterRule
		err := c.Post(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules", filterID), "content_view_filter_rule", &rule, &resp)
		if err != nil {
			return nil, err
		}
		created = append(created, resp)
	}
	return created, nil
}

// UpdateContentViewFilterRules updates rules on a filter.
func (c *ForemanClient) UpdateContentViewFilterRules(ctx context.Context, filterID int, rules []ForemanKatelloContentViewFilterRule) ([]ForemanKatelloContentViewFilterRule, error) {
	var updated []ForemanKatelloContentViewFilterRule
	for _, rule := range rules {
		var resp ForemanKatelloContentViewFilterRule
		err := c.Put(ctx, fmt.Sprintf("/katello/api/content_view_filters/%d/rules/%d", filterID, rule.ID), "content_view_filter_rule", &rule, &resp)
		if err != nil {
			return nil, err
		}
		updated = append(updated, resp)
	}
	return updated, nil
}

// SyncContentViewFilters syncs filters for a content view: creates new, updates existing, deletes removed.
func (c *ForemanClient) SyncContentViewFilters(ctx context.Context, cvID int, desired []ForemanKatelloContentViewFilter) error {
	existing, err := c.ReadContentViewFilters(ctx, cvID)
	if err != nil {
		// If read fails (e.g., 404), treat as no existing filters
		existing = nil
	}

	existingByID := make(map[int]ForemanKatelloContentViewFilter)
	for _, f := range existing {
		existingByID[f.ID] = f
	}
	desiredByID := make(map[int]ForemanKatelloContentViewFilter)
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
