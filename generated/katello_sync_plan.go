package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ForemanKatelloSyncPlanRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Interval    string `json:"interval,omitempty"`
	SyncDate    string `json:"sync_date,omitempty"`
	// Enabled is *bool with omitempty (not a plain bool+omitempty, and not
	// *bool without omitempty): a plain bool+omitempty can never send an
	// explicit "false" (Go's zero value for bool is false), but a nil *bool
	// WITHOUT omitempty sends an explicit JSON null when the field is
	// genuinely unset - Foreman's "enabled" column is NOT NULL like almost
	// every boolean column, so that null gets rejected outright. Go's
	// encoding/json only treats a nil pointer as "empty" for omitempty
	// purposes, so *bool+omitempty gets both right: nil omits the key
	// entirely, non-nil (even pointing at false) always sends it.
	Enabled        *bool  `json:"enabled,omitempty"`
	CronExpression string `json:"cron_expression,omitempty"`
}

type ForemanKatelloSyncPlan struct {
	ForemanObject
	Description    string `json:"description"`
	Interval       string `json:"interval"`
	SyncDate       string `json:"sync_date"`
	Enabled        bool   `json:"enabled"`
	CronExpression string `json:"cron_expression"`
}

func (c *ForemanClient) CreateForemanKatelloSyncPlan(ctx context.Context, req *ForemanKatelloSyncPlanRequest) (*ForemanKatelloSyncPlan, error) {
	var resp ForemanKatelloSyncPlan
	err := c.Post(ctx, fmt.Sprintf("katello/organizations/%d/sync_plans", c.config.OrganizationID), "sync_plan", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) ReadForemanKatelloSyncPlan(ctx context.Context, id int) (*ForemanKatelloSyncPlan, error) {
	var resp ForemanKatelloSyncPlan
	err := c.Get(ctx, fmt.Sprintf("katello/organizations/%d/sync_plans/%d", c.config.OrganizationID, id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) UpdateForemanKatelloSyncPlan(ctx context.Context, id int, req *ForemanKatelloSyncPlanRequest) (*ForemanKatelloSyncPlan, error) {
	var resp ForemanKatelloSyncPlan
	err := c.Put(ctx, fmt.Sprintf("katello/organizations/%d/sync_plans/%d", c.config.OrganizationID, id), "sync_plan", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ForemanClient) DeleteForemanKatelloSyncPlan(ctx context.Context, id int) error {
	return c.Delete(ctx, fmt.Sprintf("katello/organizations/%d/sync_plans/%d", c.config.OrganizationID, id))
}

func (c *ForemanClient) QueryForemanKatelloSyncPlan(ctx context.Context, name string) (*ForemanKatelloSyncPlan, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("katello/organizations/%d/sync_plans?search=name=\"%s\"", c.config.OrganizationID, url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj ForemanKatelloSyncPlan
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
