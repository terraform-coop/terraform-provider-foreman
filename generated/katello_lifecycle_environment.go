package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ForemanKatelloLifecycleEnvironmentRequest struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Label          string `json:"label,omitempty"`
	OrganizationID int    `json:"organization_id,omitempty"`
	PriorID        int    `json:"prior_id,omitempty"`
}

type ForemanKatelloLifecycleEnvironment struct {
	ForemanObject
	Description    string `json:"description"`
	Label          string `json:"label"`
	OrganizationID int    `json:"organization_id"`
	Library        bool   `json:"library"`
	Prior          struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"prior"`
	Successor struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"successor"`
}

func (c *ForemanClient) CreateForemanKatelloLifecycleEnvironment(ctx context.Context, req *ForemanKatelloLifecycleEnvironmentRequest) (*ForemanKatelloLifecycleEnvironment, error) {
	var resp ForemanKatelloLifecycleEnvironment
	err := c.Post(ctx, "/katello/api/environments", "environment", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloLifecycleEnvironment(ctx context.Context, id int) (*ForemanKatelloLifecycleEnvironment, error) {
	var resp ForemanKatelloLifecycleEnvironment
	err := c.Get(ctx, fmt.Sprintf("/katello/api/environments/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloLifecycleEnvironment(ctx context.Context, id int, req *ForemanKatelloLifecycleEnvironmentRequest) (*ForemanKatelloLifecycleEnvironment, error) {
	var resp ForemanKatelloLifecycleEnvironment
	err := c.Put(ctx, fmt.Sprintf("/katello/api/environments/%d", id), "environment", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloLifecycleEnvironment(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/environments/%d", id))
}

func (c *ForemanClient) QueryForemanKatelloLifecycleEnvironment(ctx context.Context, name string) (*ForemanKatelloLifecycleEnvironment, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/environments?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloLifecycleEnvironment
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
