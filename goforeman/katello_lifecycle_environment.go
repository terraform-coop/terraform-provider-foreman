package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type KatelloLifecycleEnvironmentRequest struct {
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	Label          string `json:"label,omitempty"`
	OrganizationID int    `json:"organization_id,omitempty"`
	PriorID        int    `json:"prior_id,omitempty"`
}

type KatelloLifecycleEnvironment struct {
	Base
	Description    string `json:"description"`
	Label          string `json:"label"`
	OrganizationID int    `json:"organization_id"`
	Library        bool   `json:"library"`
	// PriorID/SuccessorID absorb a Katello read/write asymmetry: the
	// request accepts a flat "prior_id", but responses return nested
	// {"prior": {"id": ..., "name": ...}} / {"successor": ...} objects.
	// UnmarshalJSON flattens those back to plain IDs so callers see the
	// same shape they write.
	PriorID     int `json:"-"`
	SuccessorID int `json:"-"`
}

func (e *KatelloLifecycleEnvironment) UnmarshalJSON(b []byte) error {
	type alias KatelloLifecycleEnvironment
	aux := &struct {
		*alias
		Prior struct {
			ID int `json:"id"`
		} `json:"prior"`
		Successor struct {
			ID int `json:"id"`
		} `json:"successor"`
	}{alias: (*alias)(e)}
	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}
	e.PriorID = aux.Prior.ID
	e.SuccessorID = aux.Successor.ID
	return nil
}

func (c *Client) CreateKatelloLifecycleEnvironment(ctx context.Context, req *KatelloLifecycleEnvironmentRequest) (*KatelloLifecycleEnvironment, error) {
	var resp KatelloLifecycleEnvironment
	err := c.Post(ctx, "/katello/api/environments", "environment", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ReadKatelloLifecycleEnvironment(ctx context.Context, id int) (*KatelloLifecycleEnvironment, error) {
	var resp KatelloLifecycleEnvironment
	err := c.Get(ctx, fmt.Sprintf("/katello/api/environments/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateKatelloLifecycleEnvironment(ctx context.Context, id int, req *KatelloLifecycleEnvironmentRequest) (*KatelloLifecycleEnvironment, error) {
	var resp KatelloLifecycleEnvironment
	err := c.Put(ctx, fmt.Sprintf("/katello/api/environments/%d", id), "environment", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteKatelloLifecycleEnvironment(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("/katello/api/environments/%d", id))
}

func (c *Client) FindKatelloLifecycleEnvironmentByName(ctx context.Context, name string) (*KatelloLifecycleEnvironment, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("/katello/api/environments?search=name=\"%s\"", url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj KatelloLifecycleEnvironment
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
