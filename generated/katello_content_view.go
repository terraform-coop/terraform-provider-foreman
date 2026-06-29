package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
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
